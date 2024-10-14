package e2e_test

import (
	"billing_enginee/api/routes"
	"billing_enginee/internal/model"
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = ginkgo.Describe("Create Loan Endpoint", func() {
	var db *gorm.DB
	var router *gin.Engine

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Initialize the test database
		db, _ = pkg.InitTestDB() // Assume this initializes a test DB
		// Migrate the database schema for testing
		db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})

		// Initialize repositories and use cases
		loanRepo := repository.NewLoanRepository(db)
		customerRepo := repository.NewCustomerRepository(db)
		loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo)

		// Setup router without running the server
		router = routes.SetupRouter(loanUsecase)
	})

	// Tear down after each test
	ginkgo.AfterEach(func() {
		// Clean up the database by truncating tables
		db.Exec("TRUNCATE TABLE loans, customers, payments RESTART IDENTITY CASCADE;")
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
		json.Unmarshal(resp.Body.Bytes(), &response)

		Expect(response["loan_id"]).NotTo(BeNil())
		Expect(response["total_amount"]).To(BeEquivalentTo(5500000.0))      // amount + rates (5000000 + 10% = 5500000)
		Expect(response["outstanding_amount"]).To(BeEquivalentTo(110000.0)) // total amount / 50 weeks (5500000 / 50 = 110000)

		// Verify that the customer was created in the database
		var customer model.Customer
		err := db.Where("id = ?", response["loan_id"]).First(&customer).Error
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
		expectedDueDate := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour) // Zero out time portion

		// Verify that the due_date in the payment only matches the date part, ignoring time
		Expect(payments[0].DueDate.Year()).To(Equal(expectedDueDate.Year()))
		Expect(payments[0].DueDate.Month()).To(Equal(expectedDueDate.Month()))
		Expect(payments[0].DueDate.Day()).To(Equal(expectedDueDate.Day()))
	})
})

