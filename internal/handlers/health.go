package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v3"
)

type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler Create a new HealthHandler instance
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check database connectivity and return service health
// GET /health
func (h *HealthHandler) Check(c fiber.Ctx) error {
	if err := h.db.PingContext(c.Context()); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
