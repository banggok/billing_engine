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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = ginkgo.Describe("Get Outstanding Endpoint", func() {
	var db *gorm.DB
	var sqlDB *sql.DB
	var router *gin.Engine
	var paymentRepo repository.PaymentRepository
	var loanUsecase usecase.LoanUsecase
	var paymentUsecase usecase.PaymentUsecase

	// Set up the test environment before each test
	ginkgo.BeforeEach(func() {
		// Initialize the test database
		db, sqlDB, _ = pkg.InitTestDB() // Assume this initializes a test DB
		// Migrate the database schema for testing, and handle any migration errors
		err := db.AutoMigrate(&model.Customer{}, &model.Loan{}, &model.Payment{})
		Expect(err).ToNot(HaveOccurred()) // Ensure that migrations succeed
		// Initialize repositories and use cases
		loanRepo := repository.NewLoanRepository(db)
		customerRepo := repository.NewCustomerRepository(db)
		paymentRepo = repository.NewPaymentRepository(db)
		loanUsecase = usecase.NewLoanUsecase(loanRepo, customerRepo, paymentRepo)
		paymentUsecase = usecase.NewPaymentUsecase(paymentRepo)

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

	ginkgo.It("should return 1 outstanding payment after loan creation", func() {
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
		err := json.Unmarshal(loanResp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred()) // Ensure conversion succeeded

		// Verify the first payment is outstanding and the rest are scheduled, ordered by week asc
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week ASC").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		for i := 0; i < len(payments); i++ {
			if payments[i].Week == 1 {
				Expect(payments[i].Status).To(Equal("outstanding"))
				continue
			}
			Expect(payments[i].Status).To(Equal("scheduled"))
		}

		// Step 2: Call the GetOutstanding endpoint
		outstandingReq, _ := http.NewRequest("GET", "/api/v1/loans/"+strconv.Itoa(loanID)+"/outstanding", nil)
		outstandingResp := httptest.NewRecorder()
		router.ServeHTTP(outstandingResp, outstandingReq)

		// Verify the response status code
		Expect(outstandingResp.Code).To(Equal(http.StatusOK))

		// Parse the GetOutstanding response
		var outstandingResponse map[string]interface{}
		err = json.Unmarshal(outstandingResp.Body.Bytes(), &outstandingResponse)
		Expect(err).ToNot(HaveOccurred())

		// Verify the outstanding amount, due date, and week
		Expect(outstandingResponse["loan_id"]).To(Equal(loanIDStr))
		Expect(outstandingResponse["outstanding_amount"]).To(BeEquivalentTo(110000.0))
		Expect(outstandingResponse["week"]).To(BeEquivalentTo(1))
	})

	ginkgo.It("should mark first payment as pending, second payment as outstanding after scheduler runs once", func() {
		// Step 1: Create a loan
		loanPayload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		loanPayloadJSON, _ := json.Marshal(loanPayload)
		loanReq, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(loanPayloadJSON))
		loanReq.Header.Set("Content-Type", "application/json")
		loanResp := httptest.NewRecorder()
		router.ServeHTTP(loanResp, loanReq)

		var loanResponse map[string]interface{}
		err := json.Unmarshal(loanResp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred())

		// Step 2: Run the scheduler once
		currentDate := time.Now().AddDate(0, 0, 8) // Move one week ahead
		err = paymentUsecase.RunDaily(currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Verify that the first payment is pending, second is outstanding, ordered by week asc
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week ASC").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		for i := 0; i < len(payments); i++ {
			if payments[i].Week == 1 {
				Expect(payments[i].Status).To(Equal("pending"))
				continue
			}
			if payments[i].Week == 2 {
				Expect(payments[i].Status).To(Equal("outstanding"))
				continue
			}

			Expect(payments[i].Status).To(Equal("scheduled"))
		}

		// Step 3: Call the GetOutstanding endpoint and verify pending payment
		outstandingReq, _ := http.NewRequest("GET", "/api/v1/loans/"+strconv.Itoa(loanID)+"/outstanding", nil)
		outstandingResp := httptest.NewRecorder()
		router.ServeHTTP(outstandingResp, outstandingReq)

		// Verify the response status code and outstanding payment
		Expect(outstandingResp.Code).To(Equal(http.StatusOK))

		var outstandingResponse map[string]interface{}
		err = json.Unmarshal(outstandingResp.Body.Bytes(), &outstandingResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(outstandingResponse["loan_id"]).To(Equal(loanIDStr))
		Expect(outstandingResponse["outstanding_amount"]).To(BeEquivalentTo(110000.0)) // Single pending payment amount
		Expect(outstandingResponse["week"]).To(BeEquivalentTo(1))                      // Week 1 (first payment is pending)
	})

	ginkgo.It("should mark first two payments as pending, third one as outstanding after scheduler runs twice", func() {
		// Step 1: Create a loan
		loanPayload := map[string]interface{}{
			"customer_id": 1,
			"name":        "John Doe",
			"email":       "johndoe@example.com",
			"amount":      5000000,
			"term_weeks":  50,
			"rates":       10,
		}
		loanPayloadJSON, _ := json.Marshal(loanPayload)
		loanReq, _ := http.NewRequest("POST", "/api/v1/loans", bytes.NewBuffer(loanPayloadJSON))
		loanReq.Header.Set("Content-Type", "application/json")
		loanResp := httptest.NewRecorder()
		router.ServeHTTP(loanResp, loanReq)

		var loanResponse map[string]interface{}
		err := json.Unmarshal(loanResp.Body.Bytes(), &loanResponse)
		Expect(err).ToNot(HaveOccurred())

		loanIDStr := loanResponse["loan_id"].(string)
		loanID, err := strconv.Atoi(loanIDStr)
		Expect(err).ToNot(HaveOccurred())

		// Step 2: Run the scheduler twice
		currentDate := time.Now().AddDate(0, 0, 8) // Move one week ahead
		err = paymentUsecase.RunDaily(currentDate)
		Expect(err).ToNot(HaveOccurred())
		currentDate = time.Now().AddDate(0, 0, 15) // Move another week ahead
		err = paymentUsecase.RunDaily(currentDate)
		Expect(err).ToNot(HaveOccurred())

		// Verify that the first two payments are pending, third is outstanding, ordered by week asc
		var payments []model.Payment
		err = db.Where("loan_id = ?", loanID).Order("week ASC").Find(&payments).Error
		Expect(err).ToNot(HaveOccurred())
		for i := 0; i < len(payments); i++ {
			if payments[i].Week == 1 || payments[i].Week == 2 {
				Expect(payments[i].Status).To(Equal("pending"))
				continue
			}
			if payments[i].Week == 3 {
				Expect(payments[i].Status).To(Equal("outstanding"))
				continue
			}

			Expect(payments[i].Status).To(Equal("scheduled"))
		}

		// Step 3: Call the GetOutstanding endpoint and verify outstanding amount includes pending and outstanding payments
		outstandingReq, _ := http.NewRequest("GET", "/api/v1/loans/"+strconv.Itoa(loanID)+"/outstanding", nil)
		outstandingResp := httptest.NewRecorder()
		router.ServeHTTP(outstandingResp, outstandingReq)

		// Verify the response status code
		Expect(outstandingResp.Code).To(Equal(http.StatusOK))

		var outstandingResponse map[string]interface{}
		err = json.Unmarshal(outstandingResp.Body.Bytes(), &outstandingResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(outstandingResponse["loan_id"]).To(Equal(loanIDStr))

		// Outstanding amount should sum both pending and outstanding payments (2 * 110,000 + 110,000)
		Expect(outstandingResponse["outstanding_amount"]).To(BeEquivalentTo(330000.0))
		Expect(outstandingResponse["week"]).To(BeEquivalentTo(3)) // Week 3 (third payment is outstanding)
	})
})
