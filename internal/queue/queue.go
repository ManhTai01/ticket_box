package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"strconv"
	"ticket_app/domain"
	"ticket_app/internal/redis"
	bookingRepo "ticket_app/internal/repository/booking"
	eventRepo "ticket_app/internal/repository/event"
	paymentRepo "ticket_app/internal/repository/payment"
)

type QueueService struct {
	redis       *redis.Redis
	ctx         context.Context
	paymentRepo paymentRepo.PaymentRepository
	bookingRepo bookingRepo.BookingRepository
	eventRepo   eventRepo.EventRepository
}

type PaymentJob struct {
	BookingID uint    `json:"booking_id"`
	Amount    float64 `json:"amount"`
}

const (
	QueueName      = "payment_queue"
	PaymentTimeout = 15 * time.Minute
)

func NewQueueService(redisClient *redis.Redis, paymentRepo paymentRepo.PaymentRepository, bookingRepo bookingRepo.BookingRepository, eventRepo eventRepo.EventRepository) *QueueService {
	if redisClient == nil {
		log.Fatalf("Redis client is nil in NewQueueService")
	}
	log.Printf("QueueService created with redisClient: %p, GetClient: %p", redisClient, redisClient.GetClient())
	return &QueueService{
		redis:       redisClient,
		ctx:         context.Background(),
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
		eventRepo:   eventRepo,
	}
}

func (s *QueueService) EnqueuePayment(job PaymentJob) error {
	log.Printf("Enqueuing payment job for booking %d", job.BookingID)
	if s.redis == nil {
		return fmt.Errorf("QueueService redis instance is nil")
	}
	client := s.redis.GetClient()
	if client == nil {
		return fmt.Errorf("Redis client is nil")
	}
	log.Printf("Check Redis Client: %p", client)
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}
	log.Printf("Enqueuing job: %v", jobData)
	if err := client.RPush(s.ctx, QueueName, jobData).Err(); err != nil {
		return err
	}
	timeoutKey := fmt.Sprintf("payment:timeout:%d", job.BookingID)
	if err := client.Set(s.ctx, timeoutKey, "pending", PaymentTimeout).Err(); err != nil {
		return err
	}
	log.Printf("Successfully enqueued job for booking %d", job.BookingID)
	return nil
}

func (s *QueueService) StartWorker() {
	log.Printf("Starting worker with redisClient: %p", s.redis)
	for {
		if s.redis == nil {
			log.Printf("QueueService redis instance is nil, skipping worker iteration")
			time.Sleep(5 * time.Second)
			continue
		}
		client := s.redis.GetClient()
		if client == nil {
			log.Printf("Redis client is nil, skipping worker iteration")
			time.Sleep(5 * time.Second)
			continue
		}

		result, err := client.BLPop(s.ctx, 5*time.Second, QueueName).Result()
		if err != nil {
			log.Printf("Error popping from queue: %v", err)
			if err.Error() == "redis: connection closed" || err.Error() == "dial tcp 127.0.0.1:6379: connect: connection refused" {
				log.Printf("Redis connection lost, retrying in 5 seconds...")
				time.Sleep(5 * time.Second)
			}
			continue
		}
		if len(result) == 0 {
			continue
		}
		if len(result) < 2 {
			log.Printf("Invalid result format from queue: %v", result)
			continue
		}

		var job PaymentJob
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			log.Printf("Error unmarshaling job: %v", err)
			continue
		}

		if err := s.processPayment(job); err != nil {
			log.Printf("Error processing payment for booking %d: %v", job.BookingID, err)
		}
	}
}
func (s *QueueService) processPayment(job PaymentJob) error {
	log.Printf("Checking payment status for booking %d", job.BookingID)
	// Simulate checking payment status from external system (no update)
	payment, err := s.paymentRepo.FindByBookingID(job.BookingID)
	if err != nil || payment == nil {
		log.Printf("Payment for booking %d not found or error: %v", job.BookingID, err)
		return fmt.Errorf("payment not found or error")
	}

	// Assume status is updated externally, simulate for testing
	// In real case, replace with actual payment status check
	if payment.Status == domain.PaymentStatusCompleted {
		if err := s.updateBookingStatus(job.BookingID, string(domain.BookingStatusConfirmed)); err != nil {
			return err
		}
		timeoutKey := fmt.Sprintf("payment:timeout:%d", job.BookingID)
		if s.redis == nil {
			log.Printf("Redis instance is nil, cannot clear timeout key for booking %d", job.BookingID)
			return fmt.Errorf("Redis instance is nil")
		}
		if err := s.redis.GetClient().Del(s.ctx, timeoutKey).Err(); err != nil {
			log.Printf("Failed to clear timeout key for booking %d: %v", job.BookingID, err)
		}
	} else if payment.Status == domain.PaymentStatusFailed {
		if err := s.updateBookingStatus(job.BookingID, string(domain.BookingStatusCancelled)); err != nil {
			return err
		}
		timeoutKey := fmt.Sprintf("payment:timeout:%d", job.BookingID)
		if s.redis == nil {
			log.Printf("Redis instance is nil, cannot clear timeout key for booking %d", job.BookingID)
			return fmt.Errorf("Redis instance is nil")
		}
		if err := s.redis.GetClient().Del(s.ctx, timeoutKey).Err(); err != nil {
			log.Printf("Failed to clear timeout key for booking %d: %v", job.BookingID, err)
		}
	}
	return nil
}

