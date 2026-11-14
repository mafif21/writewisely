package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

const BasePath = "/api/v1"

// RegisterRoutes mounts the Swagger UI and all API routes on app.
func RegisterRoutes(app *fiber.App, health *HealthHandler) {
	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group(BasePath)
	api.Get("/health", health.Health)
}
