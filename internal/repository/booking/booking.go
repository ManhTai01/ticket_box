package booking

import (
	"log"
	"ticket_app/domain"

	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *domain.Booking) error
	FindAll() ([]domain.Booking, error)
	FindById(id uint) (*domain.Booking, error)
	Update(booking *domain.Booking) error
	Delete(id uint) error
	UpdateStatusByID(id uint, status domain.BookingStatus) error
	Count() (int64, error)
	FindAllWithPagination(offset int, limit int) ([]domain.Booking, error)
}

type GormBookingRepository struct {
	db *gorm.DB
}

func NewGormBookingRepository(db *gorm.DB) BookingRepository {
	return &GormBookingRepository{db: db}
}

func (r *GormBookingRepository) Create(booking *domain.Booking) error {
	log.Println("Creating booking:", booking)
	return r.db.Create(booking).Error
}


func (r *GormBookingRepository) FindAll() ([]domain.Booking, error) {
	var bookings []domain.Booking
	log.Println("Finding all bookings")
	err := r.db.
		Preload("User").
		Preload("Event").
		Find(&bookings).Error
	return bookings, err
}

func (r *GormBookingRepository) FindAllWithPagination(offset int, limit int) ([]domain.Booking, error) {
	var bookings []domain.Booking
	err := r.db.
		Preload("User").
		Preload("Event").
		Offset(offset).
		Limit(limit).
		Find(&bookings).Error 
	return bookings, err
}

func (r *GormBookingRepository) FindById(id uint) (*domain.Booking, error) {
	var booking domain.Booking
	if err := r.db.Preload("User").Preload("Event").First(&booking, id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *GormBookingRepository) Update(booking *domain.Booking) error {
	return r.db.Preload("User").Preload("Event").Save(booking).Error
}

func (r *GormBookingRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Booking{}, id).Error
}

func (r *GormBookingRepository) UpdateStatusByID(id uint, status domain.BookingStatus) error {
	return r.db.Model(&domain.Booking{}).Where("id = ?", id).Update("status", status).Error
}
func (r *GormBookingRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.Booking{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}