package handler

import (
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	if err != nil || customerID == 0 {
		log.WithFields(log.Fields{
			"customerIDParam": customerIDParam,
			"error":           err,
		}).Error("Invalid customer ID format")
		c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Code:    "INVALID_INPUT",
			Message: "Customer ID must be a valid positive integer",
			TraceID: pkg.GenerateTraceID(),
		})
		return
	}

	isDelinquent, err := h.customerUsecase.IsDelinquent(c, uint(customerID))
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
