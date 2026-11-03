package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/phosphoriat/invest-pool-go/internal/models"
)

type EventRepository struct {
	DB *sql.DB
}

// NewEventRepository Create a new EventRepository instance
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{DB: db}
}

// Pending Read undelivered outbox events in insertion order
func (r *EventRepository) Pending(ctx context.Context) ([]models.Event, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT document FROM outbox WHERE delivered=0 ORDER BY rowid")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	events := []models.Event{}
	for rows.Next() {
		var raw string
		var e models.Event
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

// Audit Save an audit event once, deduplicating by event ID
func (r *EventRepository) Audit(ctx context.Context, e models.Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx, "INSERT OR IGNORE INTO audit(event_id,document) VALUES(?,?)", e.ID, string(b))

	return err
}

// Delivered Mark an outbox event as delivered after all subscribers succeed
func (r *EventRepository) Delivered(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE outbox SET delivered=1 WHERE id=?", id)
	return err
}

// Events Read the persisted audit log in insertion order
func (r *EventRepository) Events(ctx context.Context) ([]models.Event, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT document FROM audit ORDER BY rowid")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	out := []models.Event{}

	for rows.Next() {
		var raw string
		var e models.Event

		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &e); err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, rows.Err()
}
