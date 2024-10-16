package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"

	"gorm.io/gorm"
)

type LoanRepository interface {
	SaveLoan(tx *gorm.DB, loan *entity.Loan) error
	GetLoanByID(tx *gorm.DB, loanID uint) (*entity.Loan, error)
	GetOutstandingPayments(tx *gorm.DB, loanID uint) (*entity.Loan, error)
	UpdateLoanStatus(tx *gorm.DB, loan *entity.Loan) error
}

type loanRepository struct {
}

func NewLoanRepository() LoanRepository {
	return &loanRepository{}
}

func (r *loanRepository) SaveLoan(tx *gorm.DB, loan *entity.Loan) error {
	// Convert entity.Loan to model.Loan
	loanModel := loan.ToModel()

	// Save loan to the database
	if err := tx.Create(&loanModel).Error; err != nil {
		return err
	}

	// Update entity with the generated ID
	loan.SetID(loanModel.ID)
	return nil
}

func (r *loanRepository) GetLoanByID(tx *gorm.DB, loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan
	if err := tx.First(&loanModel, loanID).Error; err != nil {
		return nil, err
	}

	// Convert model to entity
	loanEntity, err := entity.MakeLoan(&loanModel) // Now handling the error
	if err != nil {
		return nil, err
	}

	return loanEntity, nil
}

func (r *loanRepository) GetOutstandingPayments(tx *gorm.DB, loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan

	// Query loan and preload payments with status 'pending' or 'outstanding', ordered by week ascending
	if err := tx.Preload("Payments", func(db *gorm.DB) *gorm.DB {
		return db.Where("status IN ?", []string{"pending", "outstanding"}).Order("week ASC")
	}).First(&loanModel, loanID).Error; err != nil {
		return nil, err
	}

	// Convert loan model to loan entity, including payments
	loanEntity, err := entity.MakeLoan(&loanModel)
	if err != nil {
		return nil, err
	}

	return loanEntity, nil
}

func (r *loanRepository) UpdateLoanStatus(tx *gorm.DB, loan *entity.Loan) error {
	// Convert entity.Loan to model.Loan
	loanModel := loan.ToModel()

	// Update the status in the database using the loan's ID
	return tx.Model(&model.Loan{}).Where("id = ?", loanModel.ID).Update("status", loanModel.Status).Error
}
