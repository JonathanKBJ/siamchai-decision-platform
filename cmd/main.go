package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"ns.co.th/siamchai-decision-platform/database"
	"ns.co.th/siamchai-decision-platform/pkg/ingestion"
	"ns.co.th/siamchai-decision-platform/pkg/models"
)

func main() {
	// Parse Command Line Flags
	modeFlag := flag.String("mode", "", "Execution mode: 'daily', 'repair', or 'full' (default: 'daily')")
	startFlag := flag.String("start", "", "Sales start date in YYYY-MM-DD format")
	endFlag := flag.String("end", "", "Sales end date in YYYY-MM-DD format")
	lookbackDaysFlag := flag.Int("lookback-days", 7, "Number of lookback days for daily mode (default: 7)")
	fullResetFlag := flag.Bool("full-reset", false, "Explicitly force drop and recreate all PostgreSQL tables")
	flag.Parse()

	log.Println("==========================================================")
	log.Println(" Starting Siamchai AI Decision Platform Ingestion Service")
	log.Println("==========================================================")

	// Load Environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: .env file not found, relying on environment variables")
	}

	// Determine Execution Mode
	mode := *modeFlag
	if mode == "" {
		mode = os.Getenv("MIGRATION_MODE")
	}
	if mode == "" {
		mode = "daily" // Default to Daily Incremental Sync
	}

	// Calculate Sales Date Range based on Mode
	now := time.Now()
	var startDate, endDate string

	switch mode {
	case "repair":
		startDate = *startFlag
		if startDate == "" {
			startDate = os.Getenv("MIGRATION_SALES_START_DATE")
		}
		if startDate == "" {
			startDate = now.AddDate(0, 0, -7).Format("2006-01-02")
		}

		endDate = *endFlag
		if endDate == "" {
			endDate = os.Getenv("MIGRATION_SALES_END_DATE")
		}
		if endDate == "" {
			endDate = now.Format("2006-01-02")
		}

	case "full":
		startDate = *startFlag
		if startDate == "" {
			startDate = os.Getenv("MIGRATION_SALES_START_DATE")
		}
		if startDate == "" {
			startDate = "2026-01-01"
		}

		endDate = *endFlag
		if endDate == "" {
			endDate = os.Getenv("MIGRATION_SALES_END_DATE")
		}
		if endDate == "" {
			endDate = now.Format("2006-01-02")
		}

	default: // "daily"
		mode = "daily"
		days := *lookbackDaysFlag
		if days <= 0 {
			days = 7
		}

		startDate = now.AddDate(0, 0, -days).Format("2006-01-02")
		endDate = now.Format("2006-01-02")
	}

	fullReset := *fullResetFlag
	if mode == "full" {
		fullReset = true
	}

	log.Printf("[Execution Context] Mode: '%s' | Range: %s to %s | Full Reset: %v", mode, startDate, endDate, fullReset)

	// 1. Connect to PostgreSQL (122.155.164.15:5434) and auto-create target database if not exists
	pgDB, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("PostgreSQL Setup Error: %v", err)
	}

	// 2. Execute Migration SQL DDL to create PostgreSQL Tables, Partitions & Indexes (IF NOT EXISTS)
	sqlBytes, err := os.ReadFile("migrations/0001_init_schema.sql")
	if err != nil {
		sqlBytes, err = os.ReadFile("../migrations/0001_init_schema.sql")
	}

	if err == nil {
		log.Println("Ensuring PostgreSQL Schema & Tables (IF NOT EXISTS)...")
		if err := pgDB.Exec(string(sqlBytes)).Error; err != nil {
			log.Printf("Warning during DDL migration execution: %v", err)
		} else {
			log.Println("PostgreSQL Schema & Tables verified!")
		}
	} else {
		log.Printf("Migration DDL file read error: %v", err)
	}

	// Auto-migrate missing columns and indexes on existing PostgreSQL tables
	log.Println("Auto-migrating PostgreSQL table schema columns...")
	if err := pgDB.AutoMigrate(
		&models.Branch{},
		&models.ProductBrand{},
		&models.ProductCategory{},
		&models.ProductGroup{},
		&models.ProductType{},
		&models.Product{},
		&models.Supplier{},
		&models.Customer{},
		&models.StockBalance{},
		&models.StockTarget{},
	); err != nil {
		log.Printf("AutoMigrate Warning: %v", err)
	}

	// 2.1 Verify Dynamic Monthly Partitions for current and upcoming 3 months
	log.Println("Ensuring Monthly Sales Partitions (Current + 3 Months ahead)...")
	if err := database.EnsureMonthlyPartitions(pgDB, 3); err != nil {
		log.Printf("Monthly Partitions Warning: %v", err)
	}

	// 3. Connect to Existing Core ERP Oracle DB (10.0.1.32)
	oraDB, err := database.ConnectOracle()
	if err != nil {
		log.Fatalf("Oracle ERP Connection Error: %v", err)
	}

	// 4. Run Connection & Query Verification
	ext := ingestion.NewExtractor(oraDB, pgDB)
	if err := ext.TestOracleConnectionQueries(); err != nil {
		log.Printf("Oracle Extraction Test Warning: %v", err)
	}

	// 5. Execute Data Migration Engine
	migrator := ingestion.NewDataMigrator(oraDB, pgDB)
	if err := migrator.RunMigration(context.Background(), mode, startDate, endDate, fullReset); err != nil {
		log.Fatalf("Data Migration Error: %v", err)
	}

	log.Println("==========================================================")
	log.Println(" Data Ingestion Service Process Completed Successfully!")
	log.Println(" PostgreSQL DB: 122.155.164.15:5434/siamchai_decision_db")
	log.Println(" Oracle ERP DB: 10.0.1.152 (Read-Only Safety Mode)")
	log.Println("==========================================================")
}
