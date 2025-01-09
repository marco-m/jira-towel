package towel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"

	"github.com/marco-m/jira-towel/pkg/jira"
)

type queryRequest struct {
	// TODO if we list explicitly the fields we want, we might even get
	//   a faster reply.
	// Fields []string    `json:"fields"`
	JQL        string `json:"jql"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
}

type queryResponse struct {
	pagination
	Expand string       `json:"expand"`
	Issues []jira.Issue `json:"issues"`
}

type pagination struct {
	StartAt    int `json:"startAt"`
	MaxResults int `json:"maxResults"`
	Total      int `json:"total"`
}

func doQuery(httpClient *http.Client, config Config, jql string,
) ([][]byte, error) {
	// It seems that the only difference between v2 and v3 is that v2
	// returns a plain text "Description" field, while v3 returns a
	// Jira-specific sort of rich text format.
	// For what we want to do, plain text is preferable.
	//
	//endpoint := "https://" + config.Server + "/rest/api/3/search"
	endpoint := "https://" + config.Server + "/rest/api/2/search"

	ctx := context.Background()
	req := queryRequest{
		JQL:        jql,
		MaxResults: 1_000,
		StartAt:    0,
	}
	var result [][]byte
	var pagination pagination

	defer fmt.Fprintln(os.Stderr)
	// The "range 500" is here only as a safety net to avoid infinite loops.
	for range 500 {
		reqBody, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("query: %s", err)
		}
		reply, err := post(ctx, httpClient, config.Email, config.ApiToken, endpoint,
			bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(reply, &pagination); err != nil {
			return nil, err
		}
		result = append(result, reply)

		req.StartAt += pagination.MaxResults
		if req.StartAt >= pagination.Total {
			break
		}
		fmt.Fprint(os.Stderr, req.StartAt, " ")
	}

	return result, nil
}

func post(
	ctx context.Context, hclient *http.Client, user string, token string,
	uri string, reqBody io.Reader,
) ([]byte, error) {
	return do(ctx, hclient, user, token, uri, reqBody, http.MethodPost)
}

//nolint:unused
func get(
	ctx context.Context, hclient *http.Client, user string, token string,
	uri string,
) ([]byte, error) {
	return do(ctx, hclient, user, token, uri, nil, http.MethodGet)
}

func do(
	ctx context.Context, hclient *http.Client, user string, token string,
	uri string, reqBody io.Reader, method string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, uri, reqBody)
	if err != nil {
		return nil, fmt.Errorf("do: new queryRequest: %s", err)
	}
	req.SetBasicAuth(user, token)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %s", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	body, errBody := io.ReadAll(resp.Body)

	// Shadoks: "Why do it the easy way when you can do it the hard way?"
	// https://en.wikipedia.org/wiki/Les_Shadoks
	seraph := resp.Header.Get("X-Seraph-Loginreason")

	if resp.StatusCode != http.StatusOK {
		if errBody != nil {
			return nil, fmt.Errorf("do: StatusCode: %d seraph: %q (%s)", resp.StatusCode, seraph, errBody)
		}
		return nil, fmt.Errorf("do: StatusCode: %d seraph: %q (%s)", resp.StatusCode, seraph, string(body))
	}
	if errBody != nil {
		return body, fmt.Errorf("do: read body: %s seraph: %q", errBody, seraph)
	}
	if seraph != "" {
		return body, fmt.Errorf("do: StatusCode 200 but Jira says seraph: %q", seraph)
	}

	return body, nil
}

// Like [doQuery], but it also does the parsing to a slice of [issue].
func fetchIssues(httpClient *http.Client, config Config, jql string) ([]jira.Issue, error) {
	const op = "fetch"
	jsonResponses, err := doQuery(httpClient, config, jql)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", op, err)
	}
	return parseResponses(jsonResponses)
}

func parseResponses(jsonResponses [][]byte) ([]jira.Issue, error) {
	const op = "parseResonses"
	issues := make([]jira.Issue, 0, len(jsonResponses))
	for _, jsonresp := range jsonResponses {
		var parsedMap map[string]any
		if err := json.Unmarshal(jsonresp, &parsedMap); err != nil {
			return nil, fmt.Errorf("%s: JSON: %s", op, err)
		}
		var queryResp queryResponse
		if err := mapstructure.Decode(parsedMap, &queryResp); err != nil {
			return nil, fmt.Errorf("%s: mapstructure: %s", op, err)
		}
		issues = append(issues, queryResp.Issues...)
	}
	return issues, nil
}
