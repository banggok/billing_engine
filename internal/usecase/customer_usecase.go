package usecase

import (
	"billing_enginee/internal/repository"

	"gorm.io/gorm"
)

type CustomerUsecase interface {
	IsDelinquent(tx *gorm.DB, customerID uint) (bool, error)
}

type customerUsecase struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerUsecase(customerRepo repository.CustomerRepository) CustomerUsecase {
	return &customerUsecase{
		customerRepo: customerRepo,
	}
}

func (u *customerUsecase) IsDelinquent(tx *gorm.DB, customerID uint) (bool, error) {
	// Fetch customer by ID
	customer, err := u.customerRepo.GetCustomerByID(tx, customerID)
	if err != nil {
		return false, err
	}

	// Use the entity method to check if customer is delinquent
	return customer.IsDelinquent(), nil
}
