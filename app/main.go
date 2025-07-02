package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"fmt"
	"ticket_app/auth"
	"ticket_app/booking"
	"ticket_app/domain"
	"ticket_app/event"
	"ticket_app/health"
	queueService "ticket_app/internal/queue"
	"ticket_app/internal/redis"
	bookingRepo "ticket_app/internal/repository/booking"
	eventRepo "ticket_app/internal/repository/event"
	paymentRepo "ticket_app/internal/repository/payment"
	userRepo "ticket_app/internal/repository/user"
	"ticket_app/internal/rest"
	"ticket_app/internal/rest/middleware"

	"github.com/joho/godotenv"
)

const (
	defaultTimeout = 30
	defaultAddress = ":9090"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Prepare database
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v. DSN: %s", err, dsn)
	}
	// Auto migrate to create all tables

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Booking{},
		&domain.Event{},
		&domain.Payment{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	} else {
		log.Println("Database migration completed successfully")
	}
	
	// Initialize Redis
	redisClient, err := redis.NewRedis()
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	if redisClient == nil {
		log.Fatalf("Redis client initialization returned nil")
	}
	log.Printf("Redis client initialized: %p, GetClient: %p", redisClient, redisClient.GetClient())
	defer redisClient.Close()

	// Register health service
	healthService := health.NewHealthService(db, redisClient.GetClient())

	// Initialize Fiber app
	app := fiber.New()

	// Middleware CORS
	app.Use(cors.New())

	authService := auth.NewAuthService(userRepo.NewGormUserRepository(db))
	eventService := event.NewEventService(eventRepo.NewGormEventRepository(db))
	bookingService := booking.NewBookingService(bookingRepo.NewGormBookingRepository(db), userRepo.NewGormUserRepository(db), eventRepo.NewGormEventRepository(db))
	// paymentService := payment.NewPaymentService(paymentRepo.NewGormPaymentRepository(db))
	log.Printf("Creating QueueService with redisClient: %p", redisClient)
	queueService := queueService.NewQueueService(redisClient, paymentRepo.NewGormPaymentRepository(db), bookingRepo.NewGormBookingRepository(db), eventRepo.NewGormEventRepository(db))
	rest.NewHealthHandlerFiber(app, healthService)
	rest.NewEventHandler(app, eventService)
	rest.NewAuthHandlerFiber(app, authService)


	app.Get("/auth/profile", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.Next() 
	})

	app.Use(middleware.JWTMiddleware())
	rest.NewBookingHandler(app, bookingService, authService)

	// Custom timeout middleware
	timeoutStr := os.Getenv("CONTEXT_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		log.Println("failed to parse timeout, using default timeout")
		timeout = defaultTimeout
	}
	timeoutContext := time.Duration(timeout) * time.Second
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("timeout", timeoutContext)
		return c.Next()
	})
	// Start queue worker and timeout checker in goroutines
	go queueService.StartTimeoutChecker()
	go queueService.StartWorker()
	// Start Server
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = defaultAddress
	}
	log.Fatal(app.Listen(address))
}