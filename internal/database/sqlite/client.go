package sqlite

import (
	"database/sql"

	"github.com/phosphoriat/invest-pool-go/migrations"
	_ "modernc.org/sqlite"
)

// Open SQLite, configure the connection and apply the embedded schema.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	for _, q := range []string{"PRAGMA busy_timeout=5000", "PRAGMA journal_mode=WAL", migrations.Schema} {
		if _, err = db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}
