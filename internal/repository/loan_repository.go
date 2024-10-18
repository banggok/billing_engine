package repository

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/model"

	"github.com/pkg/errors" // Use the correct errors package

	log "github.com/sirupsen/logrus"
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
	loanModel := loan.ToModel()

	if err := tx.Create(&loanModel).Error; err != nil {
		log.WithFields(log.Fields{
			"loan":  loanModel,
			"error": err,
		}).Error("Failed to save loan")
		return errors.Wrap(err, "failed to save loan")
	}

	loan.SetID(loanModel.ID)
	return nil
}

func (r *loanRepository) GetLoanByID(tx *gorm.DB, loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan
	if err := tx.First(&loanModel, loanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("loanID", loanID).Info("Loan not found")
			return nil, err
		}
		log.WithFields(log.Fields{
			"loanID": loanID,
			"error":  err,
		}).Error("Failed to retrieve loan")
		return nil, errors.Wrap(err, "failed to retrieve loan")
	}

	loanEntity, err := entity.MakeLoan(&loanModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert model to entity")
	}

	return loanEntity, nil
}

func (r *loanRepository) GetOutstandingPayments(tx *gorm.DB, loanID uint) (*entity.Loan, error) {
	var loanModel model.Loan

	if err := tx.Preload("Payments", func(db *gorm.DB) *gorm.DB {
		return db.Where("status IN ?", []string{"pending", "outstanding"}).Order("week ASC")
	}).First(&loanModel, loanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("loanID", loanID).Info("Outstanding payments not found")
			return nil, err
		}
		log.WithFields(log.Fields{
			"loanID": loanID,
			"error":  err,
		}).Error("Failed to retrieve outstanding payments")
		return nil, errors.Wrap(err, "failed to retrieve outstanding payments")
	}

	loanEntity, err := entity.MakeLoan(&loanModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert model to entity")
	}

	return loanEntity, nil
}

func (r *loanRepository) UpdateLoanStatus(tx *gorm.DB, loan *entity.Loan) error {
	loanModel := loan.ToModel()

	if err := tx.Model(&model.Loan{}).Where("id = ?", loanModel.ID).Update("status", loanModel.Status).Error; err != nil {
		log.WithFields(log.Fields{
			"loanID": loanModel.ID,
			"status": loanModel.Status,
			"error":  err,
		}).Error("Failed to update loan status")
		return errors.Wrap(err, "failed to update loan status")
	}

	return nil
}
