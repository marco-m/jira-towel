package towel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	jsonResponse, err := doQuery(app.HttpClient, config, cmd.JQL)
	if err != nil {
		return fmt.Errorf("query: %s", err)
	}

	fmt.Println(string(jsonResponse))

	// var queryResp queryResponse
	// if err := json.Unmarshal(jsonResponse, &queryResp); err != nil {
	// 	return fmt.Errorf("query: JSON: %s", err)
	// }
	// usequeryResp

	return nil
}

func doQuery(httpClient *http.Client, config configuration, jql string) ([]byte, error) {
	// It seems that the only difference between v2 and v3 is that v2
	// returns a plain text "Description" field, while v3 returns a
	// Jira-specific sort of rich text format.
	// For what we want to do, plain text is preferable.
	//
	//uri := "https://" + config.Server + "/rest/api/3/search"
	uri := "https://" + config.Server + "/rest/api/2/search"

	req := queryRequest{
		JQL:        jql,
		MaxResults: 50, // in any case pagination max is 50...
		StartAt:    0,
	}
	ctx := context.Background()
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("query: %s", err)
	}

	return post(ctx, httpClient, config.Email, config.ApiToken, uri, bytes.NewReader(reqBody))
}
