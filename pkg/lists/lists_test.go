package lists_test

import (
	"fmt"
	"testing"

	"github.com/marco-m/jira-towel/pkg/lists"
	"github.com/marco-m/rosina"
)

func ExampleListIterateWithTraverse() {
	list := lists.New(10, 2, 5)
	// 7 10 2 5
	list.PushFront(7)
	// 8 7 10 2 5
	list.PushFront(8)
	// 8 7 10 2 5 1
	list.PushBack(1)

	list.Traverse(func(v int) { fmt.Printf("%v ", v) })
	// Output: 8 7 10 2 5 1
}

func ExampleListIterateWithNext() {
	list := lists.New(4)
	// 1 4
	i1 := list.PushFront(1)
	// 1 2 4
	list.InsertAfter(2, i1)

	// Iterate through list and print its contents.
	for e := list.Front(); e != nil; e = e.Next() {
		fmt.Printf("%v ", e.Item)
	}
	// Output: 1 2 4
}

func ExampleListInsertingWhileIterating() {
	list := lists.New(1, 3, 5)

	for e := list.Front(); e != nil; e = e.Next() {
		// WARNING Be careful not to create an infinite loop!
		if e.Item%2 != 0 {
			list.InsertAfter(e.Item+1, e)
		}
	}

	list.Traverse(func(v int) { fmt.Print(v, " ") })
	// Output: 1 2 3 4 5 6
}

func TestPushFront(t *testing.T) {
	list := lists.New[int]()

	list.PushFront(1)

	want := []int{1}
	rosina.AssertDeepEqual(t, list.ToSlice(), want, "slice")
	rosina.AssertEqual(t, list.Len(), 1, "len")
	rosina.AssertEqual(t, list.Front().Item, want[0], "front")
	rosina.AssertEqual(t, list.Back().Item, want[len(want)-1], "back")
}

func TestPushBack(t *testing.T) {
	list := lists.New[int]()

	list.PushBack(1)

	want := []int{1}
	rosina.AssertDeepEqual(t, list.ToSlice(), want, "slice")
	rosina.AssertEqual(t, list.Len(), 1, "len")
	rosina.AssertEqual(t, list.Front().Item, want[0], "front")
	rosina.AssertEqual(t, list.Back().Item, want[len(want)-1], "back")
}

func TestPopFront(t *testing.T) {
	type testCase struct {
		name         string
		list         *lists.List[int]
		wantPopFront int
		wantBack     int
		wantRemain   []int
	}

	test := func(t *testing.T, tc testCase) {
		len := tc.list.Len()
		have := tc.list.PopFront()
		switch len {
		case 0:
			rosina.AssertEqual(t, have, nil, "pop front, len 0")
			rosina.AssertEqual(t, tc.list.Back(), nil, "back, len 0")
		case 1:
			rosina.AssertEqual(t, have.Item, tc.wantPopFront, "pop front, len 1")
			rosina.AssertEqual(t, tc.list.Back(), nil, "back, len 1")
		default:
			rosina.AssertEqual(t, have.Item, tc.wantPopFront, "pop front, len > 1")
			rosina.AssertEqual(t, tc.list.Back().Item, tc.wantBack, "back, len > 1")
		}
		rosina.AssertDeepEqual(t, tc.list.ToSlice(), tc.wantRemain, "remain")
	}

	testCases := []testCase{
		{
			name:       "empty list",
			list:       lists.New[int](),
			wantRemain: []int{},
		},
		{
			name:         "one element",
			list:         lists.New(1),
			wantPopFront: 1,
			wantRemain:   []int{},
		},
		{
			name:         "four element",
			list:         lists.New(1, 2, 3, 4),
			wantPopFront: 1,
			wantBack:     4,
			wantRemain:   []int{2, 3, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestPopFrontConsumesTheList(t *testing.T) {
	list := lists.New(1, 2, 3, 4)

	have := list.PopFront()
	rosina.AssertEqual(t, have.Item, 1, "pop 1")
	rosina.AssertEqual(t, list.Len(), 3, "len 3")
	rosina.AssertEqual(t, have.Next(), nil, "next")

	have = list.PopFront()
	rosina.AssertEqual(t, have.Item, 2, "pop 2")
	rosina.AssertEqual(t, list.Len(), 2, "len 2")
	rosina.AssertEqual(t, have.Next(), nil, "next")

	have = list.PopFront()
	rosina.AssertEqual(t, have.Item, 3, "pop 3")
	rosina.AssertEqual(t, list.Len(), 1, "len 1")
	rosina.AssertEqual(t, have.Next(), nil, "next")

	have = list.PopFront()
	rosina.AssertEqual(t, have.Item, 4, "pop 4")
	rosina.AssertEqual(t, list.Len(), 0, "len 0")
	rosina.AssertEqual(t, have.Next(), nil, "next")

	have = list.PopFront()
	rosina.AssertEqual(t, have, nil, "pop nil")
	rosina.AssertEqual(t, list.Len(), 0, "len 0")
}

func TestListNewAndToSlice(t *testing.T) {
	type testCase struct {
		name string
		want []int
	}

	test := func(t *testing.T, tc testCase) {
		list := lists.New(tc.want...)
		have := list.ToSlice()
		rosina.AssertDeepEqual(t, have, tc.want, "slice")
		rosina.AssertEqual(t, list.Len(), len(have), "len")
		if len(have) > 0 {
			rosina.AssertEqual(t, list.Front().Item, tc.want[0], "front")
			rosina.AssertEqual(t, list.Back().Item, tc.want[len(tc.want)-1], "back")
		}
	}

	testCases := []testCase{
		{
			name: "empty list",
			want: []int{},
		},
		{
			name: "some elements",
			want: []int{1, 3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestListInsertionSort(t *testing.T) {
	type testCase struct {
		name  string
		input *lists.List[int]
		want  []int
	}

	test := func(t *testing.T, tc testCase) {
		lists.InsertionSort(tc.input, func(a, b int) bool { return a < b })

		have := tc.input
		rosina.AssertDeepEqual(t, have.ToSlice(), tc.want, "sort")
		rosina.AssertEqual(t, have.Len(), len(tc.want), "len")
		if have.Len() > 0 {
			rosina.AssertEqual(t, have.Front().Item, tc.want[0], "front")
			rosina.AssertEqual(t, have.Back().Item, tc.want[len(tc.want)-1], "back")
		}
	}

	testCases := []testCase{
		{
			name:  "empty list",
			input: lists.New[int](),
			want:  []int{},
		},
		{
			name:  "1 element",
			input: lists.New(1),
			want:  []int{1},
		},
		{
			name:  "2 elements, unsorted",
			input: lists.New(2, 1),
			want:  []int{1, 2},
		},
		{
			name:  "4 elements, unsorted",
			input: lists.New(5, 3, 4, 2),
			want:  []int{2, 3, 4, 5},
		},
		{
			name:  "4 elements, already sorted",
			input: lists.New(2, 3, 4, 5),
			want:  []int{2, 3, 4, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestListsJosephus(t *testing.T) {
	have := lists.Josephus(9, 5)
	want := []int{5, 1, 7, 4, 3, 6, 9, 2, 8}
	rosina.AssertDeepEqual(t, have, want, "sequence")
}
