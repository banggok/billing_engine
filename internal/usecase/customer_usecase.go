package usecase

import (
	"billing_enginee/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors" // Use the correct package for error wrapping
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerUsecase interface {
	IsDelinquent(c *gin.Context, customerID uint) (bool, error)
}

type customerUsecase struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerUsecase(customerRepo repository.CustomerRepository) CustomerUsecase {
	return &customerUsecase{
		customerRepo: customerRepo,
	}
}

func (u *customerUsecase) IsDelinquent(c *gin.Context, customerID uint) (bool, error) {
	customer, err := u.customerRepo.GetCustomerByID(c, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("customerID", customerID).Info("Customer not found")
			return false, nil
		}
		log.WithFields(log.Fields{
			"customerID": customerID,
			"error":      err,
		}).Error("Failed to retrieve customer for delinquency check")
		return false, errors.Wrap(err, "failed to retrieve customer for delinquency check")
	}

	return customer.IsDelinquent(), nil
}
