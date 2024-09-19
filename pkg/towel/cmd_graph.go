package towel

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/marco-m/clim"
	"github.com/marco-m/jira-towel/pkg/text"
	"github.com/mitchellh/mapstructure"
)

type graphCmd struct {
	JQL          string
	DotPath      string
	Rankdir      string
	CustomFields []string
	CfLUT        map[string]int
	ClusterBy    string
}

func newGraphCLI() *clim.CLI[App] {
	graphCmd := graphCmd{
		CfLUT: make(map[string]int),
	}

	cli := clim.New("graph", "generate the dependency graph of a set of tickets",
		graphCmd.Run)

	cli.AddFlag(&clim.Flag{
		Value: clim.String(&graphCmd.JQL, ""),
		Long:  "jql", Label: "QUERY",
		Help:     "JQL query, for example: 'project = \"MY PROJECT\"''. An empty string is not accepted because it would query ALL the projects in the Jira instance",
		Required: true,
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.String(&graphCmd.DotPath, "graph.dot"),
		Long:  "dot",
		Help:  "File to write the DOT graph to",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.String(&graphCmd.Rankdir, "LR"),
		Long:  "rankdir",
		Help:  "DOT rankdir (LR, TB)",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.StringSlice(&graphCmd.CustomFields, nil),
		Long:  "custom-fields", Label: "name:id[,name:id,..]",
		Help: "List of customfield names to IDs (eg: product:37,feature:42)",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.String(&graphCmd.ClusterBy, ""),
		Long:  "cluster-by", Label: "CUSTOM-FIELD",
		Help: "Name of the custom field to cluster by (needs also --custom-fields)",
	})

	return cli
}

func (cmd *graphCmd) Run(app App) error {
	config, err := loadConfig(app.ConfigDir)
	if err != nil {
		return fmt.Errorf("graph: %w", err)
	}

	for _, kv := range cmd.CustomFields {
		k, v, found := strings.Cut(kv, ":")
		if !found {
			return clim.ParseError("custom-fields: %q: missing separator ':'", kv)
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			return clim.ParseError("custom-fields: %q: %q is not a number: %s",
				kv, v, err)
		}
		cmd.CfLUT[k] = id
	}

	jsonResponses, count, err := doQuery(app.HttpClient, config, cmd.JQL)
	if err != nil {
		return fmt.Errorf("graph: %s", err)
	}
	issues := make([]issue, 0, count)

	for _, jsonresp := range jsonResponses {
		var parsedMap map[string]any
		if err := json.Unmarshal(jsonresp, &parsedMap); err != nil {
			return fmt.Errorf("query: JSON: %s", err)
		}
		var queryResp queryResponse
		if err = mapstructure.Decode(parsedMap, &queryResp); err != nil {
			return fmt.Errorf("query: mapstructure: %s", err)
		}
		issues = append(issues, queryResp.Issues...)
	}

	printSummary(issues)

	dot := makeGraph(issues, cmd.Rankdir, cmd.CfLUT, cmd.ClusterBy)
	if err := os.WriteFile(cmd.DotPath, []byte(dot), 0o660); err != nil {
		return fmt.Errorf("writing %s: %s", cmd.DotPath, err)
	}
	return nil
}

func makeGraph(issues []issue, rankdir string, lut map[string]int, clusterby string) string {
	var bld strings.Builder
	fmt.Fprintln(&bld, "digraph {")
	fmt.Fprintf(&bld, "    rankdir=%s\n", rankdir)
	fmt.Fprintln(&bld, `    node [shape=box style=filled width=3.5 height=0.5 fixedsize="true"]`)
	fmt.Fprintln(&bld)

	indent := "    "
	clusters := make(map[string][]string)

	for _, ticket := range issues {
		clusterName := CustomfieldValue(ticket.Fields.CustomFields, lut, clusterby)
		clusters[clusterName] = append(clusters[clusterName], ticket.Key)
		fmt.Fprintln(&bld, makeNode(ticket, indent))
		for _, edge := range makeEdges(ticket, indent) {
			fmt.Fprintln(&bld, edge)
		}
	}

	fmt.Fprintln(&bld, makeClusters(clusters, indent))

	fmt.Fprintln(&bld, "}")
	return bld.String()
}

//	subgraph cluster_0 {
//		label = "process #1";
//		style=filled;
//		color=lightgrey;
//		node [style=filled,color=white];
//		a0 -> a1 -> a2 -> a3;
//	}
func makeClusters(clusters map[string][]string, indent string) string {
	var bld strings.Builder
	invisible := 0
	for clusterName, nodeNames := range clusters {

		// hack graphviz bug. Invisible cluster
		// https://forum.graphviz.org/t/how-to-add-space-between-clusters/1209
		invisible++
		fmt.Fprintf(&bld, "%ssubgraph cluster_wrap_%d {\n", indent, invisible)
		fmt.Fprintf(&bld, "%scolor=%q\n", indent, "white")

		if clusterName == "" {
			clusterName = "unknown"
		}
		fmt.Fprintf(&bld, "%ssubgraph \"cluster_%s\" {\n", indent, clusterName)
		// sigh this is internal margin!
		// fmt.Fprintf(&bld, "margin=55\n")
		fmt.Fprintf(&bld, "%s%slabel=%q style=filled color=%q\n", indent, indent,
			clusterName, "aquamarine")
		for _, nodeName := range nodeNames {
			fmt.Fprintf(&bld, "%s%s%q\n", indent, indent, nodeName)
		}
		fmt.Fprintf(&bld, "%s}\n", indent)

		// close wrap cluster hack see above
		fmt.Fprintf(&bld, "%s}\n", indent)
	}
	return bld.String()
}

func makeNode(ticket issue, indent string) string {
	const maxWidth = 40
	// if ticket.Fields.IssueType.Name == "Epic" {
	// 	return ""
	// }
	key := ticket.Key
	status := ticket.Fields.Status.Name
	// HACK
	// product := ticket.Fields.Product.Name
	// parent := "towel bug: no parent?"
	// FIXME this is enough not to crash but it means we still miss the parent
	//   epic for some reasons I do not understand.
	// if ticket.Fields.Parent != nil {
	// 	parent = ticket.Fields.Parent.Fields.Summary
	// }
	summary := ticket.Fields.Summary
	// label := fmt.Sprintf("%s\n(%s)\n%s %s",
	// 	text.ShortenMiddle(summary, maxWidth), text.ShortenMiddle(parent, maxWidth),
	// 	key, status)
	label := fmt.Sprintf("%s\n%s %s",
		text.ShortenMiddle(summary, maxWidth), key, status)
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
		// fmt.Printf("customfields: %#v\n", ticket.Fields.CustomFields)
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
