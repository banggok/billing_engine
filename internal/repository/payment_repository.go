package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	GetPaymentsDueBeforeDateWithStatus(tx *gorm.DB, nextWeek time.Time) ([]*entity.Payment, error)
	UpdatePaymentStatus(tx *gorm.DB, payment *entity.Payment) error
	GetNextPayment(tx *gorm.DB, loanID uint) (*entity.Payment, error)
	SavePayments(tx *gorm.DB, payments []*entity.Payment) error
}

type paymentRepository struct {
}

func NewPaymentRepository() PaymentRepository {
	return &paymentRepository{}
}

// Fetch payments due before nextWeek (ignoring time) with status 'scheduled' or 'outstanding'
func (r *paymentRepository) GetPaymentsDueBeforeDateWithStatus(tx *gorm.DB, nextWeek time.Time) ([]*entity.Payment, error) {
	var paymentModels []model.Payment

	// Fetch payments where due_date < nextWeek (comparing dates only) and status is 'scheduled' or 'outstanding'
	if err := tx.Where("DATE(due_date) < ? AND status IN ?", nextWeek.Format("2006-01-02"), []string{"scheduled", "outstanding"}).
		Find(&paymentModels).Error; err != nil {
		return nil, err
	}

	// Convert models to entities
	payments := make([]*entity.Payment, len(paymentModels))
	for i, model := range paymentModels {
		payments[i] = entity.MakePayment(&model)
	}

	return payments, nil
}

// Update the status of the payment
func (r *paymentRepository) UpdatePaymentStatus(tx *gorm.DB, payment *entity.Payment) error {
	return tx.Model(&model.Payment{}).Where("id = ?", payment.GetID()).Update("status", payment.Status()).Error
}

func (r *paymentRepository) GetNextPayment(tx *gorm.DB, loanID uint) (*entity.Payment, error) {
	var paymentModel model.Payment
	err := tx.Where("loan_id = ? AND status IN ?", loanID, []string{"scheduled", "outstanding"}).Order("week asc").First(&paymentModel).Error

	// Handle case when no record is found
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No next payment found
		}
		return nil, err // Return the actual error if it's not ErrRecordNotFound
	}

	// Convert model to entity and return
	return entity.MakePayment(&paymentModel), nil
}

func (r *paymentRepository) SavePayments(tx *gorm.DB, payments []*entity.Payment) error {
	// Convert entity.Payment to model.Payment
	paymentModels := make([]model.Payment, len(payments))
	for i, payment := range payments {
		paymentModels[i] = *payment.ToModel()
	}

	// Save payments to the database
	if err := tx.Create(&paymentModels).Error; err != nil {
		return err
	}

	// Update entities with generated IDs
	for i, paymentModel := range paymentModels {
		payments[i].SetID(paymentModel.ID)
	}

	return nil
}
