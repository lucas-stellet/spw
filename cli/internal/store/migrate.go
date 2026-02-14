package store

import (
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQLv1 string

// Migrate runs schema migrations based on PRAGMA user_version.
// It is idempotent and safe to call multiple times.
func (s *SpecStore) Migrate() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("store: read user_version: %w", err)
	}

	if version >= 1 {
		return nil // already migrated
	}

	if _, err := s.db.Exec(schemaSQLv1); err != nil {
		return fmt.Errorf("store: apply schema v1: %w", err)
	}

	if _, err := s.db.Exec("PRAGMA user_version = 1"); err != nil {
		return fmt.Errorf("store: set user_version: %w", err)
	}

	return nil
}
