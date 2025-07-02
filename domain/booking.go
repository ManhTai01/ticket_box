package domain

import "time"

// BookingStatus represents the status of a booking
type BookingStatus string

const (
    BookingStatusPending   BookingStatus = "PENDING"
    BookingStatusConfirmed BookingStatus = "CONFIRMED"
    BookingStatusCancelled BookingStatus = "CANCELLED"
)

func (s BookingStatus) Validate() bool {
    switch s {
    case BookingStatusPending, BookingStatusConfirmed, BookingStatusCancelled:
        return true
    default:
        return false
    }
}

// Booking represents a ticket booking entity
type Booking struct {
    ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID      uint         `gorm:"not null;index" json:"user_id"` // FK to User.ID
    EventID     uint         `gorm:"not null;index" json:"event_id"` // FK to Event.ID
    Quantity    int          `gorm:"not null" json:"quantity"`
    TotalPrice  float64      `gorm:"not null" json:"total_price"`
    Status      BookingStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
    CreatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
    UpdatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
    User        User         `gorm:"references:ID"` // Quan hệ ngược (optional)
    Event       Event        `gorm:"references:ID"` // Quan hệ ngược (optional)
}