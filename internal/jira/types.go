package jira

import (
	"encoding/json"
	"time"
)

// Jira Cloud REST API v3 response types

type User struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Sprint struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"` // active, closed, future
	BoardID   int    `json:"originBoardId"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type Comment struct {
	ID      string          `json:"id"`
	Author  User            `json:"author"`
	Body    json.RawMessage `json:"body"`
	Created string          `json:"created"`
}

// BodyText extracts plain text from an ADF comment body.
func (c *Comment) BodyText() string {
	// Try ADF format first (v3 API returns this)
	var doc struct {
		Content []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	}
	if err := json.Unmarshal(c.Body, &doc); err == nil {
		var text string
		for _, block := range doc.Content {
			for _, inline := range block.Content {
				text += inline.Text
			}
			text += "\n"
		}
		if text != "" {
			return text[:len(text)-1] // trim trailing newline
		}
	}
	// Fallback: try plain string
	var s string
	if err := json.Unmarshal(c.Body, &s); err == nil {
		return s
	}
	return string(c.Body)
}

type IssueFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description,omitempty"`
	Status      Status    `json:"status"`
	Assignee    *User     `json:"assignee,omitempty"`
	Reporter    *User     `json:"reporter,omitempty"`
	Priority    Priority  `json:"priority"`
	IssueType   IssueType `json:"issuetype"`
	Project     Project   `json:"project"`
	Created     string    `json:"created"`
	Updated     string    `json:"updated"`
	Sprint      *Sprint   `json:"sprint,omitempty"`
	Comment     *struct {
		Comments []Comment `json:"comments"`
	} `json:"comment,omitempty"`
}

type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

// DescriptionText extracts plain text from the ADF description.
func (i *Issue) DescriptionText() string {
	if len(i.Fields.Description) == 0 || string(i.Fields.Description) == "null" {
		return ""
	}
	// Reuse the same ADF parsing as Comment.BodyText
	c := &Comment{Body: i.Fields.Description}
	return c.BodyText()
}

func (i *Issue) AssigneeName() string {
	if i.Fields.Assignee != nil {
		return i.Fields.Assignee.DisplayName
	}
	return "Unassigned"
}

func (i *Issue) UpdatedTime() time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05.000-0700", i.Fields.Updated)
	return t
}

type SearchResult struct {
	// Legacy fields (agile API still returns these)
	StartAt    int `json:"startAt,omitempty"`
	MaxResults int `json:"maxResults,omitempty"`
	Total      int `json:"total,omitempty"`
	// New search/jql endpoint uses token-based pagination
	NextPageToken *string `json:"nextPageToken"`
	IsLast        bool    `json:"isLast,omitempty"`
	// Common
	Issues []Issue `json:"issues"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   Status `json:"to"`
}

type TransitionsResponse struct {
	Transitions []Transition `json:"transitions"`
}

type Board struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // scrum, kanban
}

type BoardsResponse struct {
	Values []Board `json:"values"`
}

type SprintsResponse struct {
	Values []Sprint `json:"values"`
}

type SprintIssuesResponse struct {
	Issues []Issue `json:"issues"`
}

// Request types

type CreateIssueRequest struct {
	Fields CreateIssueFields `json:"fields"`
}

type CreateIssueFields struct {
	Project   ProjectRef `json:"project"`
	Summary   string     `json:"summary"`
	IssueType TypeRef    `json:"issuetype"`
	Priority  *TypeRef   `json:"priority,omitempty"`
	Assignee  *UserRef   `json:"assignee,omitempty"`
}

type ProjectRef struct {
	Key string `json:"key"`
}

type TypeRef struct {
	Name string `json:"name"`
}

type UserRef struct {
	AccountID string `json:"accountId"`
}

type TransitionRequest struct {
	Transition TypeIDRef `json:"transition"`
}

type TypeIDRef struct {
	ID string `json:"id"`
}
