package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type DistributionsHandler struct {
	service *services.DistributionsService
}
type DistributeProfitRequest struct {
	Amount int64  `json:"amount"`
	Key    string `json:"key"`
}

// NewDistributionsHandler Create a new DistributionsHandler instance
func NewDistributionsHandler(service *services.DistributionsService) *DistributionsHandler {
	return &DistributionsHandler{service: service}
}

// Distribute Record profit and let the service allocate investor payouts automatically
// POST /api/problems/:id/profit
func (h *DistributionsHandler) Distribute(c fiber.Ctx) error {
	var request DistributeProfitRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	id, err := h.service.Distribute(c.Context(), middleware.User(c), c.Params("id"), request.Amount, request.Key)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"problem_id": id})
}