func (s *QueueService) StartTimeoutChecker() {
	log.Printf("Starting timeout checker with redisClient: %p", s.redis)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if s.redis == nil {
			log.Printf("QueueService redis instance is nil, skipping timeout check")
			continue
		}
		client := s.redis.GetClient()
		if client == nil {
			log.Printf("Redis client is nil, skipping timeout check")
			continue
		}

		keys, err := client.Keys(s.ctx, "payment:timeout:*").Result()
		if err != nil {
			log.Printf("Error fetching timeout keys: %v", err)
			continue
		}

		for _, key := range keys {
			bookingIDStr := strings.TrimPrefix(key, "payment:timeout:")
			bookingID, err := strconv.Atoi(bookingIDStr)
			if err != nil {
				log.Printf("Invalid booking ID in key %s: %v", key, err)
				continue
			}

			// Check if timeout has expired (key still exists after 15 minutes)
			if exists := client.Exists(s.ctx, key).Val() > 0; exists {
				payment, err := s.paymentRepo.FindByBookingID(uint(bookingID))
				if err != nil || payment == nil {
					log.Printf("Payment for booking %d not found or error: %v, updating booking to FAILED", bookingID, err)
					if err := s.updateBookingStatus(uint(bookingID), "FAILED"); err != nil {
						log.Printf("Error updating booking %d to FAILED: %v", bookingID, err)
					}
					continue
				}
				if payment.Status != "PAID" {
					log.Printf("Payment for booking %d is %s, updating booking to FAILED", bookingID, payment.Status)
					if err := s.updateBookingStatus(uint(bookingID), "FAILED"); err != nil {
						log.Printf("Error updating booking %d to FAILED: %v", bookingID, err)
					}
				} else {
					log.Printf("Payment for booking %d is PAID, no action needed", bookingID)
				}
			}
		}
	}
}

func (s *QueueService) updateBookingStatus(bookingID uint, status string) error {
	booking, err := s.bookingRepo.FindById(bookingID)
	if err != nil {
		return err
	}
	if booking.Status == "status" {
		return nil // No update needed if status is already the same
	}
	booking.Status = "status"
	if status == "FAILED" {
		// Release tickets if failed
		event, err := s.eventRepo.FindByBookingID(booking.EventID)
		if err == nil {
			event.TotalTickets += booking.Quantity
			if err := s.eventRepo.Update(event); err != nil {
				log.Printf("Failed to release tickets for booking %d: %v", bookingID, err)
			}
		}
	}
	return s.bookingRepo.Update(booking)
}

