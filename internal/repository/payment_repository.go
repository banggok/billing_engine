package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	GetScheduledPaymentsDueInAWeek(nextWeek time.Time) ([]*entity.Payment, error)
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

// Fetch payments due in a week with status 'scheduled'
func (r *paymentRepository) GetScheduledPaymentsDueInAWeek(nextWeek time.Time) ([]*entity.Payment, error) {
	var paymentModels []model.Payment

	if err := r.db.Where("due_date <= ? AND status = ?", nextWeek, "scheduled").Find(&paymentModels).Error; err != nil {
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
