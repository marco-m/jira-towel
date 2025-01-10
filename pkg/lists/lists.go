package lists

type Node[T any] struct {
	Item T
	next *Node[T]
}

func (i *Node[T]) Next() *Node[T] {
	return i.next
}

// Do not use directly; instead, instantiate with [New].
type List[T any] struct {
	// The lists in this package are based on the concept of a dummy head, as
	// explained in Sedgewick, _Algorithms in C_. See how [New] does the init.
	head *Node[T]
	tail *Node[T]
	len  int
}

func New[T any](vs ...T) *List[T] {
	list := &List[T]{head: &Node[T]{}}
	for _, v := range vs {
		list.PushBack(v)
	}
	return list
}

// Len returns the number of elements in the list.
func (ls *List[T]) Len() int {
	return ls.len
}

// Front returns the first node. It will be nil if the list is empty.
func (ls *List[T]) Front() *Node[T] {
	return ls.head.next
}

// PopFront pops the first node. It will be nil if the list is empty.
func (ls *List[T]) PopFront() *Node[T] {
	x := ls.head.next
	if ls.len == 0 {
		return x
	}
	if ls.len == 1 {
		ls.head.next = nil
		ls.tail = nil
		ls.len--
		x.next = nil
		return x
	}
	ls.head.next = x.next
	ls.len--
	x.next = nil
	return x
}

// Back returns the last node. It will be nil if the list is empty.
func (ls *List[T]) Back() *Node[T] {
	return ls.tail
}

func (ls *List[T]) PushBack(v T) *Node[T] {
	node := &Node[T]{Item: v}
	if ls.len > 0 {
		ls.tail.next = node
	} else {
		ls.head.next = node
	}
	ls.tail = node
	ls.len++
	return node
}

func (ls *List[T]) PushFront(v T) *Node[T] {
	node := &Node[T]{Item: v}
	node.next = ls.head.next
	ls.head.next = node
	// If the list is empty, we need to update also the tail.
	if ls.tail == nil {
		ls.tail = node
	}
	ls.len++
	return node
}

func (ls *List[T]) ToSlice() []T {
	result := make([]T, 0, ls.len)
	for e := ls.Front(); e != nil; e = e.Next() {
		result = append(result, e.Item)
	}
	return result
}

// InsertAfter inserts 'v' after 'node'.
func (ls *List[T]) InsertAfter(v T, node *Node[T]) *Node[T] {
	newNode := &Node[T]{Item: v}
	if node == ls.tail {
		ls.tail = newNode
	} else {
		newNode.next = node.next
	}
	node.next = newNode
	ls.len++
	return newNode
}

func (ls *List[T]) Traverse(visitor func(v T)) {
	for node := ls.Front(); node != nil; node = node.Next() {
		visitor(node.Item)
	}
}

// insertAfter
// removeAfter

// InsertionSort sorts 'list' in place. Assumes that 'list' has been created by
// [New], that is, it must have a dummy head (see explanation for [List]).
// Complexity: O(n^2)
func InsertionSort[T any](list *List[T], less func(a, b T) bool) {
	// heada := Node[T]{}
	// a := &heada
	a := list.head
	t := a

	headb := Node[T]{}
	b := &headb
	var u, x *Node[T]

	for t = a.next; t != nil; t = u {
		u = t.next
		newTail := true
		for x = b; x.next != nil; x = x.next {
			if less(t.Item, x.next.Item) {
				newTail = false
				break
			}
		}
		t.next = x.next
		x.next = t
		if newTail {
			list.tail = t
		}
	}

	// Set input 'list' to the sorted list.
	list.head = b
}

// Josephus forms a circle of 'N' persons and eliminates every 'M'th person
// around the circle. It returns the sequence of the eliminations. The last
// element is the elected.
func Josephus(N int, M int) []int {
	t := &Node[int]{Item: 1}
	t.next = t // Make a circular list.

	// Fill the list.

	x := t
	for i := 2; i <= N; i++ {
		x.next = &Node[int]{Item: i, next: t}
		x = x.next
	}
	// Now x points to the Nth node. This is fundamental.

	// Eliminate the Mth until only one remains.
	seq := make([]int, 0, N)
	// When x == x.next, there is only 1 element in the circular list.
	for x != x.next {
		for i := 1; i < M; i++ {
			x = x.next
		}
		// Eliminate x.next
		seq = append(seq, x.next.Item)
		x.next = x.next.next
	}

	// x points to the elected.
	seq = append(seq, x.Item)
	return seq
}
