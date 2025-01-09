package towel

import (
	"slices"
	"sort"
	"testing"

	tu "github.com/marco-m/jira-towel/internal/testutils"
	"github.com/marco-m/jira-towel/pkg/jira"
	"github.com/marco-m/jira-towel/pkg/lists"
	"github.com/marco-m/rosina"
)

// This test shows how to use the sorting facilities of the stdlib
// (sort.SliceStable and slices.SortStableFunc). Note that any stdlib sort
// function requires that the compare function is a _strict weak ordering_. See
// https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings.
//
// This requirement is completely reasonable! Unfortunately, this means that we
// cannot use any of these to sort the Gantt tasks, because Gantt tasks don't
// have (at least I could not find) a strict weak ordering compare function.
func TestUnderstandStdSorting(t *testing.T) {
	orig := []int{10, 100, 5, 2, 100}
	want := []int{100, 100, 2, 5, 10}

	t.Run("use old pkg sort", func(t *testing.T) {
		work := make([]int, len(orig))
		copy(work, orig)

		// Less reports whether x[i] must sort before x[j].
		// If both Less(i, j) and Less(j, i) are false,
		// then the elements at index i and j are considered equal.
		sort.SliceStable(work, func(i, j int) bool {
			// Fundamental: if uncomparable, push at the very beginning (or
			// could do also the opposite) BUT DO NOT report that the uncomparable
			// is EQUAL! Otherwise I stumble into my logic bug...
			if work[i] == 100 {
				return true // put work[i] left
			}
			if work[j] == 100 {
				return false // put work[j] left
			}
			return work[i] < work[j]
		})
		rosina.AssertDeepEqual(t, work, want, "sort")
	})

	t.Run("use new pkg slices", func(t *testing.T) {
		work := make([]int, len(orig))
		copy(work, orig)

		// cmp(a, b) should return:
		// - a negative number when a < b
		// - a positive number when a > b
		// - 0 when a == b or a and b are incomparable in the sense of a strict weak ordering.
		slices.SortStableFunc(work, func(a int, b int) int {
			// Fundamental: if uncomparable, push at the very beginning (or
			// could do also the opposite) BUT DO NOT report that the uncomparable
			// is EQUAL! Otherwise I stumble into my logic bug...
			if a == 100 {
				return -1 // put a left
			}
			if b == 100 {
				return +1 // put b left
			}
			return a - b
		})
		rosina.AssertDeepEqual(t, work, want, "slices")
	})
}

func TestExtractDependencyChains(t *testing.T) {
	issues := []jira.Issue{
		tu.BuildIssue(tu.BuildArg{
			Key: "FOO-4081", Summary: "a: 3",
			BlockedBy: []string{"FOO-4032"},
		}),
		tu.BuildIssue(tu.BuildArg{
			Key: "FOO-5051", Summary: "float 1",
		}),
		tu.BuildIssue(tu.BuildArg{
			Key: "FOO-3816", Summary: "a: 1",
			Blocks: []string{"FOO-4032"},
		}),
		tu.BuildIssue(tu.BuildArg{
			Key: "FOO-5052", Summary: "float 2",
		}),
		tu.BuildIssue(tu.BuildArg{
			Key: "FOO-4032", Summary: "a: 2",
			BlockedBy: []string{"FOO-3816"},
			Blocks:    []string{"FOO-4081"},
		}),
	}
	// FIXME this want is WRONG!
	want := lists.New(lists.New(
		tu.BuildIssue(tu.BuildArg{
			Key:     "FOO-5051",
			Summary: "float 1",
		})))

	// FIXME delete this!!!
	// sorted := tu.BuildIssuesWithDeps(
	// 	tu.BasicIssue{
	// 		Key:     "FOO-5051",
	// 		Summary: "float 1",
	// 	},
	// 	tu.BasicIssue{
	// 		Key:     "FOO-5052",
	// 		Summary: "float 2",
	// 	},
	// 	tu.BasicIssue{
	// 		Key:     "FOO-3816",
	// 		Summary: "a: 1",
	// 		Blocks:  []string{"FOO-4032"},
	// 	},
	// 	tu.BasicIssue{
	// 		Key:       "FOO-4032",
	// 		Summary:   "a: 2",
	// 		BlockedBy: []string{"FOO-3816"},
	// 		Blocks:    []string{"FOO-4081"},
	// 	},
	// 	tu.BasicIssue{
	// 		Key:       "FOO-4081",
	// 		Summary:   "a: 3",
	// 		BlockedBy: []string{"FOO-4032"},
	// 	},
	// )

	have := extractDependencyChains(issues)

	if have, want := have.Len(), want.Len(); have != want {
		t.Errorf("\nlist len mismatch\nhave: %d; want: %d", have, want)
	}
	// Iterate through list and print its contents.
	for chain := have.Front(); chain != nil; chain = chain.Next() {
		t.Logf("===")
		var keys []string
		for elem := chain.Item.Front(); elem != nil; elem = elem.Next() {
			keys = append(keys, elem.Item.Key)
		}
		t.Log("  ", keys)
	}

	// FIXME FINISH ME
	// rosina.AssertDeepEqual(t, have, want, "dependency chains")
}
