package usecase

import (
	"billing_enginee/internal/repository"
	"log"
	"time"
)

type PaymentUsecase interface {
	RunDaily(time.Time) error
}

type paymentUsecase struct {
	paymentRepo repository.PaymentRepository
}

func NewPaymentUsecase(paymentRepo repository.PaymentRepository) PaymentUsecase {
	return &paymentUsecase{
		paymentRepo: paymentRepo,
	}
}

func (pu *paymentUsecase) RunDaily(time time.Time) error {
	log.Println("Scheduler started: Checking for payments due in a week...")

	// Get the date one week from now
	nextWeek := time.AddDate(0, 0, 7)

	// Fetch scheduled payments with a due date one week from now
	payments, err := pu.paymentRepo.GetScheduledPaymentsDueInAWeek(nextWeek)
	if err != nil {
		log.Printf("Error fetching payments: %v\n", err)
		return err
	}

	// Update each payment's status to outstanding
	for _, payment := range payments {
		payment.SetStatus("outstanding")
		if err := pu.paymentRepo.UpdatePaymentStatus(payment); err != nil {
			log.Printf("Error updating payment status for payment ID %d: %v\n", payment.GetID(), err)
			return err
		}
	}

	log.Printf("Scheduler completed: Processed %d payments.\n", len(payments))
	return nil
}
