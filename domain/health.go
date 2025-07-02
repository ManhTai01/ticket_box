package domain

import "time"

// Health represents the health status of the application
type Health struct {
    Status      string    `json:"status"`
    PostgreSQL  string    `json:"postgresql"`
    Redis       string    `json:"redis"`
    LastChecked time.Time `json:"last_checked"`
}