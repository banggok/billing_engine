package entity

import (
	"billing_enginee/internal/entity/enum"
	"billing_enginee/internal/model"
	"errors"
	"math"
	"time"

	logrus "github.com/sirupsen/logrus"
)

type Loan struct {
	id          uint
	customerID  uint
	amount      float64
	totalAmount float64
	status      enum.LoanStatus
	termWeeks   int
	rates       float64
	createdAt   time.Time
	updatedAt   time.Time
	payments    *[]Payment // Pointer to a slice of associated payments
}

// CreateLoan is used to initialize a new Loan entity
func CreateLoan(customerID uint, amount float64, termWeeks int, rates float64) *Loan {
	totalAmount := amount + (amount * rates / 100)
	logrus.WithFields(logrus.Fields{
		"customerID":  customerID,
		"amount":      amount,
		"termWeeks":   termWeeks,
		"rates":       rates,
		"totalAmount": totalAmount,
	}).Info("Creating new loan")

	status, _ := enum.ParseLoanStatus("open")
	return &Loan{
		customerID:  customerID,
		amount:      amount,
		totalAmount: totalAmount,
		status:      status,
		termWeeks:   termWeeks,
		rates:       rates,
		createdAt:   time.Now(),
	}
}

// MakeLoan converts a model.Loan to an entity.Loan
func MakeLoan(m *model.Loan) (*Loan, error) {
	if m.Amount <= 0 || m.Rates < 0 || m.TermWeeks <= 0 {
		logrus.WithFields(logrus.Fields{
			"ID":         m.ID,
			"CustomerID": m.CustomerID,
			"Amount":     m.Amount,
			"Rates":      m.Rates,
			"TermWeeks":  m.TermWeeks,
			"Status":     m.Status,
		}).Error("Invalid loan data for MakeLoan")
		return nil, errors.New("invalid loan data: amount, rates, and termWeeks must be positive values")
	}

	status, err := enum.ParseLoanStatus(m.Status)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Status": m.Status,
			"Error":  err.Error(),
		}).Error("Failed to parse loan status during MakeLoan")
		return nil, err
	}

	loan := &Loan{
		id:          m.ID,
		customerID:  m.CustomerID,
		amount:      m.Amount,
		totalAmount: m.TotalAmount,
		status:      status,
		termWeeks:   m.TermWeeks,
		rates:       m.Rates,
		createdAt:   m.CreatedAt,
		updatedAt:   m.UpdatedAt,
	}

	if m.Payments != nil && len(*m.Payments) > 0 {
		logrus.WithFields(logrus.Fields{
			"loanID":   m.ID,
			"payments": len(*m.Payments),
		}).Info("Converting model payments to entity payments")

		loan.payments = &[]Payment{}
		for _, paymentModel := range *m.Payments {
			paymentConvert, err := MakePayment(&paymentModel)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"loanID":  m.ID,
					"payment": paymentModel.ID,
					"error":   err.Error(),
				}).Error("Failed to convert model payment to entity during MakeLoan")
				return nil, err
			}
			paymentEntity := paymentConvert
			*loan.payments = append(*loan.payments, *paymentEntity)
		}

		if err := loan.HasOneOutstandingPayment(); err != nil {
			logrus.WithField("loanID", m.ID).Error("Loan has more than one outstanding payment")
			return nil, err
		}
	}

	return loan, nil
}

// ToModel converts an entity.Loan to a model.Loan
func (l *Loan) ToModel() *model.Loan {
	paymentModels := []model.Payment{}
	if l.payments != nil {
		for _, payment := range *l.payments {
			paymentModels = append(paymentModels, *payment.ToModel())
		}
	}

	logrus.WithFields(logrus.Fields{
		"loanID":   l.id,
		"payments": len(paymentModels),
	}).Info("Converting loan entity to model")
	return &model.Loan{
		ID:          l.id,
		CustomerID:  l.customerID,
		Amount:      l.amount,
		TotalAmount: l.totalAmount,
		Status:      l.status.String(),
		TermWeeks:   l.termWeeks,
		Rates:       l.rates,
		CreatedAt:   l.createdAt,
		UpdatedAt:   l.updatedAt,
		Payments:    &paymentModels,
	}
}

