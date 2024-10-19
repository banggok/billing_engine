// pkg/validators.go

package pkg

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

// Initialize all custom validators
func InitValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validators globally and check for errors
		if err := v.RegisterValidation("money", validateMoney); err != nil {
			log.WithError(err).Error("Failed to register custom validator: money")
		}
		if err := v.RegisterValidation("percentage", validatePercentage); err != nil {
			log.WithError(err).Error("Failed to register custom validator: percentage")
		}
		if err := v.RegisterValidation("alpha_space", validateAlphaSpace); err != nil {
			log.WithError(err).Error("Failed to register custom validator: alpha_space")
		}
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
