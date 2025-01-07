package towel

import (
	"fmt"
	"os"

	"github.com/alecthomas/repr"
	"github.com/marco-m/clim"
)

type queryCmd struct {
	JQL      string
	DumpJson bool
	DumpGo   bool
}

func newQueryCLI() *clim.CLI[App] {
	queryCmd := queryCmd{}

	cli := clim.New("query", "issue a JQL query and dump its contents",
		queryCmd.Run)

	cli.AddFlag(&clim.Flag{
		Value: clim.String(&queryCmd.JQL, ""),
		Long:  "jql", Label: "QUERY",
		Help:     "JQL query, for example: 'project = \"MY PROJECT\"''. An empty string is not accepted because it would query ALL the projects in the Jira instance",
		Required: true,
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.Bool(&queryCmd.DumpJson, false),
		Long:  "dump-json",
		Help:  "Dump the received JSON objects to stdout",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.Bool(&queryCmd.DumpGo, false),
		Long:  "dump-go",
		Help:  "Dump the received objects as Go to stdout",
	})

	return cli
}

func (cmd *queryCmd) Run(app App) error {
	const op = "query"

	if !cmd.DumpJson && !cmd.DumpGo {
		return clim.ParseError("%s: specify at least one of --dump-json, --dump-go", op)
	}

	config, err := loadConfig(app.ConfigDir)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	jsonResponses, err := doQuery(app.HttpClient, config, cmd.JQL)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	fmt.Fprintln(os.Stderr, "total:", len(jsonResponses))

	if cmd.DumpJson {
		for _, resp := range jsonResponses {
			fmt.Println(string(resp))
		}
	}
	fmt.Println()

	if cmd.DumpGo {
		issues, err := parseResponses(jsonResponses)
		if err != nil {
			return fmt.Errorf("%s: %s", op, err)
		}
		for _, ticket := range issues {
			dumpTicket(ticket)
			fmt.Println()
		}
	}

	return nil
}

func dumpTicket(ticket issue) {
	pr := repr.New(os.Stdout, repr.IgnorePrivate())
	pr.Println(ticket)
}
