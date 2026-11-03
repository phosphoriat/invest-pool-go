package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/models"
)

// GetUsers Return seeded users for demonstration API clients
// GET /api/users
func GetUsers(c fiber.Ctx) error {
	return c.JSON(models.Users)
}
