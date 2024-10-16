package loan_dto_handler

import (
	"fmt"
	"regexp"

	"log"

	"github.com/go-playground/validator/v10"
)

// CreateLoanRequest represents the payload for creating a loan
type CreateLoanRequest struct {
	CustomerID uint    `json:"customer_id" validate:"required"`      // Now required and must be greater than zero
	Name       string  `json:"name" validate:"required,alpha_space"` // Required and must allow alphabet and space
	Email      string  `json:"email" validate:"required,email"`      // Required and must be a valid email
	Amount     float64 `json:"amount" validate:"required,money"`     // Required and must match money format
	TermWeeks  int     `json:"term_weeks" validate:"required,gt=0"`  // Required and must be greater than zero
	Rates      float64 `json:"rates" validate:"required,percentage"` // Required and must match percentage format
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

// Register custom validators
func RegisterCustomValidators(validate *validator.Validate) {
	// Custom validator for money format (two decimal points)
	if err := validate.RegisterValidation("money", func(fl validator.FieldLevel) bool {
		amount := fl.Field().Float()
		moneyRegex := regexp.MustCompile(`^\d+(\.\d{1,2})?$`)
		return moneyRegex.MatchString(fmt.Sprintf("%.2f", amount))
	}); err != nil {
		log.Printf("Failed to register 'money' validation: %v", err)
	}

	// Custom validator for percentage format (0-100)
	if err := validate.RegisterValidation("percentage", func(fl validator.FieldLevel) bool {
		rate := fl.Field().Float()
		return rate >= 0 && rate <= 100
	}); err != nil {
		log.Printf("Failed to register 'percentage' validation: %v", err)
	}

	// Custom validator to allow only alphabetic characters and spaces
	if err := validate.RegisterValidation("alpha_space", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		alphaSpaceRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
		return alphaSpaceRegex.MatchString(name)
	}); err != nil {
		log.Printf("Failed to register 'alpha_space' validation: %v", err)
	}
}
