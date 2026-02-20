package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetIssue(key string) (*Issue, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s?fields=summary,description,status,assignee,reporter,priority,issuetype,project,created,updated,sprint,comment", url.PathEscape(key))
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse issue: %w", err)
	}
	return &issue, nil
}

func (c *Client) CreateIssue(projectKey, summary, issueType string) (*Issue, error) {
	req := CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: projectKey},
			Summary:   summary,
			IssueType: TypeRef{Name: issueType},
		},
	}
	data, err := c.do("POST", "/rest/api/3/issue", req)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse created issue: %w", err)
	}
	return &issue, nil
}

func (c *Client) UpdateIssue(key string, fields map[string]interface{}) error {
	body := map[string]interface{}{"fields": fields}
	_, err := c.do("PUT", fmt.Sprintf("/rest/api/3/issue/%s", url.PathEscape(key)), body)
	return err
}

func (c *Client) GetTransitions(key string) ([]Transition, error) {
	data, err := c.do("GET", fmt.Sprintf("/rest/api/3/issue/%s/transitions", url.PathEscape(key)), nil)
	if err != nil {
		return nil, err
	}
	var resp TransitionsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse transitions: %w", err)
	}
	return resp.Transitions, nil
}

func (c *Client) TransitionIssue(key, transitionID string) error {
	req := TransitionRequest{Transition: TypeIDRef{ID: transitionID}}
	_, err := c.do("POST", fmt.Sprintf("/rest/api/3/issue/%s/transitions", url.PathEscape(key)), req)
	return err
}

func (c *Client) AddComment(key, text string) error {
	// Jira Cloud v3 requires Atlassian Document Format (ADF) for comment bodies
	body := map[string]interface{}{
		"body": map[string]interface{}{
			"version": 1,
			"type":    "doc",
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": text},
					},
				},
			},
		},
	}
	_, err := c.do("POST", fmt.Sprintf("/rest/api/3/issue/%s/comment", url.PathEscape(key)), body)
	return err
}

func (c *Client) AssignIssue(key, accountID string) error {
	body := map[string]string{"accountId": accountID}
	_, err := c.do("PUT", fmt.Sprintf("/rest/api/3/issue/%s/assignee", url.PathEscape(key)), body)
	return err
}

// LogWork adds a worklog entry to an issue.
// timeSpent is a Jira duration string like "2h", "30m", "1d".
func (c *Client) LogWork(key, timeSpent string) error {
	body := map[string]string{"timeSpent": timeSpent}
	_, err := c.do("POST", fmt.Sprintf("/rest/api/3/issue/%s/worklog", url.PathEscape(key)), body)
	return err
}

// CreateIssueWithDetails creates an issue with full field support including priority and description.
func (c *Client) CreateIssueWithDetails(projectKey, summary, issueType, priority, description string) (*Issue, error) {
	fields := map[string]interface{}{
		"project":   map[string]string{"key": projectKey},
		"summary":   summary,
		"issuetype": map[string]string{"name": issueType},
	}
	if priority != "" {
		fields["priority"] = map[string]string{"name": priority}
	}
	if description != "" {
		// Jira Cloud v3 requires ADF for description
		fields["description"] = map[string]interface{}{
			"version": 1,
			"type":    "doc",
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": description},
					},
				},
			},
		}
	}

	body := map[string]interface{}{"fields": fields}
	data, err := c.do("POST", "/rest/api/3/issue", body)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse created issue: %w", err)
	}
	return &issue, nil
}

// MoveToSprint moves an issue into a sprint using the Agile API.
func (c *Client) MoveToSprint(sprintID int, issueKeys ...string) error {
	body := map[string]interface{}{
		"issues": issueKeys,
	}
	_, err := c.do("POST", fmt.Sprintf("/rest/agile/1.0/sprint/%d/issue", sprintID), body)
	return err
}
