package jira

type Issue struct {
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	IssueType   IssueType     `json:"issuetype"`
	Parent      *Issue        `json:"parent"`
	Project     Project       `json:"project"`
	Priority    Priority      `json:"priority"`
	Labels      []interface{} `json:"labels"`
	IssueLinks  []IssueLink   `json:"issuelinks"`
	Assignee    interface{}   `json:"assignee"`
	Status      Status        `json:"status"`
	Description interface{}   `json:"description"`
	Summary     string        `json:"summary"`
	Creator     Creator       `json:"creator"`
	Subtasks    []interface{} `json:"subtasks"`
	Duedate     interface{}   `json:"duedate"`
	Progress    Progress      `json:"progress"`

	// HACK. Sigh.
	// I think I never found something as badly designed as Jira.
	// Tag 'remain' tells 'mapstructure' to collect all unknown fields, at any
	// level of nesting.
	CustomFields map[string]any `mapstructure:",remain"`
}

type Progress struct {
	Progress int `json:"progress"`
	Total    int `json:"total"`
}

type Creator struct {
	DisplayName string `json:"displayName"`
}

type Priority struct {
	Name string `json:"name"`
}

type IssueLink struct {
	OutwardIssue Issue    `json:"outwardIssue"`
	InwardIssue  Issue    `json:"inwardIssue"`
	Type         LinkType `json:"type"`
}

// NOTE The 3 fields below are always the same, for example:
//
//	Name: Blocks
//	Inward: "is blocked by"
//	Outward: "blocks"
//
// so there is no direction information! The direction is determined by
// which one of OutwardIssue or InwardIssue is filled :-/
type LinkType struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// And many others...
const LinkTypeBlocks = "Blocks"

type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type IssueType struct {
	Name    string `json:"name"`
	Subtask bool   `json:"subtask"`
}

type Status struct {
	Name string `json:"name"`
}
