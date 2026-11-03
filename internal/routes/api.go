package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/phosphoriat/invest-pool-go/internal/handlers"
	"github.com/phosphoriat/invest-pool-go/internal/middleware"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

// SetupRoutes Create domain services and handlers, then register REST API routes.
// All services share one ProblemRepository to preserve aggregate locking.
func SetupRoutes(app *fiber.App, problems *repository.ProblemRepository, events *repository.EventRepository, bus *services.Bus) {
	// Init services
	problemsService := services.NewProblemsService(problems, bus)
	solutionsService := services.NewSolutionsService(problems, bus)
	matchesService := services.NewMatchesService(problems, bus)
	poolsService := services.NewPoolsService(problems, bus)
	investmentsService := services.NewInvestmentsService(problems, bus)
	distributionsService := services.NewDistributionsService(problems, bus)
	eventsService := services.NewEventsService(events, poolsService)

	// Init handlers
	problemsHandler := handlers.NewProblemsHandler(problemsService)
	solutionsHandler := handlers.NewSolutionsHandler(solutionsService)
	matchesHandler := handlers.NewMatchesHandler(matchesService)
	poolsHandler := handlers.NewPoolsHandler(poolsService)
	investmentsHandler := handlers.NewInvestmentsHandler(investmentsService)
	distributionsHandler := handlers.NewDistributionsHandler(distributionsService)
	eventsHandler := handlers.NewEventsHandler(eventsService)
	healthHandler := handlers.NewHealthHandler(problems.DB)

	// Health check route
	app.Get("/health", healthHandler.Check)
	// Users public route
	app.Get("/api/users", handlers.GetUsers)

	// Set api routes group (with auth middleware)
	api := app.Group("/api", middleware.Auth)
	api.Get("/state", poolsHandler.State)
	api.Get("/events", eventsHandler.Audit)

	// Problem routes group
	problemRoutes := api.Group("/problems")
	problemRoutes.Post("", problemsHandler.Publish)
	problemRoutes.Post("/:id/solutions", solutionsHandler.Propose)
	problemRoutes.Post("/:id/match", matchesHandler.Confirm)
	problemRoutes.Post("/:id/pool", poolsHandler.Create)
	problemRoutes.Post("/:id/investments", investmentsHandler.Invest)
	problemRoutes.Post("/:id/profit", distributionsHandler.Distribute)
}
