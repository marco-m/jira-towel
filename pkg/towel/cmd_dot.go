package towel

import (
	"github.com/dominikbraun/graph"
	"github.com/marco-m/clim"
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
	issueHash := func(c issue) string {
		return c.Key
	}
	g := graph.New(issueHash)

	// so it seems that i need to parse the issue, since the issuelinks field is actually a list of edges!
	// and i must add the nodes in the issuelink, also if incomplete (only Key known) and maybe already present...

	_ = g.AddVertex(issue{
		Key: "CICCIO-1",
		Fields: fields{
			Status: status{Name: "to do"},
			Issuelinks: []issuelink{
				{
					OutwardIssue: issue{
						Key:    "CICCIO-2",
						Fields: fields{},
					},
				},
			},
		},
	})
	return nil
}
