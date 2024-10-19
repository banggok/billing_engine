package enum

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type PaymentStatus int

const (
	PaymentStatusScheduled PaymentStatus = iota
	PaymentStatusOutstanding
	PaymentStatusPaid
	PaymentStatusPending
)

var paymentStatusNames = []string{
	"scheduled",
	"outstanding",
	"paid",
	"pending",
}

// String method to convert PaymentStatus to string
func (status PaymentStatus) String() string {
	if int(status) < len(paymentStatusNames) {
		return paymentStatusNames[status]
	}
	return "unknown"
}

// ParsePaymentStatus converts string to PaymentStatus
func ParsePaymentStatus(status string) (PaymentStatus, error) {
	for i, name := range paymentStatusNames {
		if name == status {
			return PaymentStatus(i), nil
		}
	}
	log.WithField("status", status).Error("Failed to parse PaymentStatus")
	return -1, fmt.Errorf("invalid payment status: %s", status)
}
