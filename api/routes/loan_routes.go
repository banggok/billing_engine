package routes

import (
	"billing_enginee/api/handler"
	"billing_enginee/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupRouter(loanUsecase usecase.LoanUsecase) *gin.Engine {
	router := gin.Default()

	// Initialize the loan handler
	loanHandler := handler.NewLoanHandler(loanUsecase)

	// Define routes
	v1 := router.Group("/api/v1")
	{
		// Create loan route
		v1.POST("/loans", loanHandler.CreateLoan)
		// Get outstanding payments route
		v1.GET("/loans/:loan_id/outstanding", loanHandler.GetOutstanding)
	}

	return router
}
