package helpers

import (
	"billing_enginee/api/middleware"
	"billing_enginee/api/routes"
	"billing_enginee/internal/model"
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"database/sql"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// TestEnvironment holds the components needed for testing
type TestEnvironment struct {
	DB              *gorm.DB
	SQLDB           *sql.DB
	Router          *gin.Engine
	LoanRepo        repository.LoanRepository
	CustomerRepo    repository.CustomerRepository
	PaymentRepo     repository.PaymentRepository
	LoanUsecase     usecase.LoanUsecase
	PaymentUsecase  usecase.PaymentUsecase
	CustomerUsecase usecase.CustomerUsecase
}

// InitializeTestEnvironment sets up the common test environment, including DB, router, and validators
func InitializeTestEnvironment() *TestEnvironment {
	// Initialize the validators to be used globally
	pkg.InitValidators()

	// Initialize the test database
	db, sqlDB, _ := pkg.InitTestDB()
	// Migrate the database schema for testing
	err := db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})
	Expect(err).ToNot(HaveOccurred())

	// Initialize repositories
	loanRepo := repository.NewLoanRepository()
	customerRepo := repository.NewCustomerRepository()
	paymentRepo := repository.NewPaymentRepository()

	// Initialize use cases
	loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)

	// Setup router without running the server
	router := gin.Default()
	router.Use(middleware.TransactionMiddleware(db))
	routes.SetupLoanRoutes(router, loanUsecase)
	routes.SetupCustomerRoutes(router, customerUsecase)

	// Return a struct containing all components for flexible use in tests
	return &TestEnvironment{
		DB:              db,
		SQLDB:           sqlDB,
		Router:          router,
		LoanRepo:        loanRepo,
		CustomerRepo:    customerRepo,
		PaymentRepo:     paymentRepo,
		LoanUsecase:     loanUsecase,
		PaymentUsecase:  paymentUsecase,
		CustomerUsecase: customerUsecase,
	}
}
