package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ConnOracle   *gorm.DB
	ConnPostgres *gorm.DB
)

// EnsurePostgresDatabaseExists checks if target database exists on 122.155.164.15:5434, creating it if not.
func EnsurePostgresDatabaseExists() error {
	host := os.Getenv("DB_POSTGRES_HOST")
	port := os.Getenv("DB_POSTGRES_PORT")
	user := os.Getenv("DB_POSTGRES_USER")
	pass := os.Getenv("DB_POSTGRES_PASSWORD")
	targetDB := os.Getenv("DB_POSTGRES_DB_NAME")

	// Connect to default 'postgres' database first
	adminDsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, pass)
	db, err := sql.Open("postgres", adminDsn)
	if err != nil {
		return fmt.Errorf("failed to connect to admin postgres: %w", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = db.QueryRow(query, targetDB).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		log.Printf("Database '%s' does not exist. Creating new database...", targetDB)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDB))
		if err != nil {
			return fmt.Errorf("failed to create database '%s': %w", targetDB, err)
		}
		log.Printf("Successfully created database '%s'!", targetDB)
	} else {
		log.Printf("Database '%s' already exists on target PostgreSQL server.", targetDB)
	}

	return nil
}

// ConnectPostgres initializes GORM connection to the target PostgreSQL database
func ConnectPostgres() (*gorm.DB, error) {
	if err := EnsurePostgresDatabaseExists(); err != nil {
		return nil, err
	}

	host := os.Getenv("DB_POSTGRES_HOST")
	port := os.Getenv("DB_POSTGRES_PORT")
	user := os.Getenv("DB_POSTGRES_USER")
	pass := os.Getenv("DB_POSTGRES_PASSWORD")
	dbName := os.Getenv("DB_POSTGRES_DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Bangkok",
		host, port, user, pass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL '%s': %w", dbName, err)
	}

	ConnPostgres = db
	log.Println("Connected to PostgreSQL DB successfully")
	return ConnPostgres, nil
}

// ConnectOracle initializes GORM connection to existing Core ERP Oracle DB
func ConnectOracle() (*gorm.DB, error) {
	url := os.Getenv("DB_ORACLE_URL")
	port := 1521
	srvName := os.Getenv("DB_ORACLE_SERVICE_NAME")
	user := os.Getenv("DB_ORACLE_USER")
	pass := os.Getenv("DB_ORACLE_PASSWORD")

	connStr := oracle.BuildUrl(url, port, srvName, user, pass, nil)
	db, err := gorm.Open(oracle.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Oracle DB: %w", err)
	}

	ConnOracle = db
	log.Println("Connected to Oracle ERP DB successfully")
	return ConnOracle, nil
}
