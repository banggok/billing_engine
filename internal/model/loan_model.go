package model

import (
	"time"
)

type Loan struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	CustomerID  uint      `gorm:"not null"`
	Customer    Customer  `gorm:"foreignKey:CustomerID;references:ID"`
	Amount      float64   `gorm:"type:numeric(12,2);not null"`
	TotalAmount float64   `gorm:"type:numeric(12,2);not null"`
	Status      string    `gorm:"type:loan_status;default:'open'"` // Enum type mapped as a string
	TermWeeks   int       `gorm:"not null"`
	Rates       float64   `gorm:"type:numeric(5,2);not null"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	Payments *[]Payment `gorm:"foreignKey:LoanID"` // Foreign key relationship

}
