package payment

import (
	"log"
	"ticket_app/domain"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	FindAll() ([]domain.Payment, error)
	FindById(id uint) (*domain.Payment, error)
	FindByBookingID(bookingID uint) (*domain.Payment, error)
	UpdatePayment(payment *domain.Payment) error
	DeletePayment(id uint) error

}

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) PaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) Create(payment *domain.Payment) error {
	return r.db.Create(payment).Error
}

func (r *GormPaymentRepository) FindAll() ([]domain.Payment, error) {
	var payments []domain.Payment
	if err := r.db.Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *GormPaymentRepository) FindById(id uint) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *GormPaymentRepository) FindByBookingID(bookingID uint) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.First(&payment, "booking_id = ?", bookingID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *GormPaymentRepository) UpdatePayment(payment *domain.Payment) error {
	log.Printf("Updating payment: %v", payment)
	return r.db.Save(payment).Error
}

func (r *GormPaymentRepository) DeletePayment(id uint) error {
	return r.db.Delete(&domain.Payment{}, id).Error
}

