package domain

import "time"

type EventStatus string

const (
	EventStatusActive EventStatus = "ACTIVE"
	EventStatusInactive EventStatus = "INACTIVE"
)

// Event represents an event entity
type Event struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"type:varchar(255);not null" json:"name"`
    Description string    `gorm:"type:text" json:"description"`
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    TotalTickets int      `gorm:"not null" json:"total_tickets"`
    TicketPrice float64   `gorm:"type:decimal(10,2);not null" json:"ticket_price"`
    CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
    UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
    Bookings    []*Booking `gorm:"foreignKey:EventID"` // Quan hệ 1-n với Booking
    Status      EventStatus    `gorm:"type:varchar(255);not null;default:'ACTIVE'" json:"status"`
    // EventStats  *EventStats `gorm:"foreignKey:EventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Quan hệ 1-1 với EventStats

}



// EventWithRemainingTickets chứa thông tin Event và số vé còn lại
type EventWithRemainingTickets struct {
    Event
    RemainingTickets int64 `json:"remaining_tickets"`
}