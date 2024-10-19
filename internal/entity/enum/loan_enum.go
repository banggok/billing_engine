package enum

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type LoanStatus int

const (
	LoanStatusOpen LoanStatus = iota
	LoanStatusClosed
)

var loanStatusNames = []string{
	"open",
	"close",
}

// String method to convert LoanStatus to string
func (status LoanStatus) String() string {
	if int(status) < len(loanStatusNames) {
		return loanStatusNames[status]
	}
	return "unknown"
}

// ParseLoanStatus converts string to LoanStatus
func ParseLoanStatus(status string) (LoanStatus, error) {
	for i, name := range loanStatusNames {
		if name == status {
			return LoanStatus(i), nil
		}
	}
	log.WithField("status", status).Error("Failed to parse LoanStatus")
	return -1, fmt.Errorf("invalid loan status: %s", status)
}
