// cmd/api/main.go
package main

import (
	"billing_enginee/api/middleware"
	"billing_enginee/api/routes"
	"billing_enginee/internal/runner"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"billing_enginee/pkg/container"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	pkg.LoadEnv(".env")

	// Set up logging
	pkg.SetupLogger()

	// Create dependency container
	c, err := container.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	defer closeResources(c.SQLDB)

	// Set up middleware
	setupMiddleware(c.Router, c.DB)

	// Set up HTTP routes
	setupRoutes(c.Router, c.CustomerUsecase, c.LoanUsecase)

	// Initialize and register scheduler tasks
	scheduler := startScheduler()
	registerSchedulerTasks(scheduler, c.PaymentUsecase, c.DB)

	// Start HTTP server
	srv := createHTTPServer(c.Router)
	startHTTPServer(srv)

	// Handle graceful shutdown
	gracefulShutdown(srv, scheduler)
}

// startScheduler initializes and starts the cron scheduler.
func startScheduler() *cron.Cron {
	c := cron.New(cron.WithLocation(time.FixedZone("Asia/Jakarta", 7*60*60))) // UTC+7
	c.Start()
	return c
}

// registerSchedulerTasks registers tasks to be run by the scheduler.
func registerSchedulerTasks(scheduler *cron.Cron, paymentUsecase usecase.PaymentUsecase, db *gorm.DB) {
	// Register tasks separately
	runner.RegisterUpdatePaymentStatusScheduler(scheduler, paymentUsecase, db)

	// Easily add more scheduled tasks by calling other functions here
}

// setupMiddleware applies global middleware to the router.
func setupMiddleware(router *gin.Engine, db *gorm.DB) {
	// Apply CORS, logging, and any other middleware
	router.Use(middleware.TransactionMiddleware(db))
	// Add more middleware as needed
}

// closeResources closes the SQL database connection gracefully.
func closeResources(sqlDB *sql.DB) {
	if err := sqlDB.Close(); err != nil {
		log.Errorf("Error closing DB connection: %v", err)
	}
	log.Info("Resources closed gracefully")
}

// setupRoutes registers the application routes with the router.
func setupRoutes(router *gin.Engine, customerUsecase usecase.CustomerUsecase, loanUsecase usecase.LoanUsecase) {
	routes.SetupCustomerRoutes(router, customerUsecase)
	routes.SetupLoanRoutes(router, loanUsecase)
	// Add more route setups as needed
}

// createHTTPServer creates and configures an HTTP server.
func createHTTPServer(router *gin.Engine) *http.Server {
	// Load the port from the environment variable, with a default if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port), // Set the address using the port
		Handler: router,
	}
}

// startHTTPServer starts the HTTP server in a separate goroutine.
func startHTTPServer(srv *http.Server) {
	go func() {
		log.Infof("Server running on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
}

// gracefulShutdown handles the graceful shutdown of the HTTP server upon receiving a termination signal.
func gracefulShutdown(srv *http.Server, scheduler *cron.Cron) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Stop scheduler
	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to gracefully shutdown: %v", err)
	}
	log.Info("Server exited gracefully")
}
