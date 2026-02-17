package cache

import (
	"fmt"
	"time"

	"github.com/temujinlabs/shinkansen/internal/jira"
)

// SyncResult holds the outcome of a sync operation.
type SyncResult struct {
	ItemsSynced int
	Duration    time.Duration
	Err         error
}

// Sync fetches updated issues from Jira and caches them.
// Uses delta sync: only fetches issues updated since last sync.
// projectKey scopes results to a specific project (e.g. "SCRUM").
func Sync(client *jira.Client, store *Store, projectKey string) SyncResult {
	start := time.Now()

	// Build JQL scoped to the configured project.
	// Include unresolved issues + recently resolved (last 14 days) for Done column.
	base := "(resolution = Unresolved OR resolutiondate >= -14d)"
	if projectKey != "" {
		base = fmt.Sprintf("project = %s AND (resolution = Unresolved OR resolutiondate >= -14d)", projectKey)
	}
	jql := base + " ORDER BY updated DESC"

	lastSync, err := store.LastSync()
	if err == nil && !lastSync.IsZero() {
		since := lastSync.Format("2006-01-02 15:04")
		jql = fmt.Sprintf("%s AND updated >= '%s' ORDER BY updated DESC", base, since)
	}

	issues, err := client.SearchAll(jql)
	if err != nil {
		return SyncResult{Err: fmt.Errorf("search: %w", err)}
	}

	synced := 0
	for i := range issues {
		if err := store.UpsertIssue(&issues[i]); err != nil {
			continue
		}
		synced++

		// Also cache transitions for each issue
		transitions, err := client.GetTransitions(issues[i].Key)
		if err == nil {
			store.UpsertTransitions(issues[i].Key, transitions)
		}
	}

	duration := time.Since(start)
	store.RecordSync(synced, duration)

	return SyncResult{
		ItemsSynced: synced,
		Duration:    duration,
	}
}
