package towel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/marco-m/jira-towel/pkg/text"
)

type queryRequest struct {
	// TODO if we list explicitly the fields we want, we might even get
	//   a faster reply.
	//Fields []string    `json:"fields"`
	JQL        string `json:"jql"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
}

type queryResponse struct {
	Expand     string  `json:"expand"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []issue `json:"issues"`
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

func cmdGraph(global Global, graph GraphCmd) error {
	config, err := loadConfig(global.ConfigDir)
	if err != nil {
		return fmt.Errorf("graph: %w", err)
	}

	queryResp, err := doQuery(config, graph.JQL)
	if err != nil {
		return fmt.Errorf("graph: %s", err)
	}

	printSummary(queryResp.Issues)
	fmt.Println()
	fmt.Println(makeGraph(queryResp.Issues))
	return nil
}

func makeGraph(issues []issue) string {
	var bld strings.Builder
	fmt.Fprint(&bld, `
digraph {
    #rankdir=LR
    rankdir=TB
    node [shape=box style=filled width=3.5 height=1  fixedsize="true"]

`)
	for _, ticket := range issues {
		fmt.Fprintln(&bld, makeNode(ticket, "    "))
		for _, edge := range makeEdges(ticket, "    ") {
			fmt.Fprintln(&bld, edge)
		}
	}
	fmt.Fprintln(&bld, `
}`)
	return bld.String()
}

func makeNode(ticket issue, indent string) string {
	const maxWidth = 30
	if ticket.Fields.IssueType.Name == "Epic" {
		return ""
	}
	key := ticket.Key
	status := ticket.Fields.Status.Name
	parent := "towel bug: no parent?"
	// FIXME this is enough not to crash but it means we still miss the parent
	//   epic for some reasons I do not understand.
	if ticket.Fields.Parent != nil {
		parent = ticket.Fields.Parent.Fields.Summary
	}
	summary := ticket.Fields.Summary
	label := fmt.Sprintf("%s\n(%s)\n%s %s",
		text.ShortenMiddle(summary, maxWidth), text.ShortenMiddle(parent, maxWidth),
		key, status)
	return fmt.Sprintf("%s%q [label=%q fillcolor=%q]",
		indent, key, label, nodeColor(status))
}

func nodeColor(status string) string {
	switch strings.ToLower(status) {
	case "to do":
		return "cadetblue1"
	case "in progress":
		return "orange"
	case "done":
		return "yellowgreen"
	default:
		return "gray"
	}
}

func makeEdges(ticket issue, indent string) []string {
	var output []string
	links := ticket.Fields.Issuelinks
	src := ticket.Key
	var dst string
	var relation string
	// FIXME with this simplistic approach (skip the notion of graph),
	//   I cannot then decorate the dst with the red border if the
	//   relation is "blocks" :-(
	for _, link := range links {
		// NOTE It should be impossible to have both Outward and Inward at
		// the same time...
		if link.OutwardIssue.Key != "" {
			dst = link.OutwardIssue.Key
			relation = link.Type.Outward
			output = append(output,
				fmt.Sprintf("%s%q -> %q [label=%q color=%q]\n",
					indent, src, dst, relation, edgeColor(relation)))
		}
		//if link.InwardIssue.Key != "" {
		//	dst = link.InwardIssue.Key
		//	relation = link.Type.Inward
		//}
	}
	return output
}

func edgeColor(relation string) string {
	switch strings.ToLower(relation) {
	case "blocks":
		return "red"
	default:
		return "black"
	}
}

func printSummary(issues []issue) {
	fmt.Printf("received %d issues\n", len(issues))
	for _, ticket := range issues {
		fmt.Println("============")
		printFirstLine(ticket)
		printParent(ticket)
		printRelations(ticket)
	}
}

func printFirstLine(issue issue) {
	fmt.Printf("%s (%s) %s\n",
		issue.Key, issue.Fields.IssueType.Name, issue.Fields.Summary)
}

func printParent(issue issue) {
	fmt.Print("parent: ")
	if issue.Fields.Parent != nil {
		printFirstLine(*issue.Fields.Parent)
	} else {
		fmt.Println("<none>")
	}
}

func printRelations(issue issue) {
	issueLinks := issue.Fields.Issuelinks
	for _, link := range issueLinks {
		// NOTE It should be impossible to have both Outward and Inward.
		if link.OutwardIssue.Key != "" {
			fmt.Printf("%s: %s\n", link.Type.Outward, link.OutwardIssue.Key)
		}
		if link.InwardIssue.Key != "" {
			fmt.Printf("%s: %s\n", link.Type.Inward, link.InwardIssue.Key)
		}
	}
}

func post(
	ctx context.Context, hclient *http.Client, user string, token string, uri string, reqBody io.Reader,
) ([]byte, error) {
	return do(ctx, hclient, user, token, uri, reqBody, http.MethodPost)
}

func get(
	ctx context.Context, hclient *http.Client, user string, token string, uri string,
) ([]byte, error) {
	return do(ctx, hclient, user, token, uri, nil, http.MethodGet)
}

func do(
	ctx context.Context, hclient *http.Client, user string, token string, uri string, reqBody io.Reader, method string,
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
	defer resp.Body.Close()

	body, errBody := io.ReadAll(resp.Body)

	//fmt.Println("header:", repr.String(resp.Header, repr.Indent("  ")))

	// Shadoks: "Why do it the easy way when you can do it the hard way?"
	// https://en.wikipedia.org/wiki/Les_Shadoks
	seraph := resp.Header.Get("X-Seraph-Loginreason")

	if resp.StatusCode != http.StatusOK {
		if errBody != nil {
			return nil, fmt.Errorf("do: status code: %d seraph: %q (%s)", resp.StatusCode, seraph, errBody)
		}
		return nil, fmt.Errorf("do: status code: %d seraph: %q (%s)", resp.StatusCode, seraph, string(body))
	}
	if errBody != nil {
		return body, fmt.Errorf("do: read body: %s seraph: %q", errBody, seraph)
	}
	if seraph != "" {
		return body, fmt.Errorf("do: StatusCode 200 but Jira says seraph: %q", seraph)
	}

	return body, nil
}
