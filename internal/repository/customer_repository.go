package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"

	"gorm.io/gorm"
)

type CustomerRepository interface {
	SaveCustomer(customer *entity.Customer) error
	GetCustomerByID(customerID uint) (*entity.Customer, error)
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{
		db: db,
	}
}

func (r *customerRepository) SaveCustomer(customer *entity.Customer) error {
	// Convert entity.Customer to model.Customer
	customerModel := customer.ToModel()

	// Save customer to the database
	if err := r.db.Create(&customerModel).Error; err != nil {
		return err
	}

	// Update entity with the generated ID
	customer.SetID(customerModel.ID)
	return nil
}

func (r *customerRepository) GetCustomerByID(customerID uint) (*entity.Customer, error) {
	var customerModel model.Customer
	if err := r.db.First(&customerModel, customerID).Error; err != nil {
		return nil, err
	}

	// Convert model to entity
	customerEntity := entity.MakeCustomer(&customerModel)
	return customerEntity, nil
}
