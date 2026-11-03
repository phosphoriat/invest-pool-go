package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/phosphoriat/invest-pool-go/internal/models"
	"github.com/phosphoriat/invest-pool-go/internal/repository"
)

type Subscriber func(context.Context, models.Event) error
type Bus struct {
	repo        *repository.EventRepository
	subscribers []Subscriber
	wake        chan struct{}
	mu          sync.Mutex
}

// NewBus Create a new Bus instance
func NewBus(r *repository.EventRepository) *Bus {
	return &Bus{repo: r, wake: make(chan struct{}, 1)}
}

// Subscribe Register an event subscriber during startup, before Run is called
func (b *Bus) Subscribe(s Subscriber) {
	b.subscribers = append(b.subscribers, s)
}

// Notify Wake the dispatcher without blocking the command that committed an event
func (b *Bus) Notify() {
	select {
	case b.wake <- struct{}{}:
	default:
	}
}

// Drain Deliver pending events and acknowledge each only after all subscribers succeed
func (b *Bus) Drain(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	events, err := b.repo.Pending(ctx)
	if err != nil {
		return err
	}

	for _, e := range events {
		for _, s := range b.subscribers {
			if err = s(ctx, e); err != nil {
				return err
			}
		}

		if err = b.repo.Delivered(ctx, e.ID); err != nil {
			return err
		}
	}

	return nil
}

// Run Process persisted outbox events on startup, notifications and periodic retries
func (b *Bus) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if err := b.Drain(ctx); err != nil && ctx.Err() == nil {
			slog.Error("outbox delivery failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-b.wake:
		case <-ticker.C:
		}
	}
}
