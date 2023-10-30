package towel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type QueryCmd struct {
	JQL string `help:"JQL"`
}

func cmdQuery(global Global, queryCmd QueryCmd) error {
	config, err := loadConfig(global.ConfigDir)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	queryResp, err := doQuery(config, queryCmd.JQL)
	if err != nil {
		return fmt.Errorf("query: %s", err)
	}

	// FIXME this is not good. I must dump the JSON, not this stuff!!!
	fmt.Println(queryResp)

	return nil
}

func doQuery(config configuration, jql string) (queryResponse, error) {
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
	hclient := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return queryResponse{}, fmt.Errorf("query: %s", err)
	}

	resp, err := post(ctx, hclient, config.Email, config.ApiToken, uri, bytes.NewReader(reqBody))
	if err != nil {
		return queryResponse{}, fmt.Errorf("query: %s", err)
	}
	var queryResp queryResponse
	if err := json.Unmarshal(resp, &queryResp); err != nil {
		return queryResponse{}, fmt.Errorf("query: JSON: %s", err)
	}

	return queryResp, nil
}
