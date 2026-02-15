package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/temujinlabs/shinkansen/internal/jira"
)

type Store struct {
	db *sql.DB
}

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "shinkansen")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache.db"), nil
}

func NewStore() (*Store, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open cache db: %w", err)
	}

	// WAL mode for better concurrent reads
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migrate cache: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS issues (
		key TEXT PRIMARY KEY,
		summary TEXT,
		status TEXT,
		assignee TEXT,
		priority TEXT,
		issue_type TEXT,
		project_key TEXT,
		sprint_id INTEGER,
		updated_at TEXT,
		raw_json TEXT
	);

	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		key TEXT,
		name TEXT
	);

	CREATE TABLE IF NOT EXISTS sprints (
		id INTEGER PRIMARY KEY,
		name TEXT,
		state TEXT,
		board_id INTEGER,
		start_date TEXT,
		end_date TEXT
	);

	CREATE TABLE IF NOT EXISTS transitions (
		issue_key TEXT,
		transition_id TEXT,
		name TEXT,
		PRIMARY KEY (issue_key, transition_id)
	);

	CREATE TABLE IF NOT EXISTS sync_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		last_sync TEXT,
		items_synced INTEGER,
		duration_ms INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
	CREATE INDEX IF NOT EXISTS idx_issues_project ON issues(project_key);
	CREATE INDEX IF NOT EXISTS idx_issues_assignee ON issues(assignee);
	`
	_, err := s.db.Exec(schema)
	return err
}

// UpsertIssue stores or updates an issue in the cache.
func (s *Store) UpsertIssue(issue *jira.Issue) error {
	raw, _ := json.Marshal(issue)
	assignee := ""
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}
	sprintID := 0
	if issue.Fields.Sprint != nil {
		sprintID = issue.Fields.Sprint.ID
	}

	_, err := s.db.Exec(`
		INSERT INTO issues (key, summary, status, assignee, priority, issue_type, project_key, sprint_id, updated_at, raw_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			summary=excluded.summary, status=excluded.status, assignee=excluded.assignee,
			priority=excluded.priority, issue_type=excluded.issue_type, project_key=excluded.project_key,
			sprint_id=excluded.sprint_id, updated_at=excluded.updated_at, raw_json=excluded.raw_json`,
		issue.Key, issue.Fields.Summary, issue.Fields.Status.Name,
		assignee, issue.Fields.Priority.Name, issue.Fields.IssueType.Name,
		issue.Fields.Project.Key, sprintID, issue.Fields.Updated, string(raw),
	)
	return err
}

// GetIssues returns cached issues, optionally filtered by status.
func (s *Store) GetIssues(status string) ([]jira.Issue, error) {
	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = s.db.Query("SELECT raw_json FROM issues WHERE status = ? ORDER BY priority, updated_at DESC", status)
	} else {
		rows, err = s.db.Query("SELECT raw_json FROM issues ORDER BY priority, updated_at DESC")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []jira.Issue
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		var issue jira.Issue
		if err := json.Unmarshal([]byte(raw), &issue); err != nil {
			continue
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// GetAllIssues returns all cached issues.
func (s *Store) GetAllIssues() ([]jira.Issue, error) {
	return s.GetIssues("")
}

// GetIssue returns a single cached issue by key.
func (s *Store) GetIssue(key string) (*jira.Issue, error) {
	var raw string
	err := s.db.QueryRow("SELECT raw_json FROM issues WHERE key = ?", key).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var issue jira.Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// SearchIssues returns issues matching a text query (searches key and summary).
func (s *Store) SearchIssues(query string) ([]jira.Issue, error) {
	like := "%" + query + "%"
	rows, err := s.db.Query(
		"SELECT raw_json FROM issues WHERE key LIKE ? OR summary LIKE ? ORDER BY updated_at DESC LIMIT 50",
		like, like,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []jira.Issue
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		var issue jira.Issue
		if err := json.Unmarshal([]byte(raw), &issue); err != nil {
			continue
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// UpsertTransitions stores transitions for an issue.
func (s *Store) UpsertTransitions(issueKey string, transitions []jira.Transition) error {
	s.db.Exec("DELETE FROM transitions WHERE issue_key = ?", issueKey)
	for _, t := range transitions {
		s.db.Exec(
			"INSERT OR REPLACE INTO transitions (issue_key, transition_id, name) VALUES (?, ?, ?)",
			issueKey, t.ID, t.Name,
		)
	}
	return nil
}

// GetTransitions returns cached transitions for an issue.
func (s *Store) GetTransitions(issueKey string) ([]jira.Transition, error) {
	rows, err := s.db.Query("SELECT transition_id, name FROM transitions WHERE issue_key = ?", issueKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []jira.Transition
	for rows.Next() {
		var t jira.Transition
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			continue
		}
		transitions = append(transitions, t)
	}
	return transitions, nil
}

// LastSync returns the time of the last successful sync.
func (s *Store) LastSync() (time.Time, error) {
	var ts string
	err := s.db.QueryRow("SELECT last_sync FROM sync_log ORDER BY id DESC LIMIT 1").Scan(&ts)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, ts)
}

// RecordSync logs a completed sync.
func (s *Store) RecordSync(itemsSynced int, duration time.Duration) error {
	_, err := s.db.Exec(
		"INSERT INTO sync_log (last_sync, items_synced, duration_ms) VALUES (?, ?, ?)",
		time.Now().UTC().Format(time.RFC3339), itemsSynced, duration.Milliseconds(),
	)
	return err
}
