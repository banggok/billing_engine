package e2e_test

import (
	"billing_enginee/api/middleware"
	"billing_enginee/api/routes"
	"billing_enginee/internal/model"
	"billing_enginee/internal/repository"
	"billing_enginee/internal/usecase"
	"billing_enginee/pkg"
	"billing_enginee/tests/helpers"
	"bytes"
	"database/sql"
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

var _ = ginkgo.Describe("Is Delinquent Endpoint", func() {
	var db *gorm.DB
	var sqlDB *sql.DB
	var router *gin.Engine
	var paymentUsecase usecase.PaymentUsecase

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Initialize the test database
		db, sqlDB, _ = pkg.InitTestDB()
		// Migrate the database schema for testing
		err := db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})
		Expect(err).ToNot(HaveOccurred())

		// Initialize repositories and use cases
		loanRepo := repository.NewLoanRepository()
		customerRepo := repository.NewCustomerRepository()
		paymentRepo := repository.NewPaymentRepository()
		loanUsecase := usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)
		paymentUsecase = usecase.NewPaymentUsecase(paymentRepo)
		customerUsecase := usecase.NewCustomerUsecase(customerRepo)

		// Setup router without running the server
		router = gin.Default()
		router.Use(middleware.TransactionMiddleware(db))

		routes.SetupCustomerRoutes(router, customerUsecase)
		routes.SetupLoanRoutes(router, loanUsecase)
	})

	// Tear down after each test
	ginkgo.AfterEach(func() {
		// Clean up the database by truncating tables
		// Use the helper function to truncate tables
		err := helpers.TruncateTables(db, "loans", "customers", "payments")
		Expect(err).ToNot(HaveOccurred(), "Failed to truncate tables before running tests")
		sqlDB.Close()
	})

	// Helper function to get customer_id from the database using loan_id
	getCustomerIDFromLoan := func(loanID float64) uint {
		var loan model.Loan
		err := db.Where("id = ?", loanID).Preload("Customer").First(&loan).Error
		Expect(err).ToNot(HaveOccurred())
		return loan.CustomerID
	}

	ginkgo.It("should return false when a customer has a newly created loan (no pending payments)", func() {
		// Step 1: Create a loan using the CreateLoan endpoint
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Parse the loan creation response
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Fetch customer_id from the loans table
		customerID := getCustomerIDFromLoan(float64(loanID))

		// Step 2: Call the IsDelinquent endpoint
		req, _ = http.NewRequest("GET", "/api/v1/customers/"+strconv.Itoa(int(customerID))+"/is_delinquent", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Verify the response status code and body
		Expect(resp.Code).To(Equal(http.StatusOK))
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())
		Expect(response["is_delinquent"]).To(BeFalse()) // Customer should not be delinquent
	})

	ginkgo.It("should return false when a customer has one outstanding payment (scheduler run once)", func() {
		// Step 1: Create a loan using the CreateLoan endpoint
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "Jane Doe",
			"email":       "janedoe@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Parse the loan creation response
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Fetch customer_id from the loans table
		customerID := getCustomerIDFromLoan(float64(loanID))

		// Step 2: Run the scheduler once
		currentDate := time.Now().AddDate(0, 0, 8) // Simulate 8 days later
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Step 3: Call the IsDelinquent endpoint
		req, _ = http.NewRequest("GET", "/api/v1/customers/"+strconv.Itoa(int(customerID))+"/is_delinquent", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Verify the response status code and body
		Expect(resp.Code).To(Equal(http.StatusOK))
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())

		Expect(response["is_delinquent"]).To(BeFalse()) // Customer should not be delinquent
	})

	ginkgo.It("should return true when a customer has two pending payments (scheduler run twice)", func() {
		// Step 1: Create a loan using the CreateLoan endpoint
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "Jane Doe",
			"email":       "janedoe@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Parse the loan creation response
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Fetch customer_id from the loans table
		customerID := getCustomerIDFromLoan(float64(loanID))

		// Step 2: Run the scheduler twice
		currentDate := time.Now().AddDate(0, 0, 8) // Simulate 8 days later
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		currentDate = time.Now().AddDate(0, 0, 15) // Simulate 15 days later
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Step 3: Call the IsDelinquent endpoint
		req, _ = http.NewRequest("GET", "/api/v1/customers/"+strconv.Itoa(int(customerID))+"/is_delinquent", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Verify the response status code and body
		Expect(resp.Code).To(Equal(http.StatusOK))
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())

		Expect(response["is_delinquent"]).To(BeTrue()) // Customer should be delinquent
	})

	ginkgo.It("should remain delinquent after running the scheduler multiple times", func() {
		// Step 1: Create a loan using the CreateLoan endpoint
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Smith",
			"email":       "johnsmith@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		payloadJSON, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Parse the loan creation response
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Fetch customer_id from the loans table
		customerID := getCustomerIDFromLoan(float64(loanID))

		// Step 2: Run the scheduler multiple times (simulate many overdue payments)
		for i := 1; i <= 3; i++ {
			currentDate := time.Now().AddDate(0, 0, 7*i) // Simulate multiple weeks later
			err := paymentUsecase.RunDaily(db, currentDate)
			Expect(err).ToNot(HaveOccurred())
		}

		// Step 3: Call the IsDelinquent endpoint
		req, _ = http.NewRequest("GET", "/api/v1/customers/"+strconv.Itoa(int(customerID))+"/is_delinquent", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Verify the response status code and body
		Expect(resp.Code).To(Equal(http.StatusOK))
		var response map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		Expect(err).ToNot(HaveOccurred())

		Expect(response["is_delinquent"]).To(BeTrue()) // Customer should remain delinquent
	})
})
