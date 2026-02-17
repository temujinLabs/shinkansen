package jira

import (
	"encoding/json"
	"fmt"
)

// Search calls POST /rest/api/3/search/jql (the new endpoint).
// Pagination uses nextPageToken, not startAt.
func (c *Client) Search(jql string, maxResults int, nextPageToken string) (*SearchResult, error) {
	body := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"fields":     []string{"summary", "status", "assignee", "priority", "issuetype", "project", "updated", "sprint", "comment", "description", "reporter", "created"},
	}
	if nextPageToken != "" {
		body["nextPageToken"] = nextPageToken
	}
	data, err := c.do("POST", "/rest/api/3/search/jql", body)
	if err != nil {
		return nil, err
	}
	var result SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	return &result, nil
}

// SearchAll pages through all results for a JQL query using token-based pagination.
func (c *Client) SearchAll(jql string) ([]Issue, error) {
	var all []Issue
	nextToken := ""

	for {
		result, err := c.Search(jql, 50, nextToken)
		if err != nil {
			return all, err
		}
		all = append(all, result.Issues...)
		if result.IsLast || result.NextPageToken == nil || *result.NextPageToken == "" {
			break
		}
		nextToken = *result.NextPageToken
	}
	return all, nil
}

// MyIssues returns issues assigned to the current user.
func (c *Client) MyIssues() ([]Issue, error) {
	return c.SearchAll("assignee = currentUser() AND resolution = Unresolved ORDER BY priority ASC, updated DESC")
}
