package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"ticket_app/domain"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) CreatePayment(payment *domain.Payment) error {
	return m.Called(payment).Error(0)
}

func (m *MockPaymentService) ConfirmPayment(payment *domain.Payment) error {
	return m.Called(payment).Error(0)
}

func (m *MockPaymentService) CancelPayment(payment *domain.Payment) error {
	return m.Called(payment).Error(0)
}

func (m *MockPaymentService) FindByBookingID(bookingID uint) (*domain.Payment, error) {
	return m.Called(bookingID).Get(0).(*domain.Payment), m.Called(bookingID).Error(1)
}


func setupPaymentApp(ps *MockPaymentService, as *MockAuthService) *fiber.App {
	app := fiber.New()
	validate := validator.New()
	h := &PaymentHandler{
		paymentService: ps,
		validate: validate,
	}

	app.Use(func(c *fiber.Ctx) error {
		claims := jwt.MapClaims{"email": "test@example.com"}
		token := &jwt.Token{Claims: claims}
		c.Locals("user", token)
		c.Locals("authService", as)
		return c.Next()
	})

	app.Post("/payments", h.CreatePayment)
	app.Put("/payments/:id/confirm", h.ConfirmPayment)
	app.Put("/payments/:id/cancel", h.CancelPayment)

	return app
}

func TestCreatePayment(t *testing.T) {
	t.Run("Success", TestCreatePaymentSuccess)
	t.Run("InvalidBody", TestCreatePaymentInvalidBody)
	t.Run("ServiceError", TestCreatePaymentServiceError)
}

func TestCreatePaymentSuccess(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("CreatePayment", mock.Anything).Return(nil)
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)

	req := httptest.NewRequest("POST", "/payments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCreatePaymentInvalidBody(t *testing.T) {
	app := setupPaymentApp(&MockPaymentService{}, nil)
	req := httptest.NewRequest("POST", "/payments", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreatePaymentServiceError(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	
	paymentSvc.On("CreatePayment", mock.Anything).Return(errors.New("service error"))
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("POST", "/payments", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestConfirmPayment(t *testing.T) {
	t.Run("Success", TestConfirmPaymentSuccess)
	t.Run("NotFound", TestConfirmPaymentNotFound)
	t.Run("InvalidID", TestConfirmPaymentInvalidID)
}

func TestConfirmPaymentSuccess(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("ConfirmPayment", mock.Anything).Return(nil)
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/1/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

}

func TestConfirmPaymentNotFound(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("ConfirmPayment", mock.Anything).Return(errors.New("payment not found"))
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/1/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestConfirmPaymentInvalidID(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("ConfirmPayment", mock.Anything).Return(errors.New("payment not found"))
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/abc/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCancelPayment(t *testing.T) {
	t.Run("Success", TestCancelPaymentSuccess)
	t.Run("NotFound", TestCancelPaymentNotFound)
	t.Run("InvalidID", TestCancelPaymentInvalidID)
}

func TestCancelPaymentSuccess(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("CancelPayment", mock.Anything).Return(nil)
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/1/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCancelPaymentNotFound(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("CancelPayment", mock.Anything).Return(errors.New("payment not found"))
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/1/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestCancelPaymentInvalidID(t *testing.T) {
	paymentSvc := new(MockPaymentService)
	authSvc := new(MockAuthService)
	paymentSvc.On("CancelPayment", mock.Anything).Return(errors.New("payment not found"))
	reqBody := CreatePaymentRequest{
		BookingID: 1,
		Amount: 100,
	}
	body, _ := json.Marshal(reqBody)
	app := setupPaymentApp(paymentSvc, authSvc)
	req := httptest.NewRequest("PUT", "/payments/abc/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
