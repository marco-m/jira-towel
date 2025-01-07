package towel

import (
	"fmt"
	"os"

	"github.com/marco-m/clim"
)

type queryCmd struct {
	JQL string
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

	return cli
}

func (cmd *queryCmd) Run(app App) error {
	config, err := loadConfig(app.ConfigDir)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	jsonResponses, err := doQuery(app.HttpClient, config, cmd.JQL)
	if err != nil {
		return fmt.Errorf("query: %s", err)
	}

	fmt.Fprintln(os.Stderr, "total:", len(jsonResponses))
	for _, resp := range jsonResponses {
		fmt.Println(string(resp))
	}

	// var queryResp queryResponse
	// if err := json.Unmarshal(jsonResponse, &queryResp); err != nil {
	// 	return fmt.Errorf("query: JSON: %s", err)
	// }
	// usequeryResp

	return nil
}
