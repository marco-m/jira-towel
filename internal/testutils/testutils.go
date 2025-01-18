package testutils

import (
	"github.com/marco-m/jira-towel/pkg/jira"
)

type BuildArg struct {
	Key       string
	Summary   string
	Blocks    []string
	BlockedBy []string
}

// Create a jira.Issue with only a minimal number of fields set, enough to
// express the required dependencies. Useful to create test cases, since writing
// a literal jira.Issue is so verbose that the dependencies are lost into the
// verbosity.
func BuildIssue(arg BuildArg) jira.Issue {
	issue := jira.Issue{
		Key: arg.Key,
		Fields: jira.Fields{
			Summary: arg.Summary,
		},
	}
	if len(arg.Blocks) > 0 {
		links := make([]jira.IssueLink, 0, len(arg.Blocks))
		for _, blocked := range arg.Blocks {
			links = append(links, jira.IssueLink{
				Type:         jira.LinkType{Name: jira.LinkTypeBlocks},
				OutwardIssue: jira.Issue{Key: blocked},
			})
		}
		issue.Fields.IssueLinks = append(issue.Fields.IssueLinks, links...)
	}
	if len(arg.BlockedBy) > 0 {
		links := make([]jira.IssueLink, 0, len(arg.BlockedBy))
		for _, blocking := range arg.BlockedBy {
			links = append(links, jira.IssueLink{
				Type:        jira.LinkType{Name: jira.LinkTypeBlocks},
				InwardIssue: jira.Issue{Key: blocking},
			})
		}
		issue.Fields.IssueLinks = append(issue.Fields.IssueLinks, links...)
	}
	return issue
}
