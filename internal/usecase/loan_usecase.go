package usecase

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors" // Import for error wrapping
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LoanUsecase interface {
	CreateLoan(c *gin.Context, customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error)
	GetOutstanding(c *gin.Context, loanID uint) (*OutstandingResponse, error)
	MakePayment(c *gin.Context, loanID uint, amount float64) error
}

type OutstandingResponse struct {
	LoanID            uint
	TotalAmount       float64
	OutstandingAmount float64
	DueDate           time.Time
	WeeksOutstanding  int
}

type loanUsecase struct {
	loanRepo     repository.LoanRepository
	customerRepo repository.CustomerRepository
	paymentRepo  repository.PaymentRepository
}

func NewLoanUsecase(
	loanRepo repository.LoanRepository,
	customerRepo repository.CustomerRepository,
	paymentrepo repository.PaymentRepository,
) LoanUsecase {
	return &loanUsecase{
		loanRepo:     loanRepo,
		customerRepo: customerRepo,
		paymentRepo:  paymentrepo,
	}
}

type LoanResponse struct {
	LoanID            uint
	TotalAmount       float64
	OutstandingAmount float64
	Week              int
	DueDate           time.Time
}

func (u *loanUsecase) CreateLoan(c *gin.Context, customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error) {
	customer, err := u.customerRepo.GetCustomerByID(c, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			customer = entity.CreateCustomer(customerID, name, email)
			if saveErr := u.customerRepo.SaveCustomer(c, customer); saveErr != nil {
				log.WithFields(log.Fields{
					"customer": customer,
					"error":    saveErr,
				}).Error("Failed to save customer during loan creation")
				return nil, errors.Wrap(saveErr, "failed to save customer during loan creation")
			}
		} else {
			log.WithFields(log.Fields{
				"customerID": customerID,
				"error":      err,
			}).Error("Failed to retrieve customer during loan creation")
			return nil, errors.Wrap(err, "failed to retrieve customer during loan creation")
		}
	}

	loan := entity.CreateLoan(customer.GetID(), amount, termWeeks, rates)

	if err := u.loanRepo.SaveLoan(c, loan); err != nil {
		log.WithFields(log.Fields{
			"loan":  loan,
			"error": err,
		}).Error("Failed to save loan")
		return nil, errors.Wrap(err, "failed to save loan")
	}

	paymentAmount := loan.TotalAmount() / float64(termWeeks)
	payments := []*entity.Payment{}
	for week := 1; week <= termWeeks; week++ {
		status := "scheduled"
		if week == 1 {
			status = "outstanding"
		}
		dueDate := time.Now().AddDate(0, 0, 7*week)
		x, err := entity.CreatePayment(loan.GetID(), week, paymentAmount, dueDate, status)
		if err != nil {
			return nil, err
		}
		payments = append(payments, x)
	}

	if err := u.paymentRepo.SavePayments(c, payments); err != nil {
		log.WithFields(log.Fields{
			"payments": payments,
			"error":    err,
		}).Error("Failed to save payments")
		return nil, errors.Wrap(err, "failed to save payments")
	}

	response := &LoanResponse{
		LoanID:            loan.GetID(),
		TotalAmount:       loan.TotalAmount(),
		OutstandingAmount: paymentAmount,
		Week:              1,
		DueDate:           payments[0].DueDate(),
	}

	return response, nil
}

func (u *loanUsecase) GetOutstanding(c *gin.Context, loanID uint) (*OutstandingResponse, error) {
	loan, err := u.loanRepo.GetOutstandingPayments(c, loanID)
	if err != nil {
		log.WithFields(log.Fields{
			"loanID": loanID,
			"error":  err,
		}).Error("Failed to get outstanding payments for loan")
		return nil, errors.Wrap(err, "failed to get outstanding payments for loan")
	}

	payments := loan.GetPayments()

	if len(*payments) == 0 {
		log.WithField("loanID", loanID).Info("No outstanding payments found")
		return nil, nil
	}

	var pendingPayments []entity.Payment
	var outstandingPayment *entity.Payment
	var totalOutstanding float64
	var latestDueDate time.Time
	var latestWeek int

	for _, payment := range *payments {
		if payment.Status() == "pending" {
			pendingPayments = append(pendingPayments, payment)
		} else if payment.Status() == "outstanding" {
			outstandingPayment = &payment
		}
	}

	switch {
	case len(pendingPayments) == 0 && outstandingPayment != nil:
		totalOutstanding = outstandingPayment.Amount()
		latestDueDate = outstandingPayment.DueDate()
		latestWeek = outstandingPayment.Week()

	case len(pendingPayments) == 1 && outstandingPayment != nil:
		totalOutstanding = pendingPayments[0].Amount()
		latestDueDate = pendingPayments[0].DueDate()
		latestWeek = pendingPayments[0].Week()

	case len(pendingPayments) >= 2 && outstandingPayment != nil:
		for _, pending := range pendingPayments {
			totalOutstanding += pending.Amount()
		}
		totalOutstanding += outstandingPayment.Amount()
		latestDueDate = outstandingPayment.DueDate()
		latestWeek = outstandingPayment.Week()
	}

	response := &OutstandingResponse{
		LoanID:            loan.GetID(),
		TotalAmount:       loan.TotalAmount(),
		OutstandingAmount: totalOutstanding,
		DueDate:           latestDueDate,
		WeeksOutstanding:  latestWeek,
	}

	return response, nil
}

