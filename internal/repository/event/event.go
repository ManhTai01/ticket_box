package event

import (
	"math"
	"ticket_app/domain"
	"ticket_app/internal/rest/middleware"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *domain.Event) error
	FindAll() ([]domain.Event, error)
	FindById(id uint) (*domain.Event, error)
	Update(event *domain.Event) error
	Delete(id uint) error
	FindByBookingID(bookingID uint) (*domain.Event, error)
	GetEventsWithRemainingTickets(pagination middleware.Pagination) (middleware.PaginatedResponse, error)
}

type GormEventRepository struct {
	db *gorm.DB
}

func NewGormEventRepository(db *gorm.DB) EventRepository {
	return &GormEventRepository{db: db}
}

func (r *GormEventRepository) Create(event *domain.Event) error {
	return r.db.Create(event).Error
}

func (r *GormEventRepository) FindAll() ([]domain.Event, error) {
	var events []domain.Event
	if err := r.db.Omit("Bookings", "EventStats").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventsWithRemainingTickets lấy danh sách event với số vé còn lại
func (r *GormEventRepository) GetEventsWithRemainingTickets(pagination middleware.Pagination) (middleware.PaginatedResponse, error) {
    var events []domain.EventWithRemainingTickets
    var response middleware.PaginatedResponse

    // Raw query cho PostgreSQL
    query := `
        SELECT events.*, 
               (events.total_tickets - COALESCE(COUNT(bookings.id), 0)) as remaining_tickets
        FROM events
        LEFT JOIN bookings 
            ON bookings.event_id = events.id 
            AND bookings.status IN ('PENDING', 'CONFIRMED')
        GROUP BY events.id
        ORDER BY events.id
        LIMIT $1 OFFSET $2
    `

    // Thực thi query với LIMIT và OFFSET
    err := r.db.Raw(query, pagination.Limit, pagination.Offset).Scan(&events).Error
    if err != nil {
        return response, err
    }

    // Đếm tổng số bản ghi
    var totalRows int64
    countQuery := `
        SELECT COUNT(DISTINCT events.id)
        FROM events
        LEFT JOIN bookings 
            ON bookings.event_id = events.id 
            AND bookings.status IN ('PENDING', 'CONFIRMED')
    `
    err = r.db.Raw(countQuery).Scan(&totalRows).Error
    if err != nil {
        return response, err
    }

    // Tính tổng số trang
    totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))

    // Gán kết quả
    response = middleware.PaginatedResponse{
        Data:        &events,
        CurrentPage: pagination.Page,
        TotalPages:  totalPages,
        TotalItems:  totalRows,
    }

    return response, nil
}


func (r *GormEventRepository) FindById(id uint) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.Omit("Bookings", "EventStats").First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *GormEventRepository) FindByBookingID(bookingID uint) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.Omit("Bookings", "EventStats").First(&event, "bookings.id = ?", bookingID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *GormEventRepository) Update(event *domain.Event) error {
	return r.db.Save(event).Error
}

func (r *GormEventRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Event{}, id).Error
}