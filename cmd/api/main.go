package main

import (
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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

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
		log.Fatalf("Failed to initialize the database: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to close database connection")
		}
		log.Println("Database connection closed gracefully")
	}()

	// Initialize repositories
	loanRepo := repository.NewLoanRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Initialize usecases
	loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)

	// Start the daily scheduler
	go startScheduler(paymentUsecase)

	// Initialize Gin
	router := gin.Default()

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
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Server running on port 8080")

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	log.WithFields(log.Fields{
		"signal": sig,
	}).Println("Received shutdown signal, shutting down server...")

	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

// startScheduler runs the daily task at 00:00:00 UTC+7
func startScheduler(paymentUsecase usecase.PaymentUsecase) {
	// Timezone UTC+7
	location, err := time.LoadLocation("Asia/Jakarta") // Set to Asia/Jakarta for UTC+7
	if err != nil {
		log.Fatalf("Error loading location: %v", err)
	}

	for {
		now := time.Now().In(location)
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location) // 00:00:00 today in UTC+7

		// If the current time has passed 00:00:00, schedule for the next day
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Run the scheduler now before the sleep
		log.Println("Running scheduler...")
		if err := paymentUsecase.RunDaily(time.Now()); err != nil {
			log.Printf("Error running daily scheduler: %v\n", err)
		}

		// Calculate duration until the next 00:00:00 UTC+7
		durationUntilNextRun := next.Sub(now)
		log.Printf("Scheduler will next run in %v\n", durationUntilNextRun)

		// Sleep until the next 00:00:00 UTC+7
		time.Sleep(durationUntilNextRun)
	}
}