func (u *loanUsecase) MakePayment(c *gin.Context, loanID uint, amount float64) error {
	loan, err := u.loanRepo.GetOutstandingPayments(c, loanID)
	if err != nil {
		log.WithFields(log.Fields{
			"loanID": loanID,
			"error":  err,
		}).Error("Failed to retrieve loan for making payment")
		return errors.Wrap(err, "failed to retrieve loan for making payment")
	}

	payments := loan.GetPayments()

	if err := loan.ValidateAmount(amount); err != nil {
		log.WithFields(log.Fields{
			"loanID": loanID,
			"amount": amount,
			"error":  err,
		}).Error("Payment amount does not match outstanding balance")
		return errors.Wrap(err, "payment amount does not match outstanding balance")
	}

	if err := u.updatePaid(c, payments, amount); err != nil {
		return errors.Wrap(err, "failed to update payments to 'paid'")
	}

	if err := u.updateNextPayment(c, loan); err != nil {
		return errors.Wrap(err, "failed to update next payment or close loan")
	}

	return nil
}

func (u *loanUsecase) updateNextPayment(c *gin.Context, loan *entity.Loan) error {
	nextPayment, err := u.paymentRepo.GetNextPayment(c, loan.GetID())
	if err != nil {
		log.WithFields(log.Fields{
			"loanID": loan.GetID(),
			"error":  err,
		}).Error("Failed to retrieve next payment for updating")
		return errors.Wrap(err, "failed to retrieve next payment")
	}

	if nextPayment == nil {
		if err := loan.SetStatus("close"); err != nil {
			log.WithFields(log.Fields{
				"loanID": loan.GetID(),
				"status": "close",
				"error":  err,
			}).Error("Failed to set loan status to closed")
			return errors.Wrap(err, "failed to set loan status to closed")
		}

		if err := u.loanRepo.UpdateLoanStatus(c, loan); err != nil {
			log.WithFields(log.Fields{
				"loanID": loan.GetID(),
				"error":  err,
			}).Error("Failed to update loan status to closed")
			return errors.Wrap(err, "failed to update loan status to closed")
		}
		return nil
	}

	if nextPayment.Status() == "scheduled" {
		if err := nextPayment.SetStatus("outstanding"); err != nil {
			log.WithFields(log.Fields{
				"paymentID": nextPayment.GetID(),
				"status":    "outstanding",
				"error":     err,
			}).Error("Failed to set payment status to outstanding")
			return errors.Wrap(err, "failed to set payment status to outstanding")
		}

		if err := u.paymentRepo.UpdatePaymentStatus(c, nextPayment); err != nil {
			log.WithFields(log.Fields{
				"paymentID": nextPayment.GetID(),
				"loanID":    loan.GetID(),
				"error":     err,
			}).Error("Failed to update next payment to outstanding")
			return errors.Wrap(err, "failed to update next payment to outstanding")
		}
	}
	return nil
}

func (u *loanUsecase) updatePaid(c *gin.Context, payments *[]entity.Payment, amount float64) error {
	for _, payment := range *payments {
		if amount >= payment.Amount() {
			if err := payment.SetStatus("paid"); err != nil {
				log.WithFields(log.Fields{
					"paymentID": payment.GetID(),
					"status":    "paid",
					"error":     err,
				}).Error("Failed to set payment status to paid")
				return errors.Wrap(err, "failed to set payment status to paid")
			}

			amount -= payment.Amount()
			if err := u.paymentRepo.UpdatePaymentStatus(c, &payment); err != nil {
				log.WithFields(log.Fields{
					"paymentID": payment.GetID(),
					"amount":    payment.Amount(),
					"error":     err,
				}).Error("Failed to update payment to paid")
				return errors.Wrap(err, "failed to update payment to paid")
			}
		} else {
			break
		}
	}
	return nil
}
