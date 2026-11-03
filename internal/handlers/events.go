package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type EventsHandler struct {
	service *services.EventsService
}

// NewEventsHandler Create a new EventsHandler instance
func NewEventsHandler(service *services.EventsService) *EventsHandler {
	return &EventsHandler{service: service}
}

// Audit Return the authorized user's persisted audit events
// GET /api/events
func (h *EventsHandler) Audit(c fiber.Ctx) error {
	events, err := h.service.Audit(c.Context(), middleware.User(c))
	if err != nil {
		return err
	}

	return c.JSON(events)
}
