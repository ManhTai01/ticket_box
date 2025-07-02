package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http/httptest"
	"testing"
	"ticket_app/domain"
	"ticket_app/event"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define mock services
type MockBookingService struct{ mock.Mock }

func (m *MockBookingService) CreateBooking(ctx context.Context, userID, eventID uint, quantity int) (*domain.Booking, error) {
	args := m.Called(ctx, userID, eventID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingService) GetBookingById(id uint) (*domain.Booking, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingService) UpdateBooking(b *domain.Booking) error {
	return m.Called(b).Error(0)
}

func (m *MockBookingService) CancelBooking(id uint) (*domain.Booking, error) {
	log.Println("CancelBooking")
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingService) CountBookings() (int64, error) {
	return 0, nil
}

func (m *MockBookingService) GetAllBookingsWithPagination(offset, limit int) ([]domain.Booking, error) {
	return []domain.Booking{}, nil
}

func (m *MockBookingService) ConfirmBooking(id uint) (*domain.Booking, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}





func setupBookingApp(bs *MockBookingService, as *MockAuthService, es event.EventService, ) *fiber.App {
	app := fiber.New()
	validate := validator.New()
	h := &BookingHandler{
		bookingService: bs,
		eventService:   es,
		authService:    as,
		validate:       validate,
	}

	app.Use(func(c *fiber.Ctx) error {
		claims := jwt.MapClaims{"email": "test@example.com"}
		token := &jwt.Token{Claims: claims}
		c.Locals("user", token)
		c.Locals("authService", as)
		return c.Next()
	})

	app.Post("/bookings", h.CreateBooking)
	app.Get("/bookings/:id", h.GetBookingById)
	app.Put("/bookings/:id", h.UpdateBooking)
	app.Put("/bookings/:id/cancel", h.CancelBooking)
	app.Put("/bookings/:id/confirm", h.ConfirmBooking)
	return app
}

// CreateBooking Tests
func TestCreateBooking(t *testing.T) {
	t.Run("Success", TestCreateBookingSuccess)
	t.Run("InvalidBody", TestCreateBookingInvalidBody)
	t.Run("ServiceError", TestCreateBookingServiceError)
}

func TestCreateBookingSuccess(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)
	authSvc.On("FindByEmail", "test@example.com").Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)
	bookingSvc.On("CreateBooking", mock.Anything, uint(1), uint(1), 2).
		Return(&domain.Booking{
			ID:         1,
			UserID:     1,
			EventID:    1,
			Quantity:   2,
			TotalPrice: 100,
			Status:     domain.BookingStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			User:       domain.User{ID: 1, Email: "test@example.com"},
			Event:      domain.Event{ID: 1, Name: "Event", TicketPrice: 50},
		}, nil)

	app := setupBookingApp(bookingSvc, authSvc, nil)
	body, _ := json.Marshal(map[string]interface{}{"event_id": 1, "quantity": 2})
	req := httptest.NewRequest("POST", "/bookings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestCreateBookingInvalidBody(t *testing.T) {
	app := setupBookingApp(&MockBookingService{}, &MockAuthService{}, nil)
	req := httptest.NewRequest("POST", "/bookings", bytes.NewBufferString(`invalid`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateBookingServiceError(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)
	authSvc.On("FindByEmail", "test@example.com").Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)
	bookingSvc.On("CreateBooking", mock.Anything, uint(1), uint(1), 2).
		Return(nil, errors.New("internal error"))

	app := setupBookingApp(bookingSvc, authSvc, nil)
	body, _ := json.Marshal(map[string]interface{}{"event_id": 1, "quantity": 2})
	req := httptest.NewRequest("POST", "/bookings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
}


func TestGetBookingById(t *testing.T) {
	t.Run("Success", TestGetBookingByIdSuccess)
	t.Run("NotFound", TestGetBookingByIdNotFound)
	t.Run("InvalidID", TestGetBookingByIdInvalidID)
}


func TestGetBookingByIdSuccess(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)
	authSvc.On("FindByEmail", "test@example.com").Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)
	bookingSvc.On("GetBookingById", uint(1)).Return(&domain.Booking{ID: 1, UserID: 1, EventID: 1}, nil)
	app := setupBookingApp(bookingSvc, authSvc, nil)
	req := httptest.NewRequest("GET", "/bookings/1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mock-User", "test@example.com")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetBookingByIdNotFound(t *testing.T) {

	svc := new(MockBookingService)
	svc.On("GetBookingById", uint(999)).Return(nil, errors.New("not found"))
	app := setupBookingApp(svc, nil, nil)
	req := httptest.NewRequest("GET", "/bookings/999", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestGetBookingByIdInvalidID(t *testing.T) {
	app := setupBookingApp(&MockBookingService{}, nil, nil)
	req := httptest.NewRequest("GET", "/bookings/abc", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}


func TestUpdateBooking(t *testing.T) {
	t.Run("Success", TestUpdateBookingSuccess)
	t.Run("InvalidBody", TestUpdateBookingInvalidBody)
	t.Run("NotFound", TestUpdateBookingNotFound)
}
func TestUpdateBookingSuccess(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)

	authSvc.On("FindByEmail", "test@example.com").
		Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)

	booking := &domain.Booking{
		ID:     1,
		UserID: 1,
		Status: domain.BookingStatusPending,
		User:   domain.User{ID: 1, Email: "test@example.com"},
		Event:  domain.Event{ID: 1, Name: "Event", TicketPrice: 50},
	}
	bookingSvc.On("GetBookingById", uint(1)).Return(booking, nil)
	bookingSvc.On("UpdateBooking", booking).Return(nil)

	app := setupBookingApp(bookingSvc, authSvc, nil)

	body, _ := json.Marshal(map[string]string{"status": "CONFIRMED"})
	t.Log("body", string(body)) // debug

	req := httptest.NewRequest("PUT", "/bookings/1", bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	log.Println("resp", resp)
	assert.Equal(t, 200, resp.StatusCode)
}


func TestUpdateBookingInvalidBody(t *testing.T) {
	app := setupBookingApp(&MockBookingService{}, nil, nil)
	req := httptest.NewRequest("PUT", "/bookings/1", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestUpdateBookingNotFound(t *testing.T) {
	svc := new(MockBookingService)
	svc.On("GetBookingById", uint(1)).Return(nil, errors.New("not found"))
	app := setupBookingApp(svc, nil, nil)

	body, _ := json.Marshal(map[string]string{"status": string(domain.BookingStatusConfirmed)})
	req := httptest.NewRequest("PUT", "/bookings/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestCancelBooking(t *testing.T) {
	t.Run("Success", TestCancelBookingSuccess)
	t.Run("NotFound", TestCancelBookingNotFound)
	t.Run("InvalidID", TestCancelBookingInvalidID)
}

func TestCancelBookingSuccess(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)
	authSvc.On("FindByEmail", "test@example.com").
		Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)

	booking := &domain.Booking{
		ID:     1,
		UserID: 1,
		Status: domain.BookingStatusPending,
		User:   domain.User{ID: 1, Email: "test@example.com"},
		Event:  domain.Event{ID: 1, Name: "Event", TicketPrice: 50},
	}

	bookingSvc.On("GetBookingById", uint(1)).Return(booking, nil)
	
	bookingSvc.On("CancelBooking", uint(1)).Return(booking, nil)
	app := setupBookingApp(bookingSvc, authSvc, nil)
	req := httptest.NewRequest("PUT", "/bookings/1/cancel", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	log.Println("resp", resp)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCancelBookingNotFound(t *testing.T) {
	svc := new(MockBookingService)
	svc.On("CancelBooking", uint(1)).Return(nil, errors.New("booking not found"))
	app := setupBookingApp(svc, nil, nil)
	req := httptest.NewRequest("PUT", "/bookings/1/cancel", nil)
	resp, _ := app.Test(req)
	log.Println("resp", resp)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestCancelBookingInvalidID(t *testing.T) {
	app := setupBookingApp(&MockBookingService{}, nil, nil)
	req := httptest.NewRequest("PUT", "/bookings/abc/cancel", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestConfirmBooking(t *testing.T) {
	t.Run("Success", TestConfirmBookingSuccess)
	t.Run("NotFound", TestConfirmBookingNotFound)
	t.Run("InvalidID", TestConfirmBookingInvalidID)
}

func TestConfirmBookingSuccess(t *testing.T) {
	bookingSvc := new(MockBookingService)
	authSvc := new(MockAuthService)
	authSvc.On("FindByEmail", "test@example.com").
		Return(&domain.User{ID: 1, Email: "test@example.com"}, nil)
	bookingSvc.On("GetBookingById", uint(1)).Return(&domain.Booking{ID: 1, UserID: 1, EventID: 1}, nil)
	bookingSvc.On("ConfirmBooking", uint(1)).Return(&domain.Booking{ID: 1, UserID: 1, EventID: 1}, nil)
	app := setupBookingApp(bookingSvc, authSvc, nil)
	req := httptest.NewRequest("PUT", "/bookings/1/confirm", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestConfirmBookingNotFound(t *testing.T) {
	svc := new(MockBookingService)
	svc.On("ConfirmBooking", uint(1)).Return(nil, errors.New("booking not found"))
	app := setupBookingApp(svc, nil, nil)
	req := httptest.NewRequest("PUT", "/bookings/1/confirm", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestConfirmBookingInvalidID(t *testing.T) {	
	app := setupBookingApp(&MockBookingService{}, nil, nil)
	req := httptest.NewRequest("PUT", "/bookings/abc/confirm", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, 400, resp.StatusCode)
}