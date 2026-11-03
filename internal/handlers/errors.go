package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/services"
)

// ErrorHandler Convert domain and HTTP errors into consistent JSON responses.
func ErrorHandler(c fiber.Ctx, err error) error {
	status := 500
	message := "Внутренняя ошибка"

	var domain *services.Error
	var http *fiber.Error

	if errors.As(err, &domain) {
		status = domain.Status
		message = domain.Message
	} else if errors.As(err, &http) {
		status = http.Code
		message = http.Message
	}

	if status == 500 {
		slog.Error("request failed", "error", err)
	}
	return c.Status(status).JSON(fiber.Map{"error": message})
}
