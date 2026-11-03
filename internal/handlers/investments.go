package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type InvestmentsHandler struct {
	service *services.InvestmentsService
}
type InvestRequest struct {
	Amount int64  `json:"amount"`
	Key    string `json:"key"`
}

// NewInvestmentsHandler Create a new InvestmentsHandler instance
func NewInvestmentsHandler(service *services.InvestmentsService) *InvestmentsHandler {
	return &InvestmentsHandler{service: service}
}

// Invest Submit an investor contribution with its idempotency key
// POST /api/problems/:id/investments
func (h *InvestmentsHandler) Invest(c fiber.Ctx) error {
	var request InvestRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	res, err := h.service.Invest(c.Context(), middleware.User(c), c.Params("id"), request.Amount, request.Key)
	if err != nil {
		return err
	}

	return c.JSON(res)
}
