package payment

import (
	"fmt"
	"ticket_app/domain"
	"ticket_app/internal/repository/payment"
	"time"
)

type PaymentService interface {
	CreatePayment(payment *domain.Payment) error
	ConfirmPayment(payment *domain.Payment) error
	CancelPayment(payment *domain.Payment) error
	FindByBookingID(bookingID uint) (*domain.Payment, error)
}

type paymentService struct {
	paymentRepo payment.PaymentRepository
}

func NewPaymentService(paymentRepo payment.PaymentRepository) PaymentService {
	return &paymentService{paymentRepo: paymentRepo}
}

func (s *paymentService) CreatePayment(payment *domain.Payment) error {
	return s.paymentRepo.Create(payment)
}

func (s *paymentService) ConfirmPayment(payment *domain.Payment) error {
	payment, err := s.paymentRepo.FindById(payment.ID)
	if err != nil {
		return fmt.Errorf("payment not found")
	}
	payment.Status = domain.PaymentStatusCompleted
	payment.UpdatedAt = time.Now()
	return s.paymentRepo.UpdatePayment(payment)
}

func (s *paymentService) CancelPayment(payment *domain.Payment) error {
	payment, err := s.paymentRepo.FindByBookingID(payment.BookingID)
	if err != nil {
		return fmt.Errorf("payment not found")
	}
	return s.paymentRepo.UpdatePayment(payment)
}
func (s *paymentService) FindByBookingID(bookingID uint) (*domain.Payment, error) {
	return s.paymentRepo.FindByBookingID(bookingID)
}