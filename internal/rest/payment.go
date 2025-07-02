package rest

import (
	"log"
	"ticket_app/payment"

	"strconv"
	"ticket_app/domain"

	validator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentService payment.PaymentService
	validate     *validator.Validate
}

type CreatePaymentRequest struct {
	BookingID uint `json:"booking_id" validate:"required"`
	Amount    int    `json:"amount" validate:"required"`
}

func NewPaymentHandler(app *fiber.App, paymentService payment.PaymentService) *PaymentHandler {
	handler := &PaymentHandler{
		paymentService: paymentService,
		validate:     validator.New(),
	}

	app.Post("/payments", handler.CreatePayment)
	app.Put("/payments/:id/confirm", handler.ConfirmPayment)
	app.Put("/payments/:id/cancel", handler.CancelPayment)

	return handler
}


func (h *PaymentHandler) CreatePayment(c *fiber.Ctx) error {
	var req CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	payment := domain.Payment{
		BookingID: req.BookingID,
		Amount: float64(req.Amount),
		Status: domain.PaymentStatusPending,
	}

	if err := h.validate.Struct(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validation failed", "details": err.Error()})
	}

	err := h.paymentService.CreatePayment(&payment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
	}

	return c.Status(fiber.StatusCreated).JSON(payment)
}

func (h *PaymentHandler) ConfirmPayment(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payment ID"})
	}

	payment := domain.Payment{ID: uint(id)}

	err = h.paymentService.ConfirmPayment(&payment)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}

	return c.Status(fiber.StatusOK).JSON(payment)
}

func (h *PaymentHandler) CancelPayment(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payment ID"})
	}

	payment := domain.Payment{ID: uint(id)}
	payment.Status = domain.PaymentStatusFailed
	err = h.paymentService.CancelPayment(&payment)
	if err != nil {
		log.Printf("Error canceling payment: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}

	return c.Status(fiber.StatusOK).JSON(payment)
}