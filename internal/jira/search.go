package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) Search(jql string, startAt, maxResults int) (*SearchResult, error) {
	params := url.Values{
		"jql":        {jql},
		"startAt":    {fmt.Sprintf("%d", startAt)},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
		"fields":     {"summary,status,assignee,priority,issuetype,project,updated,sprint"},
	}
	path := "/rest/api/3/search?" + params.Encode()
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var result SearchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse search: %w", err)
	}
	return &result, nil
}

// SearchAll pages through all results for a JQL query.
func (c *Client) SearchAll(jql string) ([]Issue, error) {
	var all []Issue
	startAt := 0
	pageSize := 50

	for {
		result, err := c.Search(jql, startAt, pageSize)
		if err != nil {
			return all, err
		}
		all = append(all, result.Issues...)
		if startAt+len(result.Issues) >= result.Total {
			break
		}
		startAt += len(result.Issues)
	}
	return all, nil
}

// MyIssues returns issues assigned to the current user.
func (c *Client) MyIssues() ([]Issue, error) {
	return c.SearchAll("assignee = currentUser() AND resolution = Unresolved ORDER BY priority ASC, updated DESC")
}
