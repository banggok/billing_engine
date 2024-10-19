// pkg/validators.go

package pkg

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Initialize all custom validators
func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validators globally
		v.RegisterValidation("money", validateMoney)
		v.RegisterValidation("percentage", validatePercentage)
		v.RegisterValidation("alpha_space", validateAlphaSpace)
		// Add any other custom validations you need
	}
}

// Custom validator for money format (two decimal points)
func validateMoney(fl validator.FieldLevel) bool {
	amount := fl.Field().Float()
	moneyRegex := regexp.MustCompile(`^\d+(\.\d{1,2})?$`)
	return moneyRegex.MatchString(fmt.Sprintf("%.2f", amount))
}

// Custom validator for percentage format (0-100)
func validatePercentage(fl validator.FieldLevel) bool {
	rate := fl.Field().Float()
	return rate >= 0 && rate <= 100
}

// Custom validator to allow only alphabetic characters and spaces
func validateAlphaSpace(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	alphaSpaceRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
	return alphaSpaceRegex.MatchString(name)
}
