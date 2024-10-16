package usecase

import (
	"billing_enginee/internal/entity"
	"billing_enginee/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type LoanUsecase interface {
	CreateLoan(tx *gorm.DB, customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error)
	GetOutstanding(tx *gorm.DB, loanID uint) (*OutstandingResponse, error)
	MakePayment(tx *gorm.DB, loanID uint, amount float64) error
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

func (u *loanUsecase) CreateLoan(tx *gorm.DB, customerID uint, name string, email string, amount float64, termWeeks int, rates float64) (*LoanResponse, error) {
	// Check if the customer exists
	customer, err := u.customerRepo.GetCustomerByID(tx, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create and save customer if not found
			customer = entity.CreateCustomer(customerID, name, email)
			if err := u.customerRepo.SaveCustomer(tx, customer); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Create loan entity
	loan := entity.CreateLoan(customer.GetID(), amount, termWeeks, rates)

	// Save loan to the repository
	if err := u.loanRepo.SaveLoan(tx, loan); err != nil {
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
	if err := u.paymentRepo.SavePayments(tx, payments); err != nil {
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

func (u *loanUsecase) GetOutstanding(tx *gorm.DB, loanID uint) (*OutstandingResponse, error) {
	// Fetch the loan with outstanding payments from the repository
	loan, err := u.loanRepo.GetOutstandingPayments(tx, loanID)
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

func (u *loanUsecase) MakePayment(tx *gorm.DB, loanID uint, amount float64) error {
	// Retrieve loan along with outstanding and pending payments by loan number
	loan, err := u.loanRepo.GetOutstandingPayments(tx, loanID)
	if err != nil {
		return err
	}

	payments := loan.GetPayments()

	// Define a small epsilon value for floating-point comparison
	// Validate the amount provided with tolerance for floating-point comparison
	if err := loan.ValidateAmount(amount); err != nil {
		return err
	}

	// Loop through payments and mark them as 'paid' until the amount runs out
	if err := u.updatePaid(tx, payments, amount); err != nil {
		return err
	}

	// Update the next scheduled payment to 'outstanding'
	// If no more payments are due, mark the loan as "closed"
	if err := u.updateNextPayment(tx, loan); err != nil {
		return err
	}

	return nil
}

func (u *loanUsecase) updateNextPayment(tx *gorm.DB, loan *entity.Loan) error {
	nextPayment, err := u.paymentRepo.GetNextPayment(tx, loan.GetID())
	if err == nil && nextPayment == nil {

		loan.SetStatus("close")
		if err := u.loanRepo.UpdateLoanStatus(tx, loan); err != nil {
			return err
		} else {
			return nil
		}
	}

	if err == nil && nextPayment.Status() == "scheduled" {
		nextPayment.SetStatus("outstanding")
		if err := u.paymentRepo.UpdatePaymentStatus(tx, nextPayment); err != nil {
			return err
		}
	}
	return nil
}

func (u *loanUsecase) updatePaid(tx *gorm.DB, payments *[]entity.Payment, amount float64) error {
	for _, payment := range *payments {
		if amount >= payment.Amount() {
			payment.SetStatus("paid")
			amount -= payment.Amount()
			if err := u.paymentRepo.UpdatePaymentStatus(tx, &payment); err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}
