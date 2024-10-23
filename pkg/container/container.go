// pkg/container/container.go
package container

import (
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Container struct {
	DB              *gorm.DB
	SQLDB           *sql.DB
	Router          *gin.Engine
	CustomerUsecase usecase.CustomerUsecase
	PaymentUsecase  usecase.PaymentUsecase
	LoanUsecase     usecase.LoanUsecase
}

func NewContainer() (*Container, error) {
	// Initialize DB
	db, sqlDb, err := pkg.InitDB()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DB: %w", err)
	}

	// Initialize router
	router := gin.Default()
	pkg.InitValidators()

	// Repositories and Usecases
	customerRepo := repository.NewCustomerRepository(db)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo)

	loanRepo := repository.NewLoanRepository(db)
	loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)

	return &Container{
		DB:              db,
		SQLDB:           sqlDb,
		Router:          router,
		CustomerUsecase: customerUsecase,
		PaymentUsecase:  paymentUsecase,
		LoanUsecase:     loanUsecase,
	}, nil
}
