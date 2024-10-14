package entity

import "billing_enginee/internal/model"

type Customer struct {
	id    uint
	name  string
	email string
}

func CreateCustomer(id uint, name, email string) *Customer {
	return &Customer{
		id:    id,
		name:  name,
		email: email,
	}
}

func MakeCustomer(m *model.Customer) *Customer {
	return &Customer{
		id:    m.ID,
		name:  m.Name,
		email: m.Email,
	}
}

func (c *Customer) ToModel() *model.Customer {
	return &model.Customer{
		ID:    c.id,
		Name:  c.name,
		Email: c.email,
	}
}

func (c *Customer) SetID(id uint) {
	c.id = id
}

func (c *Customer) GetID() uint {
	return c.id
}
