package model

import (
	"time"
)

type Payment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	LoanID    uint      `gorm:"not null"`
	Loan      Loan      `gorm:"foreignKey:LoanID;references:ID"` // Foreign key to Loan
	Week      int       `gorm:"not null"`
	Amount    float64   `gorm:"type:numeric(12,2);not null"`
	DueDate   time.Time `gorm:"type:date;not null"`
	Status    string    `gorm:"type:payment_status;default:'scheduled'"` // Enum for status
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
