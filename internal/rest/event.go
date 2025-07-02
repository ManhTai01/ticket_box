package rest

import (
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"ticket_app/domain"
	event "ticket_app/event"
	middleware "ticket_app/internal/rest/middleware"
)

type EventHandler struct {
	eventService event.EventService
	validate     *validator.Validate
}

func NewEventHandler(app *fiber.App, eventService event.EventService) *EventHandler {
	// Khởi tạo handler với eventService và validate
	handler := &EventHandler{
		eventService: eventService,
		validate:     validator.New(),
	}

	// Đăng ký routes
	app.Post("/events", handler.CreateEvent)
	app.Get("/events", handler.GetAllEvents)
	app.Get("/events/remaining-tickets", handler.GetEventsWithRemainingTickets)
	app.Get("/events/:id", handler.GetEventById)
	app.Put("/events/:id", handler.UpdateEvent)
	app.Delete("/events/:id", handler.DeleteEvent)

	return handler
}

type CreateEventRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	TotalTickets int    `json:"total_tickets" validate:"required"`
	TicketPrice float64 `json:"ticket_price" validate:"required"`
}

type EventResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	TotalTickets int      `json:"total_tickets"`
	TicketPrice float64   `json:"ticket_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var req CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Chuyển request sang domain model và gọi service
	event := domain.Event{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   time.Time{}, // Sẽ được xử lý trong service
		EndDate:     time.Time{}, // Sẽ được xử lý trong service
		TotalTickets: req.TotalTickets,
		TicketPrice: req.TicketPrice,
	}

	err := h.eventService.CreateEvent(&event)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create event", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		TotalTickets: event.TotalTickets,
		TicketPrice: event.TicketPrice,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	})
}

func (h *EventHandler) GetAllEvents(c *fiber.Ctx) error {
	events, err := h.eventService.GetAllEvents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get events"})
	}

	var eventResponses []EventResponse
	for _, e := range events {
		eventResponses = append(eventResponses, EventResponse{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			StartDate:   e.StartDate,
			EndDate:     e.EndDate,
			TotalTickets: e.TotalTickets,
			TicketPrice: e.TicketPrice,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		})
	}
	return c.JSON(eventResponses)
}

func (h *EventHandler) GetEventsWithRemainingTickets(c *fiber.Ctx) error {
	pagination := middleware.Pagination{
		Page: c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
	}
	events, err := h.eventService.GetEventsWithRemainingTickets(pagination)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get events with remaining tickets"})
	}
	return c.JSON(events)
	
}
func (h *EventHandler) GetEventById(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	event, err := h.eventService.GetEventById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}
	if event == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	return c.JSON(EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		TotalTickets: event.TotalTickets,
		TicketPrice: event.TicketPrice,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	})
}

func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	var event domain.Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Set the ID from the URL parameter
	event.ID = uint(id)

	err = h.eventService.UpdateEvent(&event)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update event", "details": err.Error()})
	}

	return c.JSON(EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		TotalTickets: event.TotalTickets,
		TicketPrice: event.TicketPrice,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	})
}

func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid event ID"})
	}

	err = h.eventService.DeleteEvent(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete event", "details": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{"message": "Event deleted successfully"})
}