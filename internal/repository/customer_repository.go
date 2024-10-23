package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	SaveCustomer(c *gin.Context, customer *entity.Customer) error
	GetCustomerByID(c *gin.Context, customerID uint) (*entity.Customer, error)
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{
		db: db,
	}
}

func (r *customerRepository) SaveCustomer(c *gin.Context, customer *entity.Customer) error {
	customerModel := customer.ToModel()
	tx := GetDB(c, r.db)

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

func (r *customerRepository) GetCustomerByID(c *gin.Context, customerID uint) (*entity.Customer, error) {
	tx := GetDB(c, r.db)

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
