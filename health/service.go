package health

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthService defines the interface for health checking
type HealthService interface {
	CheckHealth(ctx context.Context) (map[string]interface{}, error)
}

// HealthServiceImpl implements HealthService
type HealthServiceImpl struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewHealthService(db *gorm.DB, redisClient *redis.Client) HealthService {
	return &HealthServiceImpl{
		db:          db,
		redisClient: redisClient,
	}
}

func (s *HealthServiceImpl) CheckHealth(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status":  "healthy",
		"details": map[string]interface{}{},
	}

	// Check PostgreSQL
	dbCheck := s.checkDB(ctx)
	if dbCheck != nil {
		health["status"] = "unhealthy"
		health["details"].(map[string]interface{})["db"] = dbCheck.Error()
	} else {
		health["details"].(map[string]interface{})["db"] = "ok"
	}

	// Check Redis
	redisCheck := s.checkRedis(ctx)
	if redisCheck != nil {
		health["status"] = "unhealthy"
		health["details"].(map[string]interface{})["redis"] = redisCheck.Error()
	} else {
		health["details"].(map[string]interface{})["redis"] = "ok"
	}

	if health["status"] == "unhealthy" {
		return health, errors.New("one or more services are unhealthy")
	}

	return health, nil
}

func (s *HealthServiceImpl) checkDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var ping int
	if err := s.db.Raw("SELECT 1").Scan(&ping).Error; err != nil {
		return err
	}
	return nil
}

func (s *HealthServiceImpl) checkRedis(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}