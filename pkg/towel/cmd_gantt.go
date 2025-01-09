package towel

import (
	"slices"

	"github.com/marco-m/jira-towel/pkg/jira"
	"github.com/marco-m/jira-towel/pkg/lists"
)

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
