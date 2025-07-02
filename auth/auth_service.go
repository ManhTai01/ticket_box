package auth

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"ticket_app/domain"
	"ticket_app/internal/repository/user"

	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error) // Chỉ trả về AccessToken
	FindByEmail(email string) (*domain.User, error)
}

type authService struct {
	userRepo user.UserRepository
	jwtKey   []byte
}

func NewAuthService(userRepo user.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtKey:   []byte("your-secret-key"), // Thay bằng key bảo mật trong production
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	log.Println("Registering user with email:", email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		log.Println("Error creating user:", err)
		return nil, err
	}

	// Không trả về password
	user.Password = ""
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}
	exp, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_TIME"))
	if err != nil {
		return "", err
	}
	// Tạo Access Token (hết hạn sau 15 phút)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Minute * time.Duration(exp)).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) FindByEmail(email string) (*domain.User, error) {
	log.Println("Finding user by email:", email)
	return s.userRepo.FindByEmail(email)
}