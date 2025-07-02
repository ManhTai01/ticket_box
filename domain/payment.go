package domain

import "time"

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "PENDING"
    PaymentStatusCompleted PaymentStatus = "COMPLETED"
    PaymentStatusFailed    PaymentStatus = "FAILED"
)

func (s PaymentStatus) Validate() bool {
    switch s {
    case PaymentStatusPending, PaymentStatusCompleted, PaymentStatusFailed:
        return true
    default:
        return false
    }
}

// Payment represents a payment entity
type Payment struct {
    ID        uint         `gorm:"primaryKey;autoIncrement" json:"id"`
    BookingID uint         `gorm:"not null;index" json:"booking_id"` // FK to Booking.ID
    Amount    float64      `gorm:"type:decimal(10,2);not null" json:"amount"`
    Status    PaymentStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
    CreatedAt time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
    UpdatedAt time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
    Booking   Booking      `gorm:"references:ID"` // Quan hệ ngược (optional)
}