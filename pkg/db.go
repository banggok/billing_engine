package pkg

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var err error

func InitDB() (*gorm.DB, *sql.DB, error) {
	// Read database connection parameters from the environment
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Create the DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	// Open the database connection with custom GORM logger
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Prepare statements for better performance
		QueryFields: true, // Log all query fields
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // Log to stdout
			logger.Config{
				SlowThreshold:             time.Second,   // Log slow queries
				LogLevel:                  logger.Silent, // Set log level (adjust as needed)
				IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound
				Colorful:                  true,          // Enable colorful logs
			},
		),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB connection from the gorm.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	return DB, sqlDB, nil
}

func InitTestDB() (*gorm.DB, error) {
	// Load the .env.test file
	err = godotenv.Load("../../.env.test")
	if err != nil {
		log.Fatalf("Error loading .env.test file from config folder: %v", err)
	}

	// Read environment variables for the test database
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Create DSN for the test DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	// Connect to the test database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
