package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	auth "ticket_app/auth"
)

type AuthHandler struct {
	authService auth.AuthService
	validate    *validator.Validate
}

func NewAuthHandlerFiber(app *fiber.App, authService auth.AuthService) {
	validate := validator.New()
	handler := &AuthHandler{
		authService: authService,
		validate:    validate,
	}

	// Middleware kiểm tra JSON trước route /register
	app.Post("/register", func(c *fiber.Ctx) error {
		if c.Get("Content-Type") == "application/json" {
			body := c.Body()
			if len(body) > 0 {
				var temp interface{}
				if err := json.Unmarshal(body, &temp); err != nil {
					log.Println("JSON check middleware: Invalid JSON:", err)
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
				}
			}
		}
		return c.Next()
	}, handler.Register)
	app.Post("/login", handler.Login)
	app.Get("/auth/profile", handler.Profile)
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {

	// Lấy raw body và parse JSON
	body := c.Body()
	var req RegisterRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			log.Println("JSON Unmarshal error:", err)
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
		}
	}

	log.Println("Parsed request:", req)
	if err := h.validate.Struct(req); err != nil {
		log.Println("Validation error:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.authService.Register(c.Context(), req.Email, req.Password)
	log.Println("AuthService.Register returned user:", user, "error:", err)
	if err != nil {
		log.Println("Register error:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if user == nil {
		log.Println("User is nil, returning internal server error")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	log.Println("Register success, response:", response)
	return c.Status(http.StatusCreated).JSON(response)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	accessToken, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"access_token": accessToken})
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {

	userInterface := c.Locals("user")
	if userInterface == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No user in context"})
	}

	claims, ok := userInterface.(map[string]interface{})
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token type"})
	}

	email, ok := claims["email"].(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Email not found in token claims"})
	}

	user, err := h.authService.FindByEmail(email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return c.JSON(response)
}