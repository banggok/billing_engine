package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"

	"gorm.io/gorm"
)

type LoanRepository interface {
	SaveLoan(loan *entity.Loan) error
	GetLoanByID(loanID uint) (*entity.Loan, error)
	SavePayments(payments []*entity.Payment) error
	GetOutstandingPayments(loanID uint) (*entity.Loan, error)
	UpdateLoanStatus(loan *entity.Loan) error
}

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{
		db: db,
	}
}

func (r *loanRepository) SaveLoan(loan *entity.Loan) error {
	// Convert entity.Loan to model.Loan
	loanModel := loan.ToModel()

	// Save loan to the database
	if err := r.db.Create(&loanModel).Error; err != nil {
		return err
	}

	// Update entity with the generated ID
	loan.SetID(loanModel.ID)
	return nil
}

func (r *loanRepository) GetLoanByID(loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan
	if err := r.db.First(&loanModel, loanID).Error; err != nil {
		return nil, err
	}

	// Convert model to entity
	loanEntity, err := entity.MakeLoan(&loanModel) // Now handling the error
	if err != nil {
		return nil, err
	}

	return loanEntity, nil
}

func (r *loanRepository) SavePayments(payments []*entity.Payment) error {
	// Convert entity.Payment to model.Payment
	paymentModels := make([]model.Payment, len(payments))
	for i, payment := range payments {
		paymentModels[i] = *payment.ToModel()
	}

	// Save payments to the database
	if err := r.db.Create(&paymentModels).Error; err != nil {
		return err
	}

	// Update entities with generated IDs
	for i, paymentModel := range paymentModels {
		payments[i].SetID(paymentModel.ID)
	}

	return nil
}

func (r *loanRepository) GetOutstandingPayments(loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan

	// Query loan and preload payments with status 'pending' or 'outstanding', ordered by week ascending
	if err := r.db.Preload("Payments", func(db *gorm.DB) *gorm.DB {
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

func (r *loanRepository) UpdateLoanStatus(loan *entity.Loan) error {
	// Convert entity.Loan to model.Loan
	loanModel := loan.ToModel()

	// Update the status in the database using the loan's ID
	return r.db.Model(&model.Loan{}).Where("id = ?", loanModel.ID).Update("status", loanModel.Status).Error
}
