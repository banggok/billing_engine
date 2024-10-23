// internal/scheduler/payment_task.go
package runner

import (
	"billing_enginee/internal/usecase"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RegisterUpdatePaymentStatusScheduler schedules a daily task at midnight UTC+7 to run the payment use case.
func RegisterUpdatePaymentStatusScheduler(scheduler *cron.Cron, paymentUsecase usecase.PaymentUsecase, db *gorm.DB) {
	_, err := scheduler.AddFunc("0 0 * * *", func() {
		log.Info("Running daily payment task...")
		if err := paymentUsecase.UpdatePaymentStatus(db, time.Now()); err != nil {
			log.WithError(err).Error("Error running daily payment task")
		}
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to schedule daily payment task")
	}
}
