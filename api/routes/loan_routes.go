package routes

import (
	"billing_enginee/api/handler"
	"billing_enginee/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupLoanRoutes(router *gin.Engine, loanUsecase usecase.LoanUsecase) {
	// Initialize the loan handler
	loanHandler := handler.NewLoanHandler(loanUsecase)

	// Define routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/loans", loanHandler.CreateLoan)
		v1.GET("/loans/:loan_id/outstanding", loanHandler.GetOutstanding)
	}
}
