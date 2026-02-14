package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// IndexStore wraps the global cross-spec index database (.spw-index.db).
type IndexStore struct {
	db *sql.DB
}

const indexSchema = `
CREATE TABLE IF NOT EXISTS specs (
	name      TEXT PRIMARY KEY,
	stage     TEXT NOT NULL,
	db_path   TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	spec     TEXT NOT NULL,
	doc_type TEXT NOT NULL,
	phase    TEXT NOT NULL,
	title    TEXT NOT NULL,
	snippet  TEXT,
	content  TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
	title, content, spec, doc_type,
	content=documents, content_rowid=id
);

-- Triggers to keep FTS5 in sync with documents table.
CREATE TRIGGER IF NOT EXISTS documents_ai AFTER INSERT ON documents BEGIN
	INSERT INTO documents_fts(rowid, title, content, spec, doc_type)
	VALUES (new.id, new.title, new.content, new.spec, new.doc_type);
END;

CREATE TRIGGER IF NOT EXISTS documents_ad AFTER DELETE ON documents BEGIN
	INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, doc_type)
	VALUES ('delete', old.id, old.title, old.content, old.spec, old.doc_type);
END;

CREATE TRIGGER IF NOT EXISTS documents_au AFTER UPDATE ON documents BEGIN
	INSERT INTO documents_fts(documents_fts, rowid, title, content, spec, doc_type)
	VALUES ('delete', old.id, old.title, old.content, old.spec, old.doc_type);
	INSERT INTO documents_fts(rowid, title, content, spec, doc_type)
	VALUES (new.id, new.title, new.content, new.spec, new.doc_type);
END;
`

// OpenIndex opens (or creates) the global index database at .spec-workflow/.spw-index.db.
func OpenIndex(workspaceRoot string) (*IndexStore, error) {
	dbPath := filepath.Join(workspaceRoot, ".spec-workflow", ".spw-index.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store: open index %s: %w", dbPath, err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("store: index pragma %q: %w", p, err)
		}
	}

	if _, err := db.Exec(indexSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: index schema: %w", err)
	}

	return &IndexStore{db: db}, nil
}

// IndexSpec registers or updates a spec in the global index.
func (ix *IndexStore) IndexSpec(name, stage, dbPath string) error {
	ts := now()
	_, err := ix.db.Exec(`
		INSERT INTO specs (name, stage, db_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			stage = excluded.stage,
			db_path = excluded.db_path,
			updated_at = excluded.updated_at`,
		name, stage, dbPath, ts, ts,
	)
	if err != nil {
		return fmt.Errorf("store: index spec: %w", err)
	}
	return nil
}

// IndexDocument adds a document to the global index for FTS5 search.
func (ix *IndexStore) IndexDocument(spec, docType, phase, title, snippet, content string) error {
	_, err := ix.db.Exec(
		"INSERT INTO documents (spec, doc_type, phase, title, snippet, content, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		spec, docType, phase, title, snippet, content, now(),
	)
	if err != nil {
		return fmt.Errorf("store: index document: %w", err)
	}
	return nil
}

// Search performs an FTS5 full-text search across indexed documents.
// If specFilter is non-empty, results are limited to that spec.
func (ix *IndexStore) Search(query, specFilter string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	// Quote each token to avoid FTS5 syntax errors with special characters.
	ftsQuery := quoteFTS5(query)

	var (
		rows *sql.Rows
		err  error
	)

	if specFilter != "" {
		rows, err = ix.db.Query(`
			SELECT d.spec, d.doc_type, d.phase, d.title, d.snippet, rank
			FROM documents_fts f
			JOIN documents d ON d.id = f.rowid
			WHERE documents_fts MATCH ? AND d.spec = ?
			ORDER BY rank
			LIMIT ?`,
			ftsQuery, specFilter, limit,
		)
	} else {
		rows, err = ix.db.Query(`
			SELECT d.spec, d.doc_type, d.phase, d.title, d.snippet, rank
			FROM documents_fts f
			JOIN documents d ON d.id = f.rowid
			WHERE documents_fts MATCH ?
			ORDER BY rank
			LIMIT ?`,
			ftsQuery, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("store: search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var snippet sql.NullString
		if err := rows.Scan(&r.Spec, &r.DocType, &r.Phase, &r.Title, &snippet, &r.Rank); err != nil {
			return nil, fmt.Errorf("store: scan search result: %w", err)
		}
		r.Snippet = snippet.String
		results = append(results, r)
	}
	return results, rows.Err()
}

// Close closes the index database connection.
func (ix *IndexStore) Close() error {
	return ix.db.Close()
}

// quoteFTS5 wraps each token in double quotes to safely pass through FTS5 MATCH.
// This prevents special characters (hyphens, colons) from being interpreted as operators.
func quoteFTS5(query string) string {
	return `"` + strings.ReplaceAll(query, `"`, `""`) + `"`
}
