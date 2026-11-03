package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/phosphoriat/invest-pool-go/internal/config"
	"github.com/phosphoriat/invest-pool-go/internal/database/sqlite"
	"github.com/phosphoriat/invest-pool-go/internal/handlers"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
	"github.com/phosphoriat/invest-pool-go/internal/routes"
	"github.com/phosphoriat/invest-pool-go/internal/services"
)

// main Start the service and report lifecycle errors
func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

// run Initialize dependencies, serve HTTP requests and release resources on shutdown.
func run() error {
	// Download the settings and open SQLite; the connection closes last
	cfg := config.Load()
	db, err := sqlite.Open(cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	// All services use a common repository of problems and its blocking aggregate
	problems := repository.NewProblemRepository(db)
	events := repository.NewEventRepository(db)

	// Audit subscriber saves events delivered from transactional outbox
	bus := services.NewBus(events)
	bus.Subscribe(events.Audit)

	// Configure HTTP server, panic processing and application routes
	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
		BodyLimit:    16 * 1024,
	})
	app.Use(recover.New())
	routes.SetupRoutes(app, problems, events, bus)

	// Bus running until HTTP requests complete and final delivery attempt
	busCtx, cancelBus := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		bus.Run(busCtx)
	}()
	defer func() {
		cancelBus()
		<-done
	}()

	// Waiting for stop signal or HTTP server startup/operation error
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(cfg.Address)
	}()

	select {
	case err := <-listenErr:
		return err
	case <-ctx.Done():
	}

	// Stop accepting requests and let current operations end
	if err = app.ShutdownWithTimeout(10 * time.Second); err != nil {
		return err
	}

	// Try to deliver the rest of the outbox; undelivered events will survive restart
	drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return bus.Drain(drainCtx)
}
