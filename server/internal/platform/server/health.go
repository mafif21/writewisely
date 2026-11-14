package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"writewisely/server/internal/common/constants"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthResponse struct {
	Status string `json:"status" validate:"required" example:"ok" enums:"ok,degraded"`
	DB     string `json:"db" validate:"required" example:"up" enums:"up,down"`
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health reports API status and database connectivity.
//
// @Summary      Health check
// @Description  Reports API status and database connectivity. Returns 503 with db "down" when the database ping fails.
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), constants.HealthPingTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(HealthResponse{
			Status: constants.HealthStatusDegraded,
			DB:     constants.DBStatusDown,
		})
	}
	return c.Status(fiber.StatusOK).JSON(HealthResponse{
		Status: constants.HealthStatusOK,
		DB:     constants.DBStatusUp,
	})
}
