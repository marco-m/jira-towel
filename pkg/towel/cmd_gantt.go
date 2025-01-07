package towel

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/marco-m/clim"
	"github.com/marco-m/jira-towel/pkg/jira"
	"github.com/marco-m/jira-towel/pkg/lists"
	"github.com/marco-m/jira-towel/pkg/text"
)

type ganttCmd struct {
	JQL       string
	GanttPath string
}

func newGanttCLI() *clim.CLI[App] {
	ganttCmd := ganttCmd{}

	cli := clim.New("gantt", "generate a GANTT chart of a set of tickets",
		ganttCmd.Run)

	cli.AddFlag(&clim.Flag{
		Value: clim.String(&ganttCmd.JQL, ""),
		Long:  "jql", Label: "QUERY",
		Help:     "JQL query, for example: 'project = \"MY PROJECT\"''. An empty string is not accepted because it would query ALL the projects in the Jira instance",
		Required: true,
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.String(&ganttCmd.GanttPath, "gantt.puml"),
		Long:  "gantt",
		Help:  "File to write the GANTT chart to",
	})

	return cli
}

func (cmd *ganttCmd) Run(app App) error {
	const op = "gantt"
	config, err := loadConfig(app.ConfigDir)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	issues, err := fetchIssues(app.HttpClient, config, cmd.JQL)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	// printSummary(issues)

	gantt := makeGantt(issues)
	if err := os.WriteFile(cmd.GanttPath, []byte(gantt), 0o660); err != nil {
		return fmt.Errorf("%s: writing %s: %s", op, cmd.GanttPath, err)
	}

	return nil
}

func makeGantt(issues []jira.Issue) string {
	var bld strings.Builder
	fmt.Fprintln(&bld, "@startgantt name")
	fmt.Fprintln(&bld)
	fmt.Fprintln(&bld, "saturday are closed")
	fmt.Fprintln(&bld, "sunday are closed")
	fmt.Fprintln(&bld, "hide footbox")

	fmt.Fprintln(&bld, "project starts 2025-01-06")
	fmt.Fprintln(&bld, "projectscale weekly zoom 6")
	fmt.Fprintln(&bld)

	idToIssue := make(map[string]jira.Issue, len(issues))
	for _, ticket := range issues {
		idToIssue[ticket.Key] = ticket
	}

	chains := extractDependencyChains(issues)

	for chainNode := chains.Front(); chainNode != nil; chainNode = chainNode.Next() {
		for issueNode := chainNode.Item.Front(); issueNode != nil; issueNode = issueNode.Next() {
			issue := issueNode.Item
			fmt.Fprintln(&bld, makeGanttNode(issue))
		}
	}
	fmt.Fprintln(&bld)
	for chainNode := chains.Front(); chainNode != nil; chainNode = chainNode.Next() {
		for issueNode := chainNode.Item.Front(); issueNode != nil; issueNode = issueNode.Next() {
			issue := issueNode.Item
			for _, dep := range makeGanttDependencies(issue, idToIssue) {
				fmt.Fprintln(&bld, dep)
			}
		}
	}
	fmt.Fprintln(&bld)

	fmt.Fprintln(&bld, "footer \\n\\n    generated 2025-01-07 by jira-towel (add full CLI here?)")
	fmt.Fprintln(&bld)
	fmt.Fprintln(&bld, "@endgantt")
	return bld.String()
}

// [FOO-1 Prototype design] as [FOO-1] requires 1 week
func makeGanttNode(ticket jira.Issue) string {
	const maxWidth = 40
	// The PlantUML language does not support nested square brackets [], not even
	// backslack-escaped or quoted, so we transform to parenthesis ().
	replacer := strings.NewReplacer(
		"[", "(",
		"]", ")",
	)
	summary := replacer.Replace(ticket.Fields.Summary)
	label := fmt.Sprintf("%s %s", ticket.Key,
		text.ShortenMiddle(summary, maxWidth))
	return fmt.Sprintf("[%s] as [%s] requires %d week", label, ticket.Key, 2)
}

func makeGanttDependencies(ticket jira.Issue, idToIssue map[string]jira.Issue) []string {
	var output []string
	src := ticket.Key
	for _, link := range ticket.Fields.IssueLinks {
		// NOTE It should be impossible to have both Outward and Inward at
		// the same time...
		if link.OutwardIssue.Key != "" {
			dst := link.OutwardIssue.Key
			// FIXME what to do if relation != "blocks"
			relation := link.Type.Outward
			if _, found := idToIssue[dst]; !found {
				// dst is not planned at the moment
				continue
			}
			// [dst] starts at [src]'s end
			commentLine := fmt.Sprintf("' src=%q dst=%q rel=%q", src, dst, relation)
			depLine := fmt.Sprintf("[%s] starts at [%s]'s end\n", dst, src)
			output = append(output, commentLine, depLine)
		}
		//if link.InwardIssue.Key != "" {
		//	dst = link.InwardIssue.Key
		//	relation = link.Type.Inward
		//}
	}
	return output
}

func extractDependencyChains(issues []jira.Issue) *lists.List[*lists.List[jira.Issue]] {
	chains := lists.New[*lists.List[jira.Issue]]()

	loop1 := func(issue jira.Issue) {
		inserted := false
		for chain := chains.Front(); chain != nil; chain = chain.Next() {
			// if chain.Item.Len() == 0 {
			// 	// FIXME I SHOULD DELETE THIS INSTEAD IN STEP 2
			// 	continue
			// }
			if slices.Contains(jira.BlockedBy(chain.Item.Back().Item), issue.Key) {
				chain.Item.PushBack(issue)
				inserted = true
				break
			}
			if slices.Contains(jira.Blocks(chain.Item.Front().Item), issue.Key) {
				chain.Item.PushFront(issue)
				inserted = true
				break
			}
		}
		// If we could not insert the issue into an existing chain, we create a
		// new chain and put the issue in it.
		if !inserted {
			chains.PushBack(lists.New(issue))
		}
	}

	//
	// Step 1. Place issues into chains. Begin putting related issues in the
	// same chain (note that this has only a partial effect).
	//
	for _, issue := range issues {
		loop1(issue)
	}

	//
	// Step 2. Ugly hack waiting for a better algorithm.
	// Remove all single-element chains.
	//
	//
	singles := lists.New[*lists.Node[jira.Issue]]()
	q := lists.New[*lists.List[jira.Issue]]()
	for chain := chains.PopFront(); chain != nil; chain = chains.PopFront() {
		if chain.Item.Len() > 1 {
			q.PushFront(chain.Item)
			continue
		}
		singles.PushFront(chain.Item.PopFront())
	}
	chains = q

	//
	// Step 3. As in Step 1, place issues into chains.
	//

	for node := singles.Front(); node != nil; node = node.Next() {
		loop1(node.Item.Item)
	}

	//
	// Step 4. Aggregate the chains.
	//
	// QUESTION Maybe we have to keep running the loop until no more changes are
	// produced?
	//
	// QUESTION It seems that in step 1, it is useless to try to be smart and
	// put multiple items in a single chain, since here we will do the same...
	//

	return chains
}
