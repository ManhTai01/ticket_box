package rest

import (
	"log"
	"math"
	"strconv"
	auth "ticket_app/auth"
	"ticket_app/booking"
	"ticket_app/domain"
	"ticket_app/event"
	middleware "ticket_app/internal/rest/middleware"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Pagination struct {
	Page     int
	Limit    int
	Offset   int
	Total    int64
}

type PaginatedResponse struct {
	Data        interface{} `json:"data"`
	CurrentPage int         `json:"current_page"`
	TotalPages  int         `json:"total_pages"`
	TotalItems  int64       `json:"total_items"`
}

type BookingHandler struct {
	bookingService booking.BookingService
	eventService   event.EventService
	authService auth.AuthService
	validate       *validator.Validate
}

type CreateBookingRequest struct {
	EventID  uint `json:"event_id" validate:"required"`
	Quantity int  `json:"quantity" validate:"required,min=1"`
}

type UpdateBookingRequest struct {
	Status domain.BookingStatus `json:"status" validate:"required"`
}

type BookingResponse struct {
	ID         uint          `json:"id"`
	UserID     uint          `json:"user_id"`
	EventID    uint          `json:"event_id"`
	Quantity   int           `json:"quantity"`
	TotalPrice float64       `json:"total_price"`
	Status     string        `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	User       UserResponse  `json:"user"`
	Event      EventResponse `json:"event"`
}

func NewBookingHandler(app *fiber.App, bookingService booking.BookingService, authService auth.AuthService) *BookingHandler {
	handler := &BookingHandler{
		bookingService: bookingService,
		authService:    authService,
		validate:      validator.New(),	}

	app.Post("/bookings", handler.CreateBooking)
	app.Get("/bookings", handler.GetAllBookings)
	app.Get("/bookings/:id", handler.GetBookingById)
	app.Put("/bookings/:id", handler.UpdateBooking)
	app.Put("/bookings/:id/cancel", handler.CancelBooking)
	app.Put("/bookings/:id/confirm", handler.ConfirmBooking)

	return handler
}

func (h *BookingHandler) CreateBooking(c *fiber.Ctx) error {
	var req CreateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": err.Error()})
	}

	// Get userID from token
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	email, ok := claims["email"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}
	log.Println("email", email)
	userData, err := h.authService.FindByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Call service to create booking and handle payment queue
	booking, err := h.bookingService.CreateBooking(c.Context(), userData.ID, req.EventID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create booking"})
	}

	return c.Status(fiber.StatusCreated).JSON(
		BookingResponse{
			ID:         booking.ID,
			UserID:     booking.UserID,
			EventID:    booking.EventID,
			Quantity:   booking.Quantity,
			TotalPrice: booking.TotalPrice,
			Status:     string(booking.Status),
			CreatedAt:  booking.CreatedAt,
			UpdatedAt:  booking.UpdatedAt,
			User:       UserResponse{ID: booking.User.ID, Email: booking.User.Email},
			Event:      EventResponse{ID: booking.Event.ID, Name: booking.Event.Name, TicketPrice: booking.Event.TicketPrice},
		},
	)
}

func (h *BookingHandler) GetAllBookings(c *fiber.Ctx) error {
	pagination, ok := c.Locals("pagination").(middleware.Pagination)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Pagination info not found"})
	}

	totalItems, err := h.bookingService.CountBookings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	bookings, err := h.bookingService.GetAllBookingsWithPagination(pagination.Offset, pagination.Limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(pagination.Limit)))

	response := PaginatedResponse{
		Data:        bookings,
		CurrentPage: pagination.Page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
	}

	return c.JSON(response)
}

func (h *BookingHandler) GetBookingById(c *fiber.Ctx) error {
	log.Println("Getting booking by id")
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid booking ID"})
	}
	booking, err := h.bookingService.GetBookingById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Booking not found"})
	}
	return c.JSON(BookingResponse{
		ID:         booking.ID,
		UserID:     booking.UserID,
		EventID:    booking.EventID,
		Quantity:   booking.Quantity,
		TotalPrice: booking.TotalPrice,
		Status:     string(booking.Status),
		CreatedAt:  booking.CreatedAt,
		UpdatedAt:  booking.UpdatedAt,
		User:       UserResponse{ID: booking.User.ID, Email: booking.User.Email},
		Event:      EventResponse{ID: booking.Event.ID, Name: booking.Event.Name, TicketPrice: booking.Event.TicketPrice},
	})
}

func (h *BookingHandler) UpdateBooking(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid booking ID"})
	}

	var req UpdateBookingRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("err1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := h.validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": err.Error()})
	}
	booking, err := h.bookingService.GetBookingById(uint(id))
	if err != nil {
		log.Println("err3", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Booking not found"})
	}

	booking.Status = req.Status


	err = h.bookingService.UpdateBooking(booking)
	if err != nil {
		log.Println("err5", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update booking"})
	}

	return c.JSON(BookingResponse{
		ID:         booking.ID,
		UserID:     booking.UserID,
		EventID:    booking.EventID,
		Quantity:   booking.Quantity,
		TotalPrice: booking.TotalPrice,
		Status:     string(booking.Status),
		CreatedAt:  booking.CreatedAt,
		UpdatedAt:  booking.UpdatedAt,
		User:       UserResponse{ID: booking.User.ID, Email: booking.User.Email},
		Event:      EventResponse{ID: booking.Event.ID, Name: booking.Event.Name, TicketPrice: booking.Event.TicketPrice},
	})
}

func (h *BookingHandler) ConfirmBooking(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid booking ID"})
	}
	booking, err := h.bookingService.ConfirmBooking(uint(id))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to confirm booking"})
	}
	return c.JSON(booking)
}

func (h *BookingHandler) CancelBooking(c *fiber.Ctx) error {
	log.Println("CancelBooking")
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.Println("err1", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid booking ID"})
	}
	booking, err := h.bookingService.CancelBooking(uint(id))
	if err != nil {
		log.Println("err2", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to cancel booking"})
	}
	return c.JSON(booking)
}

