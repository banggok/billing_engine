package handler

import (
	"billing_enginee/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerUsecase usecase.CustomerUsecase
}

func NewCustomerHandler(customerUsecase usecase.CustomerUsecase) *CustomerHandler {
	return &CustomerHandler{
		customerUsecase: customerUsecase,
	}
}

func (h *CustomerHandler) IsDelinquent(c *gin.Context) {
	customerIDParam := c.Param("customer_id")
	customerID, err := strconv.ParseUint(customerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	// Call the use case to check for delinquency
	isDelinquent, err := h.customerUsecase.IsDelinquent(uint(customerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the result
	c.JSON(http.StatusOK, gin.H{"is_delinquent": isDelinquent})
}
