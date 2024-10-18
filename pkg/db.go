package pkg

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var err error

func loadEnv(envFile string) {
	if err := godotenv.Load(envFile); err != nil {
		log.WithFields(log.Fields{
			"envFile": envFile,
			"error":   err,
		}).Warn("Error loading environment file")
	}
}

func InitDB() (*gorm.DB, *sql.DB, error) {
	// Load environment variables if not already loaded
	loadEnv(".env")

	// Read database connection parameters from the environment
	dsn := createDSN()

	// Open the database connection with custom GORM logger
	DB, err = openDBConnection(dsn)
	if err != nil {
		log.WithError(err).Error("Failed to connect to main database")
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB connection from the gorm.DB
	sqlDB, err := DB.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get sql.DB from gorm.DB")
		return nil, nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	log.Info("Successfully connected to the main database")
	return DB, sqlDB, nil
}

func InitTestDB() (*gorm.DB, *sql.DB, error) {
	// Load the .env.test file for test environment
	loadEnv("../../.env.test")

	// Create DSN for the test database
	dsn := createDSN()

	// Open the database connection with custom GORM logger
	DB, err = openDBConnection(dsn)
	if err != nil {
		log.WithError(err).Error("Failed to connect to test database")
		return nil, nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Get the underlying sql.DB connection from the gorm.DB
	sqlDB, err := DB.DB()
	if err != nil {
		log.WithError(err).Error("Failed to get sql.DB from test gorm.DB")
		return nil, nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	log.Info("Successfully connected to the test database")
	return DB, sqlDB, nil
}

// createDSN generates the Data Source Name (DSN) based on environment variables
func createDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
	log.WithField("dsn", dsn).Debug("Generated DSN for database connection")
	return dsn
}

// openDBConnection sets up and returns a new GORM DB instance
func openDBConnection(dsn string) (*gorm.DB, error) {
	// Create a new GORM logger using logrus for better logging
	gormLogger := logger.New(
		log.StandardLogger(), // Use the standard logrus logger
		logger.Config{
			SlowThreshold:             time.Second, // Log slow queries
			LogLevel:                  logger.Warn, // Set log level to warn
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
			Colorful:                  true,        // Enable colorful logs
		},
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Prepare statements for better performance
		QueryFields: true, // Log all query fields
		Logger:      gormLogger,
	})
}
