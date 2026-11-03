package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

type ProblemsHandler struct {
	service *services.ProblemsService
}
type PublishProblemRequest struct {
	Text string `json:"text"`
}

// NewProblemsHandler Create a new ProblemsHandler instance
func NewProblemsHandler(service *services.ProblemsService) *ProblemsHandler {
	return &ProblemsHandler{service: service}
}

// Publish a problem on behalf of the authenticated region
// POST /api/problems
func (h *ProblemsHandler) Publish(c fiber.Ctx) error {
	var request PublishProblemRequest
	if err := middleware.DecodeJSON(c, &request); err != nil {
		return err
	}

	id, err := h.service.Publish(c.Context(), middleware.User(c), request.Text)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"problem_id": id})
}
