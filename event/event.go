package event

import (
	"log"
	"ticket_app/domain"
	eventRepo "ticket_app/internal/repository/event"
	"ticket_app/internal/rest/middleware"
	"time"

	"github.com/go-playground/validator/v10"
)

// EventService định nghĩa các phương thức của service
type EventService interface {
	CreateEvent(event *domain.Event) error
	GetAllEvents() ([]domain.Event, error)
	GetEventById(id uint) (*domain.Event, error)
	UpdateEvent(event *domain.Event) error
	DeleteEvent(id uint) error
	GetEventsWithRemainingTickets(pagination middleware.Pagination) (middleware.PaginatedResponse, error)
}

// eventService triển khai EventService
type eventService struct {
	validate *validator.Validate
	eventRepo eventRepo.EventRepository
}

// NewEventService tạo instance của EventService
func NewEventService(eventRepo eventRepo.EventRepository) EventService {
	return &eventService{
		validate: validator.New(),
		eventRepo: eventRepo,
	}
}

// CreateEvent xử lý logic tạo sự kiện
func (s *eventService) CreateEvent(event *domain.Event) error {
	// Validate input
	if err := s.validate.Struct(event); err != nil {
		return err
	}

	// Gán StartDate mặc định nếu không truyền
	if event.StartDate.IsZero() {
		event.StartDate = time.Now()
	}
	// Gán EndDate mặc định nếu không truyền (1 tháng sau StartDate)
	if event.EndDate.IsZero() {
		event.EndDate = event.StartDate.Add(30 * 24 * time.Hour)
	}
	err := s.eventRepo.Create(event)
	if err != nil {
		return err
	}
	log.Println("Event created successfully")
	return nil 
}

// GetAllEvents lấy tất cả sự kiện		
func (s *eventService) GetAllEvents() ([]domain.Event, error) {
	log.Println("Getting all events")
	events, err := s.eventRepo.FindAll()
	if err != nil {
		return nil, err
	}
	log.Println("Events fetched successfully")
	return events, nil 
}

// GetEventById lấy sự kiện theo ID
func (s *eventService) GetEventById(id uint) (*domain.Event, error) {
	log.Println("Getting event by id")
	event, err := s.eventRepo.FindById(id)
	if err != nil {
		return nil, err
	}
	log.Println("Event fetched successfully")
	return event, nil 
}

// UpdateEvent cập nhật sự kiện
func (s *eventService) UpdateEvent(event *domain.Event) error {
	log.Println("Updating event")
	err := s.eventRepo.Update(event)
	if err != nil {
		return err
	}
	log.Println("Event updated successfully")
	return nil 
}

// DeleteEvent xóa sự kiện
func (s *eventService) DeleteEvent(id uint) error {
	log.Println("Deleting event")
	err := s.eventRepo.Delete(id)
	if err != nil {
		return err
	}
	log.Println("Event deleted successfully")
	return nil 
}

func (s *eventService) GetEventsWithRemainingTickets(pagination middleware.Pagination) (middleware.PaginatedResponse, error) {
	log.Println("Getting events with remaining tickets")
	events, err := s.eventRepo.GetEventsWithRemainingTickets(pagination)
	if err != nil {
		return middleware.PaginatedResponse{}, err
	}
	log.Println("Events fetched successfully")
	return events, nil
}