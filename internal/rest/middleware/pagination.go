package middleware

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Pagination struct {
	Page     int
	Limit    int
	Offset   int
	Total    int64
}

type PaginatedResponse struct {
	Data        interface{} `json:"data"`
	CurrentPage int         `json:"current_page"`
	TotalPages  int         `json:"total_pages"`
	TotalItems  int64       `json:"total_items"`
}

func PaginationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only apply pagination for GET requests
		if c.Method() != http.MethodGet {
			return c.Next()
		}

		// Parse query parameters
		page, err := strconv.Atoi(c.Query("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(c.Query("limit", "10"))
		if err != nil || limit < 1 || limit > 100 {
			limit = 10
		}

		// Calculate offset
		offset := (page - 1) * limit

		// Store pagination info in context
		pagination := Pagination{
			Page:   page,
			Limit:  limit,
			Offset: offset,
		}
		c.Locals("pagination", pagination)

		// Proceed to handler
		return c.Next()
	}
}