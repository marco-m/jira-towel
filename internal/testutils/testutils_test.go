package testutils_test

import (
	"testing"

	"github.com/marco-m/jira-towel/internal/testutils"
	"github.com/marco-m/jira-towel/pkg/jira"
	"github.com/marco-m/rosina"
)

func TestBuildIssue(t *testing.T) {
	have := testutils.BuildIssue(
		testutils.BuildArg{
			Key:       "FOO-7",
			Summary:   "banana",
			BlockedBy: []string{"FOO-71", "FOO-72"},
			Blocks:    []string{"FOO-31", "FOO-32"},
		},
	)
	want := jira.Issue{
		Key: "FOO-7",
		Fields: jira.Fields{
			Summary: "banana",
			IssueLinks: []jira.IssueLink{
				{
					OutwardIssue: jira.Issue{Key: "FOO-31"},
					Type:         jira.LinkType{Name: "Blocks"},
				},
				{
					OutwardIssue: jira.Issue{Key: "FOO-32"},
					Type:         jira.LinkType{Name: "Blocks"},
				},
				{
					InwardIssue: jira.Issue{Key: "FOO-71"},
					Type:        jira.LinkType{Name: "Blocks"},
				},
				{
					InwardIssue: jira.Issue{Key: "FOO-72"},
					Type:        jira.LinkType{Name: "Blocks"},
				},
			},
		},
	}

	rosina.AssertDeepEqual(t, have, want, "issues with deps")
}
