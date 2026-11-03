package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"github.com/phosphoriat/invest-pool-go/internal/models"
)

// Serialize the complete read-check-write transaction, including readers. One service instance
type ProblemRepository struct {
	DB *sql.DB
	mu sync.Mutex
}

// NewProblemRepository Create a new ProblemRepository instance
func NewProblemRepository(db *sql.DB) *ProblemRepository {
	return &ProblemRepository{DB: db}
}

// Change Execute an aggregate mutation and save its outbox event in one transaction.
// Serializes read-check-write operations through the shared repository lock.
func (r *ProblemRepository) Change(ctx context.Context, id string, fn func(*models.Problem) (*models.Event, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	p := models.Problem{ID: id, Solutions: []models.Solution{}}
	var raw string
	err = tx.QueryRowContext(ctx, "SELECT document FROM problems WHERE id=?", id).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		if err = json.Unmarshal([]byte(raw), &p); err != nil {
			return err
		}
	}

	event, err := fn(&p)
	if err != nil {
		return err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, "INSERT INTO problems(id,document) VALUES(?,?) ON CONFLICT(id) DO UPDATE SET document=excluded.document", id, string(data)); err != nil {
		return err
	}

	if event != nil {
		data, err = json.Marshal(event)

		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO outbox(id,document) VALUES(?,?)", event.ID, string(data)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Problems Read fresh problem aggregates from SQLite in creation order
func (r *ProblemRepository) Problems(ctx context.Context) ([]models.Problem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows, err := r.DB.QueryContext(ctx, "SELECT document FROM problems ORDER BY rowid")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result := []models.Problem{}

	for rows.Next() {
		var raw string
		var p models.Problem

		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, err
		}

		result = append(result, p)
	}

	return result, rows.Err()
}
