package loan_dto_handler

import (
	"github.com/go-playground/validator/v10"
)

// CreateLoanRequest represents the payload for creating a loan
type CreateLoanRequest struct {
	CustomerID uint    `json:"customer_id" binding:"required"`
	Name       string  `json:"name" binding:"required,alpha_space"`
	Email      string  `json:"email" binding:"required,email"`
	Amount     float64 `json:"amount" binding:"required,money"`
	TermWeeks  int     `json:"term_weeks" binding:"required,min=1"`
	Rates      float64 `json:"rates" binding:"required,percentage"`
}

// Custom error messages for validation
func (r *CreateLoanRequest) CustomValidationMessages(err error) map[string]string {
	validationErrors := err.(validator.ValidationErrors)
	errorMessages := make(map[string]string)

	for _, fieldError := range validationErrors {
		switch fieldError.Field() {
		case "CustomerID":
			errorMessages["customer_id"] = "customer ID is required."
		case "Name":
			errorMessages["name"] = "name is required and should contain only alphabets and spaces."
		case "Email":
			errorMessages["email"] = "email is required and should be in a valid email format."
		case "Amount":
			errorMessages["amount"] = "amount is required and should be in a valid money format."
		case "TermWeeks":
			errorMessages["term_weeks"] = "term weeks is required and should be a number greater than zero."
		case "Rates":
			errorMessages["rates"] = "rates is required and should be a valid percentage format (0-100)."
		}
	}
	return errorMessages
}
