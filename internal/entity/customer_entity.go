package entity

import "billing_enginee/internal/model"

type Customer struct {
	id    uint
	name  string
	email string
	loans *[]Loan
}

func CreateCustomer(id uint, name, email string) *Customer {
	return &Customer{
		id:    id,
		name:  name,
		email: email,
	}
}

func MakeCustomer(m *model.Customer) (*Customer, error) {
	c := &Customer{
		id:    m.ID,
		name:  m.Name,
		email: m.Email,
	}
	if m.Loans != nil && len(*m.Loans) > 0 {
		loans := make([]Loan, len(*m.Loans))
		for i, loanModel := range *m.Loans {
			loan, err := MakeLoan(&loanModel)
			if err != nil {
				return nil, err
			}
			loans[i] = *loan
		}
		c.loans = &loans
	}

	return c, nil
}

// Add ToModel method to convert entity.Customer to model.Customer
func (c *Customer) ToModel() *model.Customer {
	m := &model.Customer{
		ID:    c.id,
		Name:  c.name,
		Email: c.email,
	}

	if c.loans != nil && len(*c.loans) > 0 {
		loanModels := make([]model.Loan, len(*c.loans))
		for i, loanEntity := range *c.loans {
			loanModels[i] = *loanEntity.ToModel() // Assuming Loan has a ToModel method
		}
		m.Loans = &loanModels
	}

	return m
}

// SetID method to allow setting the ID for the Customer entity
func (c *Customer) SetID(id uint) {
	c.id = id
}

// GetID method to retrieve the ID of the Customer entity
func (c *Customer) GetID() uint {
	return c.id
}

func (c *Customer) IsDelinquent() bool {
	pendingCount := 0
	for _, loan := range *c.loans {
		for _, payment := range *loan.GetPayments() {
			if payment.Status() == "pending" {
				pendingCount++
			}

			// If 2 or more pending payments found, customer is delinquent
			if pendingCount >= 2 {
				return true
			}
		}
	}
	return false
}
