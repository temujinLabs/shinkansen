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
func Sync(client *jira.Client, store *Store) SyncResult {
	start := time.Now()

	// Build JQL with delta if we have a previous sync
	jql := "assignee = currentUser() AND resolution = Unresolved ORDER BY updated DESC"
	lastSync, err := store.LastSync()
	if err == nil && !lastSync.IsZero() {
		// Delta sync — only fetch issues updated since last sync
		since := lastSync.Format("2006-01-02 15:04")
		jql = fmt.Sprintf("assignee = currentUser() AND updated >= '%s' ORDER BY updated DESC", since)
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
