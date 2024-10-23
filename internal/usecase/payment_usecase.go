package usecase

import (
	"billing_enginee/internal/repository"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentUsecase interface {
	UpdatePaymentStatus(tx *gorm.DB, tm time.Time) error
}

type paymentUsecase struct {
	paymentRepo repository.PaymentRepository
}

func NewPaymentUsecase(paymentRepo repository.PaymentRepository) PaymentUsecase {
	return &paymentUsecase{
		paymentRepo: paymentRepo,
	}
}

func (pu *paymentUsecase) UpdatePaymentStatus(tx *gorm.DB, currentDate time.Time) error {
	logrus.Info("Scheduler started: Checking for payments due in a week...")

	// Safely truncate the current date, retaining the timezone and avoiding shifting
	today := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
	nextWeek := currentDate.AddDate(0, 0, 7)
	nextWeek = time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 0, 0, 0, 0, nextWeek.Location())

	// Fetch all payments that are scheduled, outstanding, or pending
	payments, err := pu.paymentRepo.GetPaymentsDueBeforeDateWithStatus(nil, nextWeek)
	if err != nil {
		logrus.WithError(err).Error("Error fetching payments")
		return errors.New("error fetching payments: " + err.Error())
	}

	// Update the payment statuses
	for _, payment := range payments {
		if payment.DueDate().Before(today) {
			// Mark payments that are overdue as "pending"
			if err := payment.SetStatus("pending"); err != nil {
				logrus.WithFields(logrus.Fields{
					"paymentID": payment.GetID(),
					"status":    "pending",
					"error":     err,
				}).Error("Failed to set payment status to pending")
				return errors.New("failed to set payment status to pending: " + err.Error())
			}
		} else if payment.DueDate().Before(nextWeek) && payment.Status() == "scheduled" {
			// Mark payments due today as "outstanding"
			if err := payment.SetStatus("outstanding"); err != nil {
				logrus.WithFields(logrus.Fields{
					"paymentID": payment.GetID(),
					"status":    "outstanding",
					"error":     err,
				}).Error("Failed to set payment status to outstanding")
				return errors.New("failed to set payment status to outstanding: " + err.Error())
			}
		}

		if err := pu.paymentRepo.UpdatePaymentStatus(nil, payment); err != nil {
			logrus.WithFields(logrus.Fields{
				"paymentID": payment.GetID(),
				"error":     err,
			}).Error("Error updating payment status")
			return errors.New("failed to update payment status: " + err.Error())
		}
	}

	logrus.Infof("Scheduler completed: Processed %d payments.", len(payments))
	return nil
}
