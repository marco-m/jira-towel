package towel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// CustomfieldValue returns the value of custom field 'name' from map
// 'customFields', which is assumed to be filled by github.com/mitchellh/mapstructure.
// See the tests in customfields_test for an example.
// If 'name' is not present, CustomfieldValue returns the empty string.
// CustomfieldValue assumes that lookup table 'lut' is filled by manual inspection
// of the JSON object returned by Jira.
// Yes, this sucks.
func CustomfieldValue(customFields map[string]any, lut map[string]int, name string) string {
	id, found := lut[name]
	if !found {
		return ""
	}
	cfName := fmt.Sprintf("customfield_%d", id)
	// A customfield JSON object has the following shape. We want the "value" field:
	// "customfield_11919": {
	//       "self": "https://x.atlassian.net/rest/api/2/customFieldOption/10837",
	//       "value": "Foo Bar", <=== THIS
	//       "id": "10837"
	//     },
	//
	// Convert from any to the expected shape, part 1
	cfMap, ok := customFields[cfName].(map[string]any)
	if !ok {
		return ""
	}
	// Convert from any to the expected shape, part 2
	value, ok := cfMap["value"].(string)
	if !ok {
		return ""
	}
	return value
}

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
	Expand string  `json:"expand"`
	Issues []issue `json:"issues"`
}

type pagination struct {
	StartAt    int `json:"startAt"`
	MaxResults int `json:"maxResults"`
	Total      int `json:"total"`
}

type issue struct {
	Key    string `json:"key"`
	Fields fields `json:"fields"`
}

type fields struct {
	IssueType issueType `json:"issuetype"`
	Parent    *issue    `json:"parent"`
	Project   project   `json:"project"`
	Priority  struct {
		Name string `json:"name"`
	} `json:"priority"`
	Labels      []interface{} `json:"labels"`
	Issuelinks  []issuelink   `json:"issuelinks"`
	Assignee    interface{}   `json:"assignee"`
	Status      status        `json:"status"`
	Description interface{}   `json:"description"`
	Summary     string        `json:"summary"`
	Creator     struct {
		DisplayName string `json:"displayName"`
	} `json:"creator"`
	Subtasks []interface{} `json:"subtasks"`
	Duedate  interface{}   `json:"duedate"`
	Progress struct {
		Progress int `json:"progress"`
		Total    int `json:"total"`
	} `json:"progress"`

	// HACK. Sigh.
	// I think I never found something as badly designed as Jira.
	// Tag 'remain' tells 'mapstructure' to collect all unknown fields, at any
	// level of nesting.
	CustomFields map[string]any `mapstructure:",remain"`
}

type issuelink struct {
	OutwardIssue issue `json:"outwardIssue"`
	InwardIssue  issue `json:"inwardIssue"`
	// NOTE The 3 fields below are always the same, for example Name = Blocks,
	// so there is no direction information! The direction is determined by
	// which one of OutwardIssue or InwardIssue is filled.
	Type struct {
		Inward  string `json:"inward"`
		Name    string `json:"name"`
		Outward string `json:"outward"`
	} `json:"type"`
}

type project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type issueType struct {
	Name    string `json:"name"`
	Subtask bool   `json:"subtask"`
}

type status struct {
	Name string `json:"name"`
}

func doQuery(httpClient *http.Client, config Config, jql string,
) ([][]byte, int, error) {
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
			return nil, 0, fmt.Errorf("query: %s", err)
		}
		reply, err := post(ctx, httpClient, config.Email, config.ApiToken, endpoint,
			bytes.NewReader(reqBody))
		if err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(reply, &pagination); err != nil {
			return nil, 0, err
		}
		result = append(result, reply)

		req.StartAt += pagination.MaxResults
		if req.StartAt >= pagination.Total {
			break
		}
		fmt.Fprint(os.Stderr, req.StartAt, " ")
	}

	return result, pagination.Total, nil
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
