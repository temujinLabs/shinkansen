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

func (c *Client) AddComment(key, body string) error {
	req := AddCommentRequest{Body: body}
	_, err := c.do("POST", fmt.Sprintf("/rest/api/3/issue/%s/comment", url.PathEscape(key)), req)
	return err
}

func (c *Client) AssignIssue(key, accountID string) error {
	body := map[string]string{"accountId": accountID}
	_, err := c.do("PUT", fmt.Sprintf("/rest/api/3/issue/%s/assignee", url.PathEscape(key)), body)
	return err
}
