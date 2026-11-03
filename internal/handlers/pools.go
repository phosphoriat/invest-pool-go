package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type PoolsHandler struct {
	service *services.PoolsService
}
type CreatePoolRequest struct {
	Amount int64 `json:"amount"`
}

// NewPoolsHandler Create a new PoolsHandler instance
func NewPoolsHandler(service *services.PoolsService) *PoolsHandler {
	return &PoolsHandler{service: service}
}

// Create Create an investment pool for a confirmed match
// POST /api/problems/:id/pool
func (h *PoolsHandler) Create(c fiber.Ctx) error {
	var request CreatePoolRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	res, err := h.service.Create(c.Context(), middleware.User(c), c.Params("id"), request.Amount)
	if err != nil {
		return err
	}

	return c.JSON(res)
}

// State Return projects and opportunities visible to the current user
// GET /api/state
func (h *PoolsHandler) State(c fiber.Ctx) error {
	p, o, err := h.service.View(c.Context(), middleware.User(c))
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"projects": p, "opportunities": o})
}
