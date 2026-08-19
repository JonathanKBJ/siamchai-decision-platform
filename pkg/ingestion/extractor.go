package ingestion

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Extractor struct {
	oraDB       *gorm.DB
	pgDB        *gorm.DB
	maxDays     int
	safetyDryRun bool
}

func NewExtractor(oraDB, pgDB *gorm.DB) *Extractor {
	daysStr := os.Getenv("DATA_EXTRACTION_MAX_DAYS")
	maxDays, err := strconv.Atoi(daysStr)
	if err != nil || maxDays <= 0 || maxDays > 7 {
		// Strict Security Guardrail: Force max 7 days test scope as instructed
		maxDays = 7
	}

	dryRun := os.Getenv("SAFETY_DRY_RUN") == "true"

	return &Extractor{
		oraDB:        oraDB,
		pgDB:         pgDB,
		maxDays:      maxDays,
		safetyDryRun: dryRun,
	}
}

// TestOracleConnectionQueries verifies query against Oracle DB strictly restricted to <= 7 days
func (e *Extractor) TestOracleConnectionQueries() error {
	log.Printf("==========================================================")
	log.Printf(" [SAFETY GUARDRAIL ENGAGED]")
	log.Printf(" Data query scope LIMITED to MAX %d DAYS (SYSDATE - %d)", e.maxDays, e.maxDays)
	log.Printf(" Safety Dry-Run Mode: %v (No full insertion into DB)", e.safetyDryRun)
	log.Printf("==========================================================")

	// 1. Test Query Master Shop Count
	var shopCount int64
	if err := e.oraDB.Table("shop").Where("active = 'Y'").Count(&shopCount).Error; err != nil {
		return fmt.Errorf("oracle shop count test failed: %w", err)
	}
	log.Printf("[Oracle ERP Test] Active Shops Count: %d", shopCount)

	// 2. Test Query Product Count
	var productCount int64
	if err := e.oraDB.Table("product").Where("status = 'ACTIVE'").Count(&productCount).Error; err != nil {
		return fmt.Errorf("oracle product count test failed: %w", err)
	}
	log.Printf("[Oracle ERP Test] Active Products Count: %d", productCount)

	// 3. Test Query Sales Transaction (STRICT 7 DAYS FILTER)
	startDate := time.Now().AddDate(0, 0, -e.maxDays).Format("2006-01-02")
	var salesCount int64
	querySales := fmt.Sprintf("sell_date >= DATE '%s'", startDate)

	if err := e.oraDB.Table("sell").Where(querySales).Count(&salesCount).Error; err != nil {
		return fmt.Errorf("oracle sales test (7 days limit) failed: %w", err)
	}
	log.Printf("[Oracle ERP Test] Sales Count in Last %d Days (%s to now): %d records", e.maxDays, startDate, salesCount)

	if e.safetyDryRun {
		log.Println("[SAFETY GUARDRAIL] Full data ingestion skipped to protect Core ERP Database as instructed.")
	}

	return nil
}
