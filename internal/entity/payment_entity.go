package entity

import (
	"billing_enginee/internal/entity/enum"
	"billing_enginee/internal/model"
	"time"

	logrus "github.com/sirupsen/logrus"
)

type Payment struct {
	id      uint
	loanID  uint
	loan    *Loan // Reference to the associated loan
	week    int
	amount  float64
	dueDate time.Time
	status  enum.PaymentStatus
}

func CreatePayment(loanID uint, week int, amount float64, dueDate time.Time, status string) (*Payment, error) {
	statusEnum, err := enum.ParsePaymentStatus(status)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"loanID":  loanID,
			"week":    week,
			"amount":  amount,
			"dueDate": dueDate,
			"status":  status,
			"error":   err.Error(),
		}).Error("Failed to parse payment status during CreatePayment")
		return nil, err
	}

	return &Payment{
		loanID:  loanID,
		week:    week,
		amount:  amount,
		dueDate: dueDate,
		status:  statusEnum,
	}, nil
}

func MakePayment(m *model.Payment) (*Payment, error) {
	statusEnum, err := enum.ParsePaymentStatus(m.Status)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"ID":      m.ID,
			"LoanID":  m.LoanID,
			"Week":    m.Week,
			"Amount":  m.Amount,
			"DueDate": m.DueDate,
			"Status":  m.Status,
			"Error":   err.Error(),
		}).Error("Failed to parse payment status during MakePayment")
		return nil, err
	}

	return &Payment{
		id:      m.ID,
		loanID:  m.LoanID,
		week:    m.Week,
		amount:  m.Amount,
		dueDate: m.DueDate,
		status:  statusEnum,
	}, nil
}

func (p *Payment) ToModel() *model.Payment {
	return &model.Payment{
		ID:      p.id,
		LoanID:  p.loanID,
		Week:    p.week,
		Amount:  p.amount,
		DueDate: p.dueDate,
		Status:  p.status.String(),
	}
}

func (p *Payment) SetID(id uint) {
	p.id = id
}

func (p *Payment) GetID() uint {
	return p.id
}

// Getter for Amount
func (p *Payment) Amount() float64 {
	return p.amount
}

// Getter for DueDate
func (p *Payment) DueDate() time.Time {
	return p.dueDate
}

// Add a Loan() method to fetch the associated Loan entity
func (p *Payment) Loan() *Loan {
	return p.loan
}

func (p *Payment) SetLoan(loan *Loan) {
	p.loan = loan
}

func (p *Payment) Status() string {
	return p.status.String()
}

func (p *Payment) SetStatus(status string) error {
	enum, err := enum.ParsePaymentStatus(status)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"currentStatus": status,
			"error":         err.Error(),
		}).Error("Failed to parse and set payment status")
		return err
	}
	p.status = enum
	return nil
}

func (p *Payment) Week() int {
	return p.week
}
