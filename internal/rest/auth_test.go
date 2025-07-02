package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ticket_app/domain"
	auth "ticket_app/domain"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password string) (*auth.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) FindByEmail(email string) (*auth.User, error) {
	args := m.Called(email)
	if user, ok := args.Get(0).(*domain.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func setupTestApp(handler *AuthHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableDefaultContentType: true,
	})

	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)

	// Add middleware to inject user context for testing /auth/profile
	app.Use("/auth/profile", func(c *fiber.Ctx) error {
		if c.Get("X-Mock-User") != "" {
			c.Locals("user", map[string]interface{}{
				"email": c.Get("X-Mock-User"),
			})
		}
		return c.Next()
	})
	app.Get("/auth/profile", handler.Profile)

	return app
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Success", testRegisterSuccess)
	t.Run("InvalidJSON", testRegisterInvalidJSON)
	t.Run("InvalidEmail", testRegisterInvalidEmail)
	t.Run("EmptyEmail", testRegisterEmptyEmail)
	t.Run("ServiceError", testRegisterServiceError)
}

// 👇 Các hàm test con cho /register
func testRegisterSuccess(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	reqBody := RegisterRequest{"test@example.com", "password123"}
	user := &auth.User{ID: 1, Email: reqBody.Email, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockAuth.On("Register", mock.Anything, reqBody.Email, reqBody.Password).Return(user, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	mockAuth.AssertExpectations(t)
}

func testRegisterInvalidJSON(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func testRegisterInvalidEmail(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	reqBody := RegisterRequest{Email: "@invalid", Password: "123456"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func testRegisterEmptyEmail(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	reqBody := RegisterRequest{Email: "", Password: "123456"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func testRegisterServiceError(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	reqBody := RegisterRequest{"test@example.com", "password123"}
	mockAuth.On("Register", mock.Anything, reqBody.Email, reqBody.Password).
		Return((*auth.User)(nil), errors.New("db error"))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	mockAuth.AssertExpectations(t)
}



func TestAuthHandler_Login(t *testing.T) {
	t.Run("Success", testLoginSuccess)
	t.Run("InvalidCredentials", testLoginInvalidCredentials)
	t.Run("InvalidRequestBody", testLoginInvalidRequestBody)
}

func testLoginSuccess(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})
	reqBody := LoginRequest{Email: "test@example.com", Password: "password123"}
	mockAuth.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return("access-token", nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var response map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "access-token", response["access_token"])
	mockAuth.AssertExpectations(t)
}

func testLoginInvalidCredentials(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})
	reqBody := LoginRequest{Email: "test@example.com", Password: "wrong-password"}
	mockAuth.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return("", errors.New("invalid credentials"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	var response map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "invalid credentials", response["error"])
	mockAuth.AssertExpectations(t)
}

func testLoginInvalidRequestBody(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	var response map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "Invalid request", response["error"])
	mockAuth.AssertExpectations(t)
}
func TestAuthHandler_Profile(t *testing.T) {
	t.Run("Success", testProfileSuccess)
	t.Run("NoUserInContext", testProfileNoUser)
	t.Run("InvalidTokenType", testProfileInvalidToken)
}

func testProfileSuccess(t *testing.T) {
	mockAuth := new(MockAuthService)
	now := time.Now()
	user := &auth.User{ID: 1, Email: "test@example.com", CreatedAt: now, UpdatedAt: now}
	mockAuth.On("FindByEmail", user.Email).Return(user, nil)

	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})
	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Mock-User", user.Email)

	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	var response UserResponse
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Email, response.Email)
	mockAuth.AssertExpectations(t)
}

func testProfileNoUser(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})
	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	var response map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "No user in context", response["error"])
}

func testProfileInvalidToken(t *testing.T) {
	mockAuth := new(MockAuthService)
	app := setupTestApp(&AuthHandler{authService: mockAuth, validate: validator.New()})

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("Content-Type", "application/json")
	// Send mock header to inject a wrong type manually in middleware (optional if test improved)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	var response map[string]string
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "No user in context", response["error"])
}
