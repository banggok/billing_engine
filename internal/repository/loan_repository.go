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
	GetOutstandingPayments(loanID uint) ([]*entity.Payment, error)
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
	loanEntity := entity.MakeLoan(&loanModel)
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

func (r *loanRepository) GetOutstandingPayments(loanID uint) ([]*entity.Payment, error) {
	var paymentModels []model.Payment

	// Query outstanding payments and preload the associated loan
	if err := r.db.Where("loan_id = ? AND status = ?", loanID, "outstanding").Preload("Loan").Find(&paymentModels).Error; err != nil {
		return nil, err
	}

	// Convert to entities
	outstandingPayments := make([]*entity.Payment, len(paymentModels))
	for i, model := range paymentModels {
		payment := entity.MakePayment(&model)

		// Convert model to entity and set the loan
		loanModel := model.Loan
		loan := entity.MakeLoan(&loanModel)
		payment.SetLoan(loan)

		outstandingPayments[i] = payment
	}

	return outstandingPayments, nil
}
