package store

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// SpecStore wraps a per-spec SQLite database.
type SpecStore struct {
	db      *sql.DB
	specDir string
	name    string
}

// Open opens (or creates) the spec.db file inside specDir.
// It sets WAL journal mode, busy_timeout=5000ms, and enables foreign keys.
func Open(specDir string) (*SpecStore, error) {
	dbPath := filepath.Join(specDir, "spec.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", dbPath, err)
	}

	// Set pragmas for performance and correctness.
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("store: pragma %q: %w", p, err)
		}
	}

	s := &SpecStore{
		db:      db,
		specDir: specDir,
		name:    filepath.Base(specDir),
	}

	if err := s.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}

	return s, nil
}

// TryOpen returns a SpecStore or nil on error (fail-open helper).
func TryOpen(specDir string) *SpecStore {
	s, err := Open(specDir)
	if err != nil {
		return nil
	}
	return s
}

// Close closes the underlying database connection.
func (s *SpecStore) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for advanced queries.
func (s *SpecStore) DB() *sql.DB {
	return s.db
}
