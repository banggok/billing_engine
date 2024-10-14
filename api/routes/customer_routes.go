package routes

import (
	"billing_enginee/api/handler"
	"billing_enginee/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupCustomerRoutes(router *gin.Engine, customerUsecase usecase.CustomerUsecase) {
	// Initialize the handler
	customerHandler := handler.NewCustomerHandler(customerUsecase)

	// Define routes
	api := router.Group("/api/v1")
	{
		api.GET("/customers/:customer_id/is_delinquent", customerHandler.IsDelinquent)
	}
}
