package entity

import (
	"billing_enginee/internal/model"
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
}

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

func MakeLoan(m *model.Loan) *Loan {
	return &Loan{
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
}

func (l *Loan) ToModel() *model.Loan {
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
	}
}

func (l *Loan) SetID(id uint) {
	l.id = id
}

func (l *Loan) GetID() uint {
	return l.id
}

func (l *Loan) TotalAmount() float64 {
	return l.totalAmount
}
