package booking

import (
	"context"
	"errors"
	"log"
	"ticket_app/domain"
	"ticket_app/internal/queue"
	"ticket_app/internal/repository/booking"
	eventRepo "ticket_app/internal/repository/event"
	userRepo "ticket_app/internal/repository/user"
	payment "ticket_app/payment"
	"time"
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID uint, eventID uint, quantity int) (*domain.Booking, error)
	// GetAllBookings() ([]domain.Booking, error)
	GetBookingById(id uint) (*domain.Booking, error)
	UpdateBooking(booking *domain.Booking) error
	CountBookings() (int64, error)
	GetAllBookingsWithPagination(offset int, limit int) ([]domain.Booking, error)
	CancelBooking(id uint) (*domain.Booking, error)
	ConfirmBooking(id uint) (*domain.Booking, error)
}

type bookingService struct {
	bookingRepo booking.BookingRepository
	userRepo userRepo.UserRepository
	eventRepo eventRepo.EventRepository
	paymentService payment.PaymentService
	queueService *queue.QueueService
}

func NewBookingService(bookingRepo booking.BookingRepository, userRepo userRepo.UserRepository, eventRepo eventRepo.EventRepository) BookingService {
	return &bookingService{bookingRepo: bookingRepo, userRepo: userRepo, eventRepo: eventRepo}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID uint, eventID uint, quantity int) (*domain.Booking, error) {

	user, err := s.userRepo.FindById(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	event, err := s.eventRepo.FindById(eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	if event.Status != domain.EventStatusActive {
		return nil, errors.New("event is not active")
	}
	if event.TotalTickets < quantity {
		return nil, errors.New("not enough tickets available")
	}
	if event.StartDate.Before(time.Now()) {
		return nil, errors.New("event has already started")
	}

	booking := &domain.Booking{
		UserID: userID,
		EventID: eventID,
		Quantity: quantity,
		TotalPrice: float64(quantity) * event.TicketPrice,
		Status: domain.BookingStatusPending,
	}

	err = s.bookingRepo.Create(booking)
	if err != nil {
		return nil, err
	}
	// Cập nhật số lượng vé còn lại
	event.TotalTickets -= quantity
	if err := s.eventRepo.Update(event); err != nil {
		return nil, err
	}
	payment := domain.Payment{
		BookingID: booking.ID,
		Amount:    booking.TotalPrice,
		Status:    domain.PaymentStatusPending,
	}
	if err := s.paymentService.CreatePayment(&payment); err != nil {
		return nil, err
	}

	// Enqueue payment job
	paymentJob := queue.PaymentJob{
		BookingID: booking.ID,
		Amount:    booking.TotalPrice,
	}
	log.Println("Enqueuing payment job")
	if err := s.queueService.EnqueuePayment(paymentJob); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *bookingService) GetAllBookingsWithPagination(offset int, limit int) ([]domain.Booking, error) {
	return s.bookingRepo.FindAllWithPagination(offset, limit)
}
func (s *bookingService) CountBookings() (int64, error) {
	return s.bookingRepo.Count()
}

func (s *bookingService) GetBookingById(id uint) (*domain.Booking, error) {
	return s.bookingRepo.FindById(id)
}

func (s *bookingService) UpdateBooking(booking *domain.Booking) error {
	return s.bookingRepo.Update(booking)
}

func (s *bookingService) CancelBooking(id uint) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindById(id)
	if err != nil {
		return nil, err
	}
	if booking.Status != domain.BookingStatusPending {
		return nil, errors.New("booking is not pending")
	}
	booking.Status = domain.BookingStatusCancelled
	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, err
	}
	payment, err := s.paymentService.FindByBookingID(booking.ID)
	if err != nil {
		return nil, err
	}
	payment.Status = domain.PaymentStatusFailed
	if err := s.paymentService.ConfirmPayment(payment); err != nil {
		return nil, err
	}
	event := booking.Event
	event.TotalTickets += booking.Quantity
	if err := s.eventRepo.Update(&event); err != nil {
		return nil, err
	}
	
	return booking, nil
}

func (s *bookingService) ConfirmBooking(id uint) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindById(id)
	if err != nil {
		return nil, err
	}
	if booking.Status != domain.BookingStatusPending {
		return nil, errors.New("booking is not pending")
	}
	booking.Status = domain.BookingStatusConfirmed
	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, err
	}
	payment, err := s.paymentService.FindByBookingID(booking.ID)
	if err != nil {
		return nil, err
	}
	payment.Status = domain.PaymentStatusCompleted
	if err := s.paymentService.ConfirmPayment(payment); err != nil {
		return nil, err
	}
	return booking, nil
}