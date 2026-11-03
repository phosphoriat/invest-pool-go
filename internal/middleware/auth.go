package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/models"
)

// Auth Verify the seeded user ID and role before allowing access to protected routes
func Auth(c fiber.Ctx) error {
	for _, u := range models.Users {
		if c.Get("X-User-Id") == u.ID {
			if c.Get("X-Role") != u.Role {
				return fiber.NewError(403, "Роль не соответствует пользователю")
			}

			c.Locals("user", u)
			return c.Next()
		}
	}

	return fiber.NewError(401, "Неизвестный пользователь")
}

// User Return the authenticated user previously stored by Auth middleware
func User(c fiber.Ctx) models.User {
	return c.Locals("user").(models.User)
}
