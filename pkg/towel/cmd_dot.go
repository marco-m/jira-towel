package towel

import (
	"github.com/dominikbraun/graph"
	"github.com/marco-m/clim"
	"github.com/marco-m/jira-towel/pkg/jira"
)

type dotCmd struct {
	// JQL string `help:"JQL"`
}

func newDotCLI() *clim.CLI[App] {
	dotCmd := dotCmd{}

	cli := clim.New("dot", "generate a graphviz DOT file (WIP)",
		dotCmd.Run)

	return cli
}

func (cmd *dotCmd) Run(app App) error {
	issueHash := func(c jira.Issue) string {
		return c.Key
	}
	g := graph.New(issueHash)

	// so it seems that i need to parse the jira.Issue, since the issuelinks field is actually a list of edges!
	// and i must add the nodes in the issuelink, also if incomplete (only Key known) and maybe already present...

	_ = g.AddVertex(jira.Issue{
		Key: "CICCIO-1",
		Fields: jira.Fields{
			Status: jira.Status{Name: "to do"},
			IssueLinks: []jira.IssueLink{
				{
					OutwardIssue: jira.Issue{
						Key:    "CICCIO-2",
						Fields: jira.Fields{},
					},
				},
			},
		},
	})
	return nil
}
