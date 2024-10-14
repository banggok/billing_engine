package usecase

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type LoanUsecase interface {
	CreateLoan(customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error)
	GetOutstanding(loanID uint) (*OutstandingResponse, error)
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
}

func NewLoanUsecase(loanRepo repository.LoanRepository, customerRepo repository.CustomerRepository) LoanUsecase {
	return &loanUsecase{
		loanRepo:     loanRepo,
		customerRepo: customerRepo,
	}
}

type LoanResponse struct {
	LoanID            uint
	TotalAmount       float64
	OutstandingAmount float64
	Week              int
	DueDate           time.Time
}

func (u *loanUsecase) CreateLoan(customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error) {
	// Check if the customer exists
	customer, err := u.customerRepo.GetCustomerByID(customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create and save customer if not found
			customer = entity.CreateCustomer(customerID, name, email)
			if err := u.customerRepo.SaveCustomer(customer); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Create loan entity
	loan := entity.CreateLoan(customer.GetID(), amount, termWeeks, rates)

	// Save loan to the repository
	if err := u.loanRepo.SaveLoan(loan); err != nil {
		return nil, err
	}

	// Calculate payment amount
	paymentAmount := loan.TotalAmount() / float64(termWeeks)

	// Generate payments
	payments := []*entity.Payment{}
	for week := 1; week <= termWeeks; week++ {
		status := "scheduled"
		if week == 1 {
			status = "outstanding"
		}

		dueDate := time.Now().AddDate(0, 0, 7*week)
		payments = append(payments, entity.CreatePayment(loan.GetID(), week, paymentAmount, dueDate, status))
	}

	// Save payments to the repository
	if err := u.loanRepo.SavePayments(payments); err != nil {
		return nil, err
	}

	// Prepare response
	response := &LoanResponse{
		LoanID:            loan.GetID(),
		TotalAmount:       loan.TotalAmount(),
		OutstandingAmount: paymentAmount,
		Week:              1,
		DueDate:           payments[0].DueDate(),
	}

	return response, nil
}

func (u *loanUsecase) GetOutstanding(loanID uint) (*OutstandingResponse, error) {
	// Fetch outstanding payments from the repository
	outstandingPayments, err := u.loanRepo.GetOutstandingPayments(loanID)
	if err != nil {
		return nil, err
	}

	if len(outstandingPayments) == 0 {
		return nil, nil // No outstanding payments
	}

	// Sum outstanding amounts and get latest due date
	var totalOutstanding float64
	var latestDueDate time.Time
	for _, payment := range outstandingPayments {
		totalOutstanding += payment.Amount()
		if payment.DueDate().After(latestDueDate) {
			latestDueDate = payment.DueDate()
		}
	}

	// Get the total amount from the loan
	response := &OutstandingResponse{
		LoanID:            loanID,
		TotalAmount:       outstandingPayments[0].Loan().TotalAmount(), // Corrected here
		OutstandingAmount: totalOutstanding,
		DueDate:           latestDueDate,
		WeeksOutstanding:  len(outstandingPayments),
	}

	return response, nil
}
