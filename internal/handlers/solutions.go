package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type SolutionsHandler struct {
	service *services.SolutionsService
}
type ProposeSolutionRequest struct {
	Text string `json:"text"`
}

// NewSolutionsHandler Create a new SolutionsHandler instance
func NewSolutionsHandler(service *services.SolutionsService) *SolutionsHandler {
	return &SolutionsHandler{service: service}
}

// Propose Submit a solution on behalf of the authenticated technologist
// POST /api/problems/:id/solutions
func (h *SolutionsHandler) Propose(c fiber.Ctx) error {
	var request ProposeSolutionRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	id, err := h.service.Propose(c.Context(), middleware.User(c), c.Params("id"), request.Text)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"problem_id": id})
}