// HasOneOutstandingPayment checks if the loan has only one outstanding payment
func (l *Loan) HasOneOutstandingPayment() error {
	outstandingCount := 0
	if l.payments != nil {
		for _, payment := range *l.payments {
			if payment.Status() == "outstanding" {
				outstandingCount++
			}
		}
	}

	if outstandingCount > 1 {
		logrus.WithField("loanID", l.id).Error("More than one outstanding payment found")
		return errors.New("more than one outstanding payment found, this is a bug in the system")
	}
	return nil
}

func (l *Loan) ValidateAmount(amount float64) error {
	if l.GetTotalOutstandingAmount() == nil {
		logrus.WithField("loanID", l.id).Error("No payments found for validation")
		return errors.New("payment cannot be empty")
	}
	const epsilon = 0.00001
	totalOA := l.GetTotalOutstandingAmount()
	if math.Abs(*totalOA-amount) > epsilon {
		logrus.WithFields(logrus.Fields{
			"expected": *totalOA,
			"provided": amount,
		}).Error("Payment amount does not match outstanding balance")
		return errors.New("payment amount does not match outstanding balance")
	}
	return nil
}

func (l *Loan) GetTotalOutstandingAmount() *float64 {
	if l.payments != nil {
		var pendingPayments []Payment
		var outstandingPayment *Payment
		var totalOutstanding float64

		for _, payment := range *l.payments {
			if payment.Status() == "pending" {
				pendingPayments = append(pendingPayments, payment)
			} else if payment.Status() == "outstanding" {
				outstandingPayment = &payment
			}
		}
		switch {
		case len(pendingPayments) == 0 && outstandingPayment != nil:
			totalOutstanding = outstandingPayment.Amount()

		case len(pendingPayments) == 1 && outstandingPayment != nil:
			totalOutstanding = pendingPayments[0].Amount()

		case len(pendingPayments) >= 2 && outstandingPayment != nil:
			for _, pending := range pendingPayments {
				totalOutstanding += pending.Amount()
			}
			totalOutstanding += outstandingPayment.Amount()
		}
		logrus.WithFields(logrus.Fields{
			"loanID":            l.id,
			"totalOutstanding":  totalOutstanding,
			"pendingPayments":   len(pendingPayments),
			"outstandingExists": outstandingPayment != nil,
		}).Info("Calculated total outstanding amount")
		return &totalOutstanding
	}
	return nil
}

// SetID sets the loan ID
func (l *Loan) SetID(id uint) {
	logrus.WithFields(logrus.Fields{
		"oldID": l.id,
		"newID": id,
	}).Info("Setting loan ID")
	l.id = id
}

// GetID returns the loan ID
func (l *Loan) GetID() uint {
	return l.id
}

// TotalAmount returns the total amount of the loan
func (l *Loan) TotalAmount() float64 {
	return l.totalAmount
}

// SetPayments sets the payments for the loan
func (l *Loan) SetPayments(payments *[]Payment) {
	logrus.WithFields(logrus.Fields{
		"loanID":   l.id,
		"payments": len(*payments),
	}).Info("Setting payments for loan")
	l.payments = payments
}

// GetPayments returns the payments associated with the loan
func (l *Loan) GetPayments() *[]Payment {
	return l.payments
}

// SetStatus sets the status of the loan
func (l *Loan) SetStatus(status string) error {
	logrus.WithFields(logrus.Fields{
		"loanID": l.id,
		"status": status,
	}).Info("Setting loan status")
	statusEnum, err := enum.ParseLoanStatus(status)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"status": status,
			"error":  err.Error(),
		}).Error("Failed to parse and set loan status")
		return err
	}
	l.status = statusEnum
	return nil
}

// GetStatus gets the status of the loan
func (l *Loan) GetStatus() string {
	return l.status.String()
}
