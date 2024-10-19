package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"
	"time"

	"github.com/pkg/errors" // Use the correct errors package

	log "github.com/sirupsen/logrus"
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

func (r *paymentRepository) GetPaymentsDueBeforeDateWithStatus(tx *gorm.DB, nextWeek time.Time) ([]*entity.Payment, error) {
	var paymentModels []model.Payment

	if err := tx.Where("DATE(due_date) < ? AND status IN ?", nextWeek.Format("2006-01-02"), []string{"scheduled", "outstanding"}).
		Find(&paymentModels).Error; err != nil {
		log.WithError(err).Error("Failed to retrieve payments due before date")
		return nil, errors.Wrap(err, "failed to retrieve payments due before date")
	}

	payments := make([]*entity.Payment, len(paymentModels))
	for i, model := range paymentModels {
		entityConvert, err := entity.MakePayment(&model)
		if err != nil {
			return nil, err
		}
		payments[i] = entityConvert
	}

	return payments, nil
}

func (r *paymentRepository) UpdatePaymentStatus(tx *gorm.DB, payment *entity.Payment) error {
	if err := tx.Model(&model.Payment{}).Where("id = ?", payment.GetID()).Update("status", payment.Status()).Error; err != nil {
		log.WithFields(log.Fields{
			"paymentID": payment.GetID(),
			"status":    payment.Status(),
			"error":     err,
		}).Error("Failed to update payment status")
		return errors.Wrap(err, "failed to update payment status")
	}

	return nil
}

func (r *paymentRepository) GetNextPayment(tx *gorm.DB, loanID uint) (*entity.Payment, error) {
	var paymentModel model.Payment
	err := tx.Where("loan_id = ? AND status IN ?", loanID, []string{"scheduled", "outstanding"}).Order("week asc").First(&paymentModel).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("loanID", loanID).Info("No next payment found")
			return nil, nil
		}
		log.WithFields(log.Fields{
			"loanID": loanID,
			"error":  err,
		}).Error("Failed to retrieve next payment")
		return nil, errors.Wrap(err, "failed to retrieve next payment")
	}

	return entity.MakePayment(&paymentModel)
}

func (r *paymentRepository) SavePayments(tx *gorm.DB, payments []*entity.Payment) error {
	paymentModels := make([]model.Payment, len(payments))
	for i, payment := range payments {
		paymentModels[i] = *payment.ToModel()
	}

	if err := tx.Create(&paymentModels).Error; err != nil {
		log.WithError(err).Error("Failed to save payments")
		return errors.Wrap(err, "failed to save payments")
	}

	for i, paymentModel := range paymentModels {
		payments[i].SetID(paymentModel.ID)
	}

	return nil
}
