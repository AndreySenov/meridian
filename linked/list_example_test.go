package linked_test

import (
	"fmt"

	"github.com/AndreySenov/meridian/linked"
)

func ExampleList() {
	// The zero value is an empty list ready to use.
	var l linked.List[int]

	l.PushBack(2)
	l.PushBack(3)
	l.PushFront(1)

	fmt.Println(l.ToSlice(), l.Len())
	// Output: [1 2 3] 3
}

func ExampleNewList() {
	l := linked.NewList("a", "b", "c")

	fmt.Println(l.ToSlice())
	fmt.Println(l.ToSliceBackward())
	// Output:
	// [a b c]
	// [c b a]
}

func ExampleNewListSeq() {
	l := linked.NewList(1, 2, 3)

	// Any iter.Seq works, including another list's.
	reversed := linked.NewListSeq(l.Backward())

	fmt.Println(reversed.ToSlice())
	// Output: [3 2 1]
}

func ExampleList_PopFront() {
	l := linked.NewList("a", "b")

	for {
		value, ok := l.PopFront()
		if !ok {
			break
		}
		fmt.Println(value)
	}

	fmt.Println(l.IsEmpty())
	// Output:
	// a
	// b
	// true
}

func ExampleList_All() {
	l := linked.NewList(1, 2, 3)

	for value := range l.All() {
		fmt.Println("forward:", value)
	}

	for value := range l.Backward() {
		fmt.Println("backward:", value)
	}

	// Output:
	// forward: 1
	// forward: 2
	// forward: 3
	// backward: 3
	// backward: 2
	// backward: 1
}

func ExampleList_Nodes() {
	l := linked.NewList(1, 2, 3, 4)

	// Removing the current node during iteration is safe.
	for node := range l.Nodes() {
		if node.Value()%2 == 0 {
			l.Remove(node)
		}
	}

	fmt.Println(l.ToSlice())
	// Output: [1 3]
}

func ExampleList_MoveToFront() {
	// A list of recently used keys, most recent first.
	recent := linked.NewList("a", "b", "c")

	recent.MoveToFront(recent.Back())

	fmt.Println(recent.ToSlice())
	// Output: [c a b]
}

func ExampleList_InsertAfter() {
	l := linked.NewList(1, 3)

	l.InsertAfter(2, l.Front())
	l.InsertBefore(0, l.Front())

	fmt.Println(l.ToSlice())
	// Output: [0 1 2 3]
}

func ExampleList_Remove() {
	l := linked.NewList(1, 2, 3)
	node := l.Front().Next()

	l.Remove(node)

	// A removed node is detached: removing it again does nothing, and the
	// list no longer accepts it.
	l.Remove(node)

	fmt.Println(l.ToSlice(), node.Value(), node.Next())
	// Output: [1 3] 2 <nil>
}

func ExampleNode() {
	l := linked.NewList(1, 2, 3)

	// Walking by node survives removing the node being visited.
	for node := l.Front(); node != nil; {
		next := node.Next()
		fmt.Println(node.Value())
		node = next
	}

	// Output:
	// 1
	// 2
	// 3
}

func ExampleList_PushBackList() {
	l := linked.NewList(1, 2)

	l.PushBackList(linked.NewList(3, 4))
	l.PushFrontList(linked.NewList(-1, 0))

	fmt.Println(l.ToSlice())
	// Output: [-1 0 1 2 3 4]
}
