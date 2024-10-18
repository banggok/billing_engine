package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"errors"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	SaveCustomer(tx *gorm.DB, customer *entity.Customer) error
	GetCustomerByID(tx *gorm.DB, customerID uint) (*entity.Customer, error)
}

type customerRepository struct {
}

func NewCustomerRepository() CustomerRepository {
	return &customerRepository{}
}

func (r *customerRepository) SaveCustomer(tx *gorm.DB, customer *entity.Customer) error {
	customerModel := customer.ToModel()

	if err := tx.Create(&customerModel).Error; err != nil {
		log.WithFields(log.Fields{
			"customer": customerModel,
			"error":    err,
		}).Error("Failed to save customer")
		return errors.New("failed to save customer: " + err.Error())
	}

	customer.SetID(customerModel.ID)
	return nil
}

func (r *customerRepository) GetCustomerByID(tx *gorm.DB, customerID uint) (*entity.Customer, error) {
	var customerModel model.Customer
	if err := tx.Preload("Loans.Payments").First(&customerModel, customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("customerID", customerID).Info("Customer not found")
			return nil, gorm.ErrRecordNotFound
		}
		log.WithFields(log.Fields{
			"customerID": customerID,
			"error":      err,
		}).Error("Failed to retrieve customer")
		return nil, errors.New("failed to retrieve customer: " + err.Error())
	}

	return entity.MakeCustomer(&customerModel)
}
