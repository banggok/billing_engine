package handler

import (
	loan_dto_handler "billing_enginee/api/handler/dto/loan"
	"billing_enginee/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type LoanHandler struct {
	loanUsecase usecase.LoanUsecase
}

func NewLoanHandler(loanUsecase usecase.LoanUsecase) *LoanHandler {
	return &LoanHandler{
		loanUsecase: loanUsecase,
	}
}

func (h *LoanHandler) CreateLoan(c *gin.Context) {
	var request loan_dto_handler.CreateLoanRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		// Check if it's a validation error
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := request.CustomValidationMessages(validationErrors)
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
		} else {
			// General error if it's not a validation error
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// Create the loan via the usecase
	response, err := h.loanUsecase.CreateLoan(c, request.CustomerID, request.Name, request.Email, request.Amount, request.TermWeeks, request.Rates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"loan_id":            strconv.FormatUint(uint64(response.LoanID), 10),
		"total_amount":       response.TotalAmount,
		"outstanding_amount": response.OutstandingAmount,
		"week":               response.Week,
		"due_date":           response.DueDate.Format("2006-01-02"),
	})
}

func (h *LoanHandler) GetOutstanding(c *gin.Context) {
	loanIDParam := c.Param("loan_id")
	loanID, err := strconv.ParseUint(loanIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	// Get outstanding payments via usecase
	response, err := h.loanUsecase.GetOutstanding(c, uint(loanID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"loan_id":            strconv.FormatUint(uint64(response.LoanID), 10),
		"total_amount":       response.TotalAmount,
		"outstanding_amount": response.OutstandingAmount,
		"due_date":           response.DueDate.Format("2006-01-02"),
		"week":               response.WeeksOutstanding,
	})
}

func (h *LoanHandler) MakePayment(c *gin.Context) {
	loanIDParam := c.Param("loan_id")
	loanID, err := strconv.ParseUint(loanIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid loan ID"})
		return
	}

	amountStr := c.Query("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Call the use case to process the payment
	err = h.loanUsecase.MakePayment(c, uint(loanID), amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment successful"})
}
