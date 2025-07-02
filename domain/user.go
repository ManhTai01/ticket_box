package domain

import "time"

// User represents a user entity
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Email     string    `gorm:"type:varchar(255);unique;not null" json:"email"`
    Password  string    `gorm:"type:varchar(255);not null" json:"-"` // Lưu password đã hash       
    CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
    UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
    Bookings  []Booking `gorm:"foreignKey:UserID"` // Quan hệ 1-n với Booking
}