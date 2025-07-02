package rest

import (
	"ticket_app/health"

	"github.com/gofiber/fiber/v2"
)

func NewHealthHandlerFiber(app *fiber.App, healthService health.HealthService) {
	app.Get("/health", func(c *fiber.Ctx) error {
		health, err := healthService.CheckHealth(c.Context())
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(health)
		}
		return c.JSON(health)
	})
}