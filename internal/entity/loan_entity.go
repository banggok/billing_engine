package entity

import (
	"billing_enginee/internal/model"
	"errors"
	"time"
)

type Loan struct {
	id          uint
	customerID  uint
	amount      float64
	totalAmount float64
	status      string
	termWeeks   int
	rates       float64
	createdAt   time.Time
	updatedAt   time.Time
	payments    *[]Payment // Pointer to a slice of associated payments
}

// CreateLoan is used to initialize a new Loan entity
func CreateLoan(customerID uint, amount float64, termWeeks int, rates float64) *Loan {
	totalAmount := amount + (amount * rates / 100)
	return &Loan{
		customerID:  customerID,
		amount:      amount,
		totalAmount: totalAmount,
		status:      "open",
		termWeeks:   termWeeks,
		rates:       rates,
		createdAt:   time.Now(),
	}
}

// MakeLoan converts a model.Loan to an entity.Loan
func MakeLoan(m *model.Loan) (*Loan, error) {
	if m.Amount <= 0 || m.Rates < 0 || m.TermWeeks <= 0 {
		return nil, errors.New("invalid loan data: amount, rates, and termWeeks must be positive values")
	}

	loan := &Loan{
		id:          m.ID,
		customerID:  m.CustomerID,
		amount:      m.Amount,
		totalAmount: m.TotalAmount,
		status:      m.Status,
		termWeeks:   m.TermWeeks,
		rates:       m.Rates,
		createdAt:   m.CreatedAt,
		updatedAt:   m.UpdatedAt,
	}

	// If payments are provided in the model, convert them to entity.Payments and attach to the loan
	if m.Payments != nil && len(*m.Payments) > 0 {
		loan.payments = &[]Payment{} // Initialize an empty slice of payments

		// Convert model.Payments to entity.Payments
		for _, paymentModel := range *m.Payments {
			paymentEntity := MakePayment(&paymentModel)
			*loan.payments = append(*loan.payments, *paymentEntity)
		}

		// Perform validation to ensure the loan has only one outstanding payment
		if err := loan.HasOneOutstandingPayment(); err != nil {
			return nil, err // Return an error if more than one outstanding payment is found
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

	return &model.Loan{
		ID:          l.id,
		CustomerID:  l.customerID,
		Amount:      l.amount,
		TotalAmount: l.totalAmount,
		Status:      l.status,
		TermWeeks:   l.termWeeks,
		Rates:       l.rates,
		CreatedAt:   l.createdAt,
		UpdatedAt:   l.updatedAt,
		Payments:    &paymentModels, // Attach converted payments
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
		return errors.New("more than one outstanding payment found, this is a bug in the system")
	}
	return nil
}

// SetID sets the loan ID
func (l *Loan) SetID(id uint) {
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
	l.payments = payments
}

// GetPayments returns the payments associated with the loan
func (l *Loan) GetPayments() *[]Payment {
	return l.payments
}

// SetStatus sets the status of the loan
func (l *Loan) SetStatus(status string) {
	l.status = status
}

// GetStatus gets the status of the loan
func (l *Loan) GetStatus() string {
	return l.status
}
