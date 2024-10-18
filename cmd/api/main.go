package main

import (
	"billing_enginee/api/middleware"
	"billing_enginee/api/routes"
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Set up logrus logging format and level
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	// Initialize the Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Allow frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Read ReadHeaderTimeout from the .env file and convert to time.Duration
	readHeaderTimeoutStr := os.Getenv("READ_HEADER_TIMEOUT")
	readHeaderTimeout, err := strconv.Atoi(readHeaderTimeoutStr)
	if err != nil || readHeaderTimeout <= 0 {
		log.Warn("Invalid or missing READ_HEADER_TIMEOUT, defaulting to 10 seconds")
		readHeaderTimeout = 10 // default to 10 seconds if the env variable is invalid or missing
	}

	// Initialize the database
	db, sqlDB, err := pkg.InitDB()
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize the database")
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.WithError(err).Fatal("Failed to close database connection")
		}
		log.Info("Database connection closed gracefully")
	}()

	// Apply the transaction middleware globally
	router.Use(middleware.TransactionMiddleware(db))

	// Initialize repositories
	loanRepo := repository.NewLoanRepository()
	customerRepo := repository.NewCustomerRepository()
	paymentRepo := repository.NewPaymentRepository()

	// Initialize usecases
	loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)

	// Start the daily scheduler
	go startScheduler(db, paymentUsecase)

	// Setup routes
	routes.SetupCustomerRoutes(router, customerUsecase)
	routes.SetupLoanRoutes(router, loanUsecase)

	// Create the HTTP server
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: time.Duration(readHeaderTimeout) * time.Second, // Protect from Slowloris attack
	}

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server error")
		}
	}()
	log.Info("Server running on port 8080")

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	log.WithField("signal", sig).Info("Received shutdown signal, shutting down server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited gracefully")
}

// startScheduler runs the daily task at 00:00:00 UTC+7
func startScheduler(db *gorm.DB, paymentUsecase usecase.PaymentUsecase) {
	// Timezone UTC+7
	location, err := time.LoadLocation("Asia/Jakarta") // Set to Asia/Jakarta for UTC+7
	if err != nil {
		log.WithError(err).Fatal("Error loading location")
	}

	for {
		now := time.Now().In(location)
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location) // 00:00:00 today in UTC+7

		// If the current time has passed 00:00:00, schedule for the next day
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Run the scheduler now before the sleep
		log.Info("Running scheduler...")
		if err := paymentUsecase.RunDaily(db, time.Now()); err != nil {
			log.WithError(err).Error("Error running daily scheduler")
		}

		// Calculate duration until the next 00:00:00 UTC+7
		durationUntilNextRun := next.Sub(now)
		log.WithField("durationUntilNextRun", durationUntilNextRun).Info("Scheduler will next run")

		// Sleep until the next 00:00:00 UTC+7
		time.Sleep(durationUntilNextRun)
	}
}
