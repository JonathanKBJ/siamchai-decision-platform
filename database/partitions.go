package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// EnsureMonthlyPartitionsForRange creates monthly partition tables for 'sell' and 'sell_detail'
// for all months between startDate and endDate (inclusive).
func EnsureMonthlyPartitionsForRange(db *gorm.DB, startDate, endDate time.Time) error {
	zone := time.FixedZone("ICT", 7*3600)
	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, zone)
	targetEnd := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, zone)

	for !current.After(targetEnd) {
		year := current.Year()
		month := current.Month()

		startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, zone)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)

		partitionSuffix := fmt.Sprintf("y%04dm%02d", year, month)

		sellPartition := fmt.Sprintf("sell_%s", partitionSuffix)
		sellDetailPartition := fmt.Sprintf("sell_detail_%s", partitionSuffix)

		startStr := startOfMonth.Format("2006-01-02 15:04:05+07")
		endStr := endOfMonth.Format("2006-01-02 15:04:05+07")

		// Create sell partition
		sqlSell := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS %s PARTITION OF sell FOR VALUES FROM ('%s') TO ('%s');",
			sellPartition, startStr, endStr,
		)
		if err := db.Exec(sqlSell).Error; err != nil {
			log.Printf("Warning creating partition table %s: %v", sellPartition, err)
		}

		// Create sell_detail partition
		sqlDetail := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS %s PARTITION OF sell_detail FOR VALUES FROM ('%s') TO ('%s');",
			sellDetailPartition, startStr, endStr,
		)
		if err := db.Exec(sqlDetail).Error; err != nil {
			log.Printf("Warning creating partition table %s: %v", sellDetailPartition, err)
		}

		current = current.AddDate(0, 1, 0)
	}

	log.Printf("Verified monthly partition tables from %s to %s.", startDate.Format("2006-01"), endDate.Format("2006-01"))
	return nil
}

// EnsureMonthlyPartitions helper wrapper for current month up to N months ahead
func EnsureMonthlyPartitions(db *gorm.DB, monthsAhead int) error {
	now := time.Now()
	endDate := now.AddDate(0, monthsAhead, 0)
	return EnsureMonthlyPartitionsForRange(db, now, endDate)
}