var _ = ginkgo.Describe("Get Outstanding Endpoint", func() {
	var db *gorm.DB
	var router *gin.Engine
	var paymentRepo repository.PaymentRepository

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Initialize the test database
		db, _ = pkg.InitTestDB() // Assume this initializes a test DB
		// Migrate the database schema for testing
		db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})

		// Initialize repositories and use cases
		loanRepo := repository.NewLoanRepository(db)
		customerRepo := repository.NewCustomerRepository(db)
		paymentRepo = repository.NewPaymentRepository(db)
		loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo)

		// Setup router without running the server
		router = routes.SetupRouter(loanUsecase)
	})

	// Tear down after each test
	ginkgo.AfterEach(func() {
		// Clean up the database by truncating tables
		db.Exec("TRUNCATE TABLE loans, customers, payments RESTART IDENTITY CASCADE;")
	})
	ginkgo.It("should return the first payment as outstanding for a newly created loan", func() {
		// Step 1: Create a loan
		loanPayload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000, // Loan amount
			"term_weeks":  50,      // Term weeks
			"rates":       10,      // Loan rate (percentage)
		}
		loanPayloadJSON, _ := json.Marshal(loanPayload)

		// Create a loan using HTTP POST request
		loanReq, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(loanPayloadJSON))
		loanReq.Header.Set("Content-Type", "application/json")
		loanResp := httptest.NewRecorder()
		router.ServeHTTP(loanResp, loanReq)

		// Parse the loan creation response to extract loan ID, week, and due date
		var loanResponse map[string]interface{}
		json.Unmarshal(loanResp.Body.Bytes(), &loanResponse)

		// Convert loan_id to string, then to integer
		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Extract the outstanding payment details from the creation response
		expectedOutstandingAmount := loanResponse["outstanding_amount"].(float64)
		expectedDueDate := loanResponse["due_date"].(string)
		expectedWeek := loanResponse["week"].(float64) // Should be 1 for first outstanding payment

		// Step 2: Call the GetOutstanding endpoint
		outstandingReq, _ := http.NewRequest("GET", "/api/v1/loans/"+strconv.Itoa(loanID)+"/outstanding", nil)
		outstandingResp := httptest.NewRecorder()
		router.ServeHTTP(outstandingResp, outstandingReq)

		// Verify the response status code
		Expect(outstandingResp.Code).To(Equal(http.StatusOK))

		// Parse the GetOutstanding response
		var outstandingResponse map[string]interface{}
		json.Unmarshal(outstandingResp.Body.Bytes(), &outstandingResponse)

		// Step 3: Verify the outstanding amount, due date, and weeks outstanding
		Expect(outstandingResponse["loan_id"]).To(Equal(loanIDStr))                                     // Loan ID matches
		Expect(outstandingResponse["outstanding_amount"]).To(BeEquivalentTo(expectedOutstandingAmount)) // Outstanding amount matches
		Expect(outstandingResponse["due_date"]).To(Equal(expectedDueDate))                              // Due date should match the first payment's due date
		Expect(outstandingResponse["week"]).To(BeEquivalentTo(expectedWeek))                            // Week should be 1 (first payment outstanding)
	})

	ginkgo.It("should update outstanding payments via scheduler and retrieve them", func() {
		// Step 1: Create a loan
		loanPayload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000, // Loan amount
			"term_weeks":  50,      // Term weeks
			"rates":       10,      // Loan rate (percentage)
		}
		loanPayloadJSON, _ := json.Marshal(loanPayload)

		// Create a loan using HTTP POST request
		loanReq, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(loanPayloadJSON))
		loanReq.Header.Set("Content-Type", "application/json")
		loanResp := httptest.NewRecorder()
		router.ServeHTTP(loanResp, loanReq)

		// Parse the loan creation response to extract loan ID
		var loanResponse map[string]interface{}
		json.Unmarshal(loanResp.Body.Bytes(), &loanResponse)

		// Convert loan_id to string, then to integer
		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Step 2: Mock the time and trigger the scheduler to mark the first two payments as outstanding

		// Mock the server time to simulate the passage of time
		now := time.Now()
		firstPaymentDueDate := now.AddDate(0, 0, 7) // Second payment due in two weeks

		// Step forward in time to after the first payment's due date
		mockTime := firstPaymentDueDate.Add(24 * time.Hour)

		// Use RunDaily to update the payment statuses
		paymentUsecase := usecase.NewPaymentUsecase(paymentRepo) // Assuming paymentRepo has been initialized
		paymentUsecase.RunDaily(mockTime)                        // Mocked time for testing

		// Step 3: Call the GetOutstanding endpoint
		outstandingReq, _ := http.NewRequest("GET", "/api/v1/loans/"+strconv.Itoa(loanID)+"/outstanding", nil)
		outstandingResp := httptest.NewRecorder()
		router.ServeHTTP(outstandingResp, outstandingReq)

		// Verify the response status code
		Expect(outstandingResp.Code).To(Equal(http.StatusOK))

		// Parse the GetOutstanding response
		var outstandingResponse map[string]interface{}
		json.Unmarshal(outstandingResp.Body.Bytes(), &outstandingResponse)

		// Step 4: Verify the outstanding amount, due date, and latest week outstanding
		Expect(outstandingResponse["loan_id"]).To(Equal(loanIDStr))
		Expect(outstandingResponse["total_amount"]).To(Equal(5500000.0))               // Total amount of the loan
		Expect(outstandingResponse["outstanding_amount"]).To(BeEquivalentTo(220000.0)) // 2 outstanding payments of 110,000 each

		// Verify due_date is the due date of the second outstanding payment
		expectedDueDate := firstPaymentDueDate.AddDate(0, 0, 7).Format("2006-01-02")
		Expect(outstandingResponse["due_date"]).To(Equal(expectedDueDate))

		// Verify week is the latest outstanding week (2nd payment)
		latestWeekOutstanding := 2 // We expect the second week's payment to be the latest outstanding payment
		weekOutstanding := int(outstandingResponse["week"].(float64))
		Expect(weekOutstanding).To(Equal(latestWeekOutstanding)) // Ensure it shows the latest outstanding week
	})

})
