package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type MatchesHandler struct {
	service *services.MatchesService
}
type ConfirmMatchRequest struct {
	SolutionID string `json:"solution_id"`
}

// NewMatchesHandler Create a new MatchesHandler instance
func NewMatchesHandler(service *services.MatchesService) *MatchesHandler {
	return &MatchesHandler{service: service}
}

// Confirm Confirm the problem-solution match on behalf of the expert
// POST /api/problems/:id/match
func (h *MatchesHandler) Confirm(c fiber.Ctx) error {
	var request ConfirmMatchRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	id, err := h.service.Confirm(c.Context(), middleware.User(c), c.Params("id"), request.SolutionID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"problem_id": id})
}
