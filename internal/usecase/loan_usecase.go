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
	// Fetch the loan with outstanding payments from the repository
	loan, err := u.loanRepo.GetOutstandingPayments(loanID)
	if err != nil {
		return nil, err
	}

	// Get payments from the loan
	payments := loan.GetPayments()

	if len(*payments) == 0 {
		return nil, nil // No outstanding payments
	}

	// Initialize variables for pending, outstanding, and total outstanding amounts
	var pendingPayments []entity.Payment
	var outstandingPayment *entity.Payment
	var totalOutstanding float64
	var latestDueDate time.Time
	var latestWeek int

	// Iterate through payments to classify pending and outstanding payments
	for _, payment := range *payments {
		if payment.Status() == "pending" {
			pendingPayments = append(pendingPayments, payment)
		} else if payment.Status() == "outstanding" {
			outstandingPayment = &payment
		}
	}

	// Logic based on the conditions
	switch {
	case len(pendingPayments) == 0 && outstandingPayment != nil:
		// Case 1: Only 1 outstanding payment
		totalOutstanding = outstandingPayment.Amount()
		latestDueDate = outstandingPayment.DueDate()
		latestWeek = outstandingPayment.Week()

	case len(pendingPayments) == 1 && outstandingPayment != nil:
		// Case 2: 1 pending payment and 1 outstanding payment
		totalOutstanding = pendingPayments[0].Amount()
		latestDueDate = pendingPayments[0].DueDate()
		latestWeek = pendingPayments[0].Week()

	case len(pendingPayments) >= 2 && outstandingPayment != nil:
		// Case 3: 2 or more pending payments and 1 outstanding payment
		for _, pending := range pendingPayments {
			totalOutstanding += pending.Amount()
		}
		totalOutstanding += outstandingPayment.Amount()
		latestDueDate = outstandingPayment.DueDate()
		latestWeek = outstandingPayment.Week()
	}

	// Prepare the response with total amount and outstanding details
	response := &OutstandingResponse{
		LoanID:            loan.GetID(),
		TotalAmount:       loan.TotalAmount(),
		OutstandingAmount: totalOutstanding,
		DueDate:           latestDueDate,
		WeeksOutstanding:  latestWeek, // Week of the latest outstanding payment
	}

	return response, nil
}
