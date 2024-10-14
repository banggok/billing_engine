package entity

import (
	"billing_enginee/internal/model"
	"time"
)

type Payment struct {
	id      uint
	loanID  uint
	loan    *Loan // Reference to the associated loan
	week    int
	amount  float64
	dueDate time.Time
	status  string
}

func CreatePayment(loanID uint, week int, amount float64, dueDate time.Time, status string) *Payment {
	return &Payment{
		loanID:  loanID,
		week:    week,
		amount:  amount,
		dueDate: dueDate,
		status:  status,
	}
}

func MakePayment(m *model.Payment) *Payment {
	return &Payment{
		id:      m.ID,
		loanID:  m.LoanID,
		week:    m.Week,
		amount:  m.Amount,
		dueDate: m.DueDate,
		status:  m.Status,
	}
}

func (p *Payment) ToModel() *model.Payment {
	return &model.Payment{
		ID:      p.id,
		LoanID:  p.loanID,
		Week:    p.week,
		Amount:  p.amount,
		DueDate: p.dueDate,
		Status:  p.status,
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
	return p.status
}

func (p *Payment) SetStatus(status string) {
	p.status = status
}

func (p *Payment) Week() int {
	return p.week
}
