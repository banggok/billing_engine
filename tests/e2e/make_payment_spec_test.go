package e2e_test

import (
	"billing_enginee/internal/model"
	"billing_enginee/internal/usecase"
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

var _ = ginkgo.Describe("MakePayment Endpoint", func() {
	var db *gorm.DB
	var sqlDB *sql.DB
	var router *gin.Engine
	var paymentUsecase usecase.PaymentUsecase

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Use the helper to initialize the environment
		env := helpers.InitializeTestEnvironment()
		db = env.DB
		sqlDB = env.SQLDB
		router = env.Router
		paymentUsecase = env.PaymentUsecase
	})
	// Tear down after each test
	ginkgo.AfterEach(func() {
		// Clean up the database by truncating tables
		// Use the helper function to truncate tables
		err := helpers.TruncateTables(db, "loans", "customers", "payments")
		Expect(err).ToNot(HaveOccurred(), "Failed to truncate tables before running tests")
		sqlDB.Close()
	})

	ginkgo.It("should make payment and update status to paid for week 1 and outstanding for week 2", func() {
		// Step 1: Create a loan
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

		// Parse response to get loan_id
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanID := loanResponse["loan_id"].(string)

		// Step 2: Make a payment for week 1
		req, _ = http.NewRequest("POST", "/api/v1/loans/"+loanID+"/payment?amount=110000", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Step 3: Verify payments in the database
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(payments[0].Status).To(Equal("paid"))
		Expect(payments[1].Status).To(Equal("outstanding"))
	})

	ginkgo.It("should update payment to paid for week 1, outstanding for week 2 after one scheduler run", func() {
		// Step 1: Create a loan
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

		// Parse response to get loan_id
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())
		loanID := loanResponse["loan_id"].(string)

		// Step 2: Run scheduler for one week ahead
		currentDate := time.Now().AddDate(0, 0, 8)
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Step 3: Make a payment
		req, _ = http.NewRequest("POST", "/api/v1/loans/"+loanID+"/payment?amount=110000", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Step 4: Verify payments in the database
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(payments[0].Status).To(Equal("paid"))
		Expect(payments[1].Status).To(Equal("outstanding"))
	})

	ginkgo.It("should update payment status for week 1-3 to paid and outstanding for week 4 after two scheduler runs", func() {
		// Step 1: Create a loan
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

		// Parse response to get loan_id
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())
		loanID := loanResponse["loan_id"].(string)

		// Step 2: Run scheduler for two weeks ahead (simulate two weeks of payments)
		currentDate := time.Now().AddDate(0, 0, 8) // 1st week
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		currentDate = time.Now().AddDate(0, 0, 15) // 2nd week
		err = paymentUsecase.RunDaily(db, currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Step 3: Make payment for the full outstanding balance (weeks 1, 2, and 3)
		req, _ = http.NewRequest("POST", "/api/v1/loans/"+loanID+"/payment?amount=330000", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Step 4: Verify payments in the database
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(payments[0].Status).To(Equal("paid"))
		Expect(payments[1].Status).To(Equal("paid"))
		Expect(payments[2].Status).To(Equal("paid"))
		Expect(payments[3].Status).To(Equal("outstanding"))
	})

	ginkgo.It("should mark all payments as paid and the loan as closed after multiple payments", func() {
		// Step 1: Create a loan
		termWeek := 2
		amount := 500000
		rates := 10
		payload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      amount,
			"term_weeks":  termWeek,
			"rates":       rates,
		}
		payloadJSON, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(payloadJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Parse response to get loan_id
		var loanResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanID := loanResponse["loan_id"].(string)

		// Step 2: Loop through and make payments until all payments are made
		paymentAmount := (amount + (amount * rates / 100)) / termWeek // The weekly payment amount (based on loan total/term_weeks)
		for week := 1; week <= termWeek; week++ {
			req, _ = http.NewRequest("POST", "/api/v1/loans/"+loanID+"/payment?amount="+strconv.FormatFloat(float64(paymentAmount), 'f', -1, 64), nil)
			resp = httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			// Step 3: Verify the payment for the current week
			var payments []model.Payment
			err := db.Where("loan_id = ?", loanID).Order("week").Find(&payments).Error
			Expect(err).ToNot(HaveOccurred())

			// Payments for the current week should be marked as "paid"
			for i := 0; i < week; i++ {
				Expect(payments[i].Status).To(Equal("paid"))
			}

			// The next payment (if any) should be "outstanding"
			if week < termWeek {
				Expect(payments[week].Status).To(Equal("outstanding"))
			}
		}

		// Step 4: After all payments, verify that all payments are marked as paid
		var finalPayments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week").Find(&finalPayments).Error
		Expect(err).ToNot(HaveOccurred())
		for _, payment := range finalPayments {
			Expect(payment.Status).To(Equal("paid"))
		}

		// Step 5: Verify that the loan status is "closed"
		var loan model.Loan
		err = db.Where("id = ?", loanID).First(&loan).Error
		Expect(err).ToNot(HaveOccurred())
		Expect(loan.Status).To(Equal("close"))
	})

})
