-- spec.db schema v1
-- Per-spec SQLite database for spec-workflow runtime state and artifacts.

CREATE TABLE IF NOT EXISTS spec_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artifacts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    phase         TEXT NOT NULL,
    rel_path      TEXT UNIQUE NOT NULL,
    artifact_type TEXT NOT NULL,
    content       TEXT NOT NULL,
    content_hash  TEXT NOT NULL,
    metadata      TEXT,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    command     TEXT NOT NULL,
    run_number  INTEGER NOT NULL,
    phase       TEXT NOT NULL,
    wave_number INTEGER,
    comms_path  TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'in_progress',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(command, run_number, wave_number)
);

CREATE TABLE IF NOT EXISTS subagents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id      INTEGER NOT NULL REFERENCES runs(id),
    name        TEXT NOT NULL,
    brief       TEXT,
    report      TEXT,
    status      TEXT,
    summary     TEXT,
    status_json TEXT,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS waves (
    wave_number    INTEGER PRIMARY KEY,
    status         TEXT NOT NULL DEFAULT 'pending',
    exec_runs      INTEGER DEFAULT 0,
    check_runs     INTEGER DEFAULT 0,
    summary_status TEXT,
    summary_text   TEXT,
    summary_source TEXT,
    stale_flag     INTEGER DEFAULT 0,
    updated_at     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
    task_id     TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    wave        INTEGER,
    depends_on  TEXT,
    files       TEXT,
    tdd         INTEGER DEFAULT 0,
    is_deferred INTEGER DEFAULT 0,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS impl_logs (
    task_id      TEXT PRIMARY KEY,
    content      TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS handoffs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     INTEGER NOT NULL REFERENCES runs(id),
    content    TEXT NOT NULL,
    all_pass   INTEGER NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS approvals (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_type    TEXT NOT NULL,
    approval_id TEXT NOT NULL,
    raw_json    TEXT,
    created_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS completion_summary (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    frontmatter  TEXT NOT NULL,
    body         TEXT NOT NULL,
    generated_at TEXT NOT NULL
);

-- Indexes for frequently queried columns.
CREATE INDEX IF NOT EXISTS idx_artifacts_phase ON artifacts(phase);
CREATE INDEX IF NOT EXISTS idx_artifacts_type  ON artifacts(artifact_type);
CREATE INDEX IF NOT EXISTS idx_runs_command    ON runs(command);
CREATE INDEX IF NOT EXISTS idx_subagents_run_id ON subagents(run_id);
