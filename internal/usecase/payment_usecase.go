package usecase

import (
	"billing_enginee/internal/repository"
	"log"
	"time"

	"gorm.io/gorm"
)

type PaymentUsecase interface {
	RunDaily(tx *gorm.DB, tm time.Time) error
}

type paymentUsecase struct {
	paymentRepo repository.PaymentRepository
}

func NewPaymentUsecase(paymentRepo repository.PaymentRepository) PaymentUsecase {
	return &paymentUsecase{
		paymentRepo: paymentRepo,
	}
}

func (pu *paymentUsecase) RunDaily(tx *gorm.DB, currentDate time.Time) error {
	log.Println("Scheduler started: Checking for payments due in a week...")

	// Safely truncate the current date, retaining the timezone and avoiding shifting
	today := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
	nextWeek := currentDate.AddDate(0, 0, 7)
	nextWeek = time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 0, 0, 0, 0, nextWeek.Location())
	// Fetch all payments that are scheduled, outstanding, or pending
	payments, err := pu.paymentRepo.GetPaymentsDueBeforeDateWithStatus(tx, nextWeek)
	if err != nil {
		log.Printf("Error fetching payments: %v\n", err)
		return err
	}

	// Update the payment statuses
	for _, payment := range payments {
		if payment.DueDate().Before(today) {
			// Mark payments that are overdue as "pending"
			payment.SetStatus("pending")
		} else if payment.DueDate().Before(nextWeek) && payment.Status() == "scheduled" {
			// Mark payments due today as "outstanding"
			payment.SetStatus("outstanding")
		}

		if err := pu.paymentRepo.UpdatePaymentStatus(tx, payment); err != nil {
			log.Printf("Error updating payment status for payment ID %d: %v\n", payment.GetID(), err)
			return err
		}
	}

	log.Printf("Scheduler completed: Processed %d payments.\n", len(payments))
	return nil
}
