package e2e_test

import (
	"billing_enginee/api/routes"
	"billing_enginee/internal/model"
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = ginkgo.Describe("Create Loan Endpoint", func() {
	var db *gorm.DB
	var router *gin.Engine
	var sqlDB *sql.DB

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Initialize the test database
		db, sqlDB, _ = pkg.InitTestDB() // Assume this initializes a test DB
		// Migrate the database schema for testing
		err := db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})
		Expect(err).ToNot(HaveOccurred())

		// Initialize repositories and use cases
		loanRepo := repository.NewLoanRepository(db)
		customerRepo := repository.NewCustomerRepository(db)
		paymentRepo := repository.NewPaymentRepository(db)
		loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)

		// Setup router without running the server
		router = gin.Default()
		routes.SetupLoanRoutes(router, loanUsecase)
	})

	// Tear down after each test
	ginkgo.AfterEach(func() {
		// Clean up the database by truncating tables
		db.Exec("TRUNCATE TABLE loans, customers, payments RESTART IDENTITY CASCADE;")
		sqlDB.Close()
	})

	ginkgo.It("should create a loan and verify it in the database", func() {
		// Create a loan request payload with updated values
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000, // Updated amount
			"term_weeks":  50,      // Updated term weeks
			"rates":       10,      // Updated rates
		}
		payloadJSON, _ := json.Marshal(payload)

		// Make an HTTP POST request to create a loan, but using the handler directly without starting the server
		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder() // Using an in-memory response recorder
		router.ServeHTTP(resp, req)    // Route the request to the in-memory Gin router

		// Verify that the response status code is 200 OK
		Expect(resp.Code).To(Equal(http.StatusOK))

		// Verify the response body contains the correct loan data
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())

		Expect(response["loan_id"]).NotTo(BeNil())
		Expect(response["total_amount"]).To(BeEquivalentTo(5500000.0))      // amount + rates (5000000 + 10% = 5500000)
		Expect(response["outstanding_amount"]).To(BeEquivalentTo(110000.0)) // total amount / 50 weeks (5500000 / 50 = 110000)

		// Verify that the customer was created in the database
		var customer model.Customer
		err = db.Where("id = ?", response["loan_id"]).First(&customer).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(customer.Name).To(Equal("John Doe"))
		Expect(customer.Email).To(Equal("johndoe@example.com"))

		// Verify that the loan was created in the database
		var loan model.Loan
		err = db.Where("id = ?", response["loan_id"]).First(&loan).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(loan.Amount).To(Equal(5000000.0))
		Expect(loan.TotalAmount).To(Equal(5500000.0))

		// Verify that payments were created
		var payments []model.Payment
		err = db.Where("loan_id = ?", loan.ID).Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(len(payments)).To(Equal(50)) // Expect 50 weekly payments

		// Check that the first payment is outstanding and the others are scheduled
		Expect(payments[0].Status).To(Equal("outstanding"))
		for i := 1; i < len(payments); i++ {
			Expect(payments[i].Status).To(Equal("scheduled"))
		}

		// Verify individual fields for payments
		Expect(payments[0].Amount).To(BeEquivalentTo(110000.0)) // Verify payment amount
		Expect(payments[0].Week).To(Equal(1))                   // Verify the week of the first payment
		Expect(payments[len(payments)-1].Week).To(Equal(50))    // Verify the last payment week

		// Calculate the expected due date for the first payment (1 week from now, zeroing out time)
		currentDate := time.Now().AddDate(0, 0, 7)
		expectedDueDate := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())

		// Verify that the due_date in the payment only matches the date part, ignoring time
		Expect(payments[0].DueDate.Year()).To(Equal(expectedDueDate.Year()))
		Expect(payments[0].DueDate.Month()).To(Equal(expectedDueDate.Month()))
		Expect(payments[0].DueDate.Day()).To(Equal(expectedDueDate.Day()))
	})

	// Test case: Validating required fields
	ginkgo.It("should return validation errors for missing required fields", func() {
		// Missing customer_id, name, email, amount, term_weeks, and rates
		payload := map[string]interface{}{}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Verify response status and error messages
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())

		// Check if the errors field contains the expected validation error messages
		errors := response["errors"].(map[string]interface{})

		Expect(errors["customer_id"]).To(ContainSubstring("customer ID is required"))
		Expect(errors["name"]).To(ContainSubstring("name is required"))
		Expect(errors["email"]).To(ContainSubstring("email is required"))
		Expect(errors["amount"]).To(ContainSubstring("amount is required"))
		Expect(errors["term_weeks"]).To(ContainSubstring("term weeks is required"))
		Expect(errors["rates"]).To(ContainSubstring("rates is required"))
	})

	// Test case: Validating incorrect formats
})
