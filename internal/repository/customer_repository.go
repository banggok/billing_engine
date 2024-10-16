package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"

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
	// Convert entity.Customer to model.Customer
	customerModel := customer.ToModel()

	// Save customer to the database
	if err := tx.Create(&customerModel).Error; err != nil {
		return err
	}

	// Update entity with the generated ID
	customer.SetID(customerModel.ID)
	return nil
}

func (r *customerRepository) GetCustomerByID(tx *gorm.DB, customerID uint) (*entity.Customer, error) {
	var customerModel model.Customer
	if err := tx.Preload("Loans.Payments").First(&customerModel, customerID).Error; err != nil {
		return nil, err
	}

	return entity.MakeCustomer(&customerModel)
}
