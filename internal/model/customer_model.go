package model

import "time"

type Customer struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Name      string  `gorm:"type:varchar(100);not null"`
	Email     string  `gorm:"type:varchar(100);unique;not null"`
	Loans     *[]Loan `gorm:"foreignKey:CustomerID;constraint:OnDelete:CASCADE"` // Add the Loans field
	CreatedAt time.Time
	UpdatedAt time.Time
}
