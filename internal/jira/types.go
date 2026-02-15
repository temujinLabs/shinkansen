package jira

import "time"

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
	ID      string `json:"id"`
	Author  User   `json:"author"`
	Body    string `json:"body"`
	Created string `json:"created"`
}

type IssueFields struct {
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
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
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
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

type AddCommentRequest struct {
	Body string `json:"body"`
}

type TransitionRequest struct {
	Transition TypeIDRef `json:"transition"`
}

type TypeIDRef struct {
	ID string `json:"id"`
}
