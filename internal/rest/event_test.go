package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"ticket_app/domain"
	middleware "ticket_app/internal/rest/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) CreateEvent(event *domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventService) GetAllEvents() ([]domain.Event, error) {
	args := m.Called()
	return args.Get(0).([]domain.Event), args.Error(1)
}

func (m *MockEventService) GetEventById(id uint) (*domain.Event, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Event), args.Error(1)
}
func (m *MockEventService) UpdateEvent(event *domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventService) DeleteEvent(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEventService) GetEventsWithRemainingTickets(pagination middleware.Pagination) (middleware.PaginatedResponse, error) {
	args := m.Called(pagination)
	return args.Get(0).(middleware.PaginatedResponse), args.Error(1)
}

func setupEventApp(svc *MockEventService) *fiber.App {
	app := fiber.New()
	NewEventHandler(app, svc)
	return app
}

func TestCreateEvent(t *testing.T) {
	t.Run("Success", testCreateEventSuccess)
	t.Run("InvalidBody", testCreateEventInvalidBody)
	t.Run("ServiceError", testCreateEventServiceError)
}

func testCreateEventSuccess(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("CreateEvent", mock.Anything).Return(nil)

	reqBody := CreateEventRequest{
		Name: "Concert", Description: "Live", TotalTickets: 100, TicketPrice: 50.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func testCreateEventInvalidBody(t *testing.T) {
	app := setupEventApp(new(MockEventService))
	req := httptest.NewRequest("POST", "/events", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func testCreateEventServiceError(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("CreateEvent", mock.Anything).Return(errors.New("create error"))
	reqBody := CreateEventRequest{
		Name: "Concert", Description: "Live", TotalTickets: 100, TicketPrice: 50.0,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
func TestGetAllEvents(t *testing.T) {
	t.Run("Success", testGetAllEventsSuccess)
	t.Run("ServiceError", testGetAllEventsServiceError)
	t.Run("EmptyList", testGetAllEventsEmptyList)
}

func testGetAllEventsSuccess(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("GetAllEvents").Return([]domain.Event{{ID: 1, Name: "Concert"}}, nil)
	req := httptest.NewRequest("GET", "/events", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func testGetAllEventsServiceError(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("GetAllEvents").Return([]domain.Event(nil), errors.New("error"))
	req := httptest.NewRequest("GET", "/events", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func testGetAllEventsEmptyList(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("GetAllEvents").Return([]domain.Event{}, nil)
	req := httptest.NewRequest("GET", "/events", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetEventById(t *testing.T) {
	t.Run("Success", testGetEventByIdSuccess)
	t.Run("NotFound", testGetEventByIdNotFound)
	t.Run("InvalidID", testGetEventByIdInvalidID)
}

func testGetEventByIdSuccess(t *testing.T) {
	mock := new(MockEventService)

		event := &domain.Event{
			ID:           1,
			Name:         "Tech Conference",
			Description:  "Annual tech conference",
			StartDate:    time.Now().Add(24 * time.Hour),
			EndDate:      time.Now().Add(48 * time.Hour),
			TotalTickets: 200,
			TicketPrice:  99.99,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mock.On("GetEventById", uint(1)).Return(event, nil)

		app := setupEventApp(mock)
		req := httptest.NewRequest("GET", "/events/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var got domain.Event
		json.NewDecoder(resp.Body).Decode(&got)
		assert.Equal(t, event.ID, got.ID)
		assert.Equal(t, event.Name, got.Name)
		assert.Equal(t, event.Description, got.Description)
		assert.Equal(t, event.TotalTickets, got.TotalTickets)
		assert.Equal(t, event.TicketPrice, got.TicketPrice)
}

func testGetEventByIdNotFound(t *testing.T) {
	
	mock_event := new(MockEventService)
	mock_event.On("GetEventById", uint(999)).Return(nil, errors.New("not found"))

	app := setupEventApp(mock_event)
	req := httptest.NewRequest("GET", "/events/999", nil)
	resp, err := app.Test(req)
	log.Println("resp", resp)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func testGetEventByIdInvalidID(t *testing.T) {
	mock := new(MockEventService)
	mock.On("GetEventById", uint(999)).Return(nil, errors.New("not found"))

	app := setupEventApp(mock)
	req := httptest.NewRequest("GET", "/events/999", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestUpdateEvent(t *testing.T) {
	t.Run("Success", testUpdateEventSuccess)
	t.Run("InvalidBody", testUpdateEventInvalidBody)
	t.Run("ServiceError", testUpdateEventServiceError)
}

func testUpdateEventSuccess(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("UpdateEvent", mock.Anything).Return(nil)
	body, _ := json.Marshal(domain.Event{Name: "Updated"})
	req := httptest.NewRequest("PUT", "/events/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func testUpdateEventInvalidBody(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	req := httptest.NewRequest("PUT", "/events/1", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func testUpdateEventServiceError(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("UpdateEvent", mock.Anything).Return(errors.New("fail"))
	body, _ := json.Marshal(domain.Event{Name: "Updated"})
	req := httptest.NewRequest("PUT", "/events/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteEvent(t *testing.T) {
	t.Run("Success", testDeleteEventSuccess)
	t.Run("InvalidID", testDeleteEventInvalidID)
	t.Run("ServiceError", testDeleteEventServiceError)
}

func testDeleteEventSuccess(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("DeleteEvent", uint(1)).Return(nil)
	req := httptest.NewRequest("DELETE", "/events/1", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func testDeleteEventInvalidID(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	req := httptest.NewRequest("DELETE", "/events/abc", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func testDeleteEventServiceError(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("DeleteEvent", uint(2)).Return(errors.New("fail"))
	req := httptest.NewRequest("DELETE", "/events/2", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetEventsWithRemainingTickets(t *testing.T) {
	t.Run("Success", testGetEventsWithRemainingTicketsSuccess)
	t.Run("ServiceError", testGetEventsWithRemainingTicketsServiceError)
}

func testGetEventsWithRemainingTicketsSuccess(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("GetEventsWithRemainingTickets", mock.Anything).Return(middleware.PaginatedResponse{
		Data: []domain.Event{{ID: 1, Name: "Concert"}},
		CurrentPage: 1,
		TotalPages: 1,
		TotalItems: 1,
	}, nil)
	req := httptest.NewRequest("GET", "/events/remaining-tickets", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func testGetEventsWithRemainingTicketsServiceError(t *testing.T) {
	mock_event := new(MockEventService)
	app := setupEventApp(mock_event)
	mock_event.On("GetEventsWithRemainingTickets", mock.Anything).Return(middleware.PaginatedResponse{}, errors.New("error"))
	req := httptest.NewRequest("GET", "/events/remaining-tickets", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}