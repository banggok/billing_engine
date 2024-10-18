package handler

import (
	"billing_enginee/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
		log.WithFields(log.Fields{
			"customerIDParam": customerIDParam,
			"error":           err,
		}).Error("Invalid customer ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	tx, err := h.getTransactionFromMiddleware(c)
	if err != nil {
		log.WithFields(log.Fields{
			"customerID": customerID,
			"error":      err,
		}).Error("Failed to retrieve transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		return
	}

	isDelinquent, err := h.customerUsecase.IsDelinquent(tx, uint(customerID))
	if err != nil {
		log.WithFields(log.Fields{
			"customerID": customerID,
			"error":      err,
		}).Error("Failed to check if customer is delinquent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check delinquency status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_delinquent": isDelinquent})
}

// Helper function to extract transaction from context
func (h *CustomerHandler) getTransactionFromMiddleware(c *gin.Context) (*gorm.DB, error) {
	tx, exists := c.Get("db_tx")
	if !exists {
		log.Error("Transaction not found in context")
		return nil, gorm.ErrInvalidTransaction
	}
	return tx.(*gorm.DB), nil
}
