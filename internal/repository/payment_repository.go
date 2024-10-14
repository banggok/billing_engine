package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	GetPaymentsDueBeforeDateWithStatus(nextWeek time.Time) ([]*entity.Payment, error)
	UpdatePaymentStatus(payment *entity.Payment) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

// Fetch payments due before nextWeek (ignoring time) with status 'scheduled' or 'outstanding'
func (r *paymentRepository) GetPaymentsDueBeforeDateWithStatus(nextWeek time.Time) ([]*entity.Payment, error) {
	var paymentModels []model.Payment

	// Fetch payments where due_date < nextWeek (comparing dates only) and status is 'scheduled' or 'outstanding'
	if err := r.db.Where("DATE(due_date) < ? AND status IN ?", nextWeek.Format("2006-01-02"), []string{"scheduled", "outstanding"}).
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
func (r *paymentRepository) UpdatePaymentStatus(payment *entity.Payment) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", payment.GetID()).Update("status", payment.Status()).Error
}
