package loan_dto_handler

// CreateLoanRequest represents the payload for creating a loan
type CreateLoanRequest struct {
	CustomerID uint    `json:"customer_id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
	TermWeeks  int     `json:"term_weeks" binding:"required"`
	Rates      float64 `json:"rates" binding:"required"`
}
