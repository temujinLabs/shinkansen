package jira

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetSprints(boardID int) ([]Sprint, error) {
	path := fmt.Sprintf("/rest/agile/1.0/board/%d/sprint?state=active,future", boardID)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp SprintsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse sprints: %w", err)
	}
	return resp.Values, nil
}

func (c *Client) GetSprintIssues(sprintID int) ([]Issue, error) {
	path := fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue?fields=summary,status,assignee,priority,issuetype,project,updated", sprintID)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp SearchResult
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse sprint issues: %w", err)
	}
	return resp.Issues, nil
}
