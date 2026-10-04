package linked_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AndreySenov/meridian/v3/linked"
)

// copyList returns a shallow copy of l made through reflection, which the
// copylocks vet check does not see: the copy is exactly what List
// guards against, and these tests exercise that guard.
func copyList(l *linked.List[int]) *linked.List[int] {
	c := linked.NewList[int]()
	reflect.ValueOf(c).Elem().Set(reflect.ValueOf(l).Elem())
	return c
}

func TestList(t *testing.T) {
	t.Run("Zero value is an empty list", func(t *testing.T) {
		var l linked.List[int]

		require.True(t, l.IsEmpty())
		require.Zero(t, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
		require.Empty(t, l.ToSlice())
		require.Empty(t, l.ToSliceBackward())
	})

	t.Run("NewListSeq collects an iterator in order", func(t *testing.T) {
		l := linked.NewListSeq(slices.Values([]int{1, 2, 3}))
		require.Equal(t, []int{1, 2, 3}, l.ToSlice())
		require.Equal(t, 3, l.Len())

		require.Equal(t, []int{3, 2, 1}, linked.NewListSeq(l.Backward()).ToSlice())

		require.True(t, linked.NewListSeq(slices.Values([]int{})).IsEmpty())
	})

	t.Run("PushBack appends in order", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)

		require.Equal(t, []int{1, 2, 3}, l.ToSlice())
		require.Equal(t, []int{3, 2, 1}, l.ToSliceBackward())
		require.Equal(t, 3, l.Len())
		require.False(t, l.IsEmpty())
		require.Equal(t, 1, l.Front().Value())
		require.Equal(t, 3, l.Back().Value())
	})

	t.Run("PushFront prepends in order", func(t *testing.T) {
		var l linked.List[int]
		l.PushFront(1)
		l.PushFront(2)
		l.PushFront(3)

		require.Equal(t, []int{3, 2, 1}, l.ToSlice())
		require.Equal(t, []int{1, 2, 3}, l.ToSliceBackward())
		require.Equal(t, 3, l.Front().Value())
		require.Equal(t, 1, l.Back().Value())
	})

	t.Run("PopFront and PopBack take values from the ends", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)

		v, ok := l.PopFront()
		require.True(t, ok)
		require.Equal(t, 1, v)

		v, ok = l.PopBack()
		require.True(t, ok)
		require.Equal(t, 4, v)

		require.Equal(t, []int{2, 3}, l.ToSlice())
		require.Equal(t, 2, l.Len())
	})

	t.Run("PopFront and PopBack report an empty list", func(t *testing.T) {
		l := linked.NewList(1)

		_, ok := l.PopBack()
		require.True(t, ok)
		require.True(t, l.IsEmpty())

		v, ok := l.PopFront()
		require.False(t, ok)
		require.Zero(t, v)

		v, ok = l.PopBack()
		require.False(t, ok)
		require.Zero(t, v)
	})

	t.Run("A single node is both front and back", func(t *testing.T) {
		var l linked.List[int]
		node := l.PushFront(1)

		require.Same(t, node, l.Front())
		require.Same(t, node, l.Back())

		l.PushBack(2)
		require.Equal(t, []int{1, 2}, l.ToSlice())
	})

	t.Run("Nodes link to their neighbours", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		first, second, third := l.Front(), l.Front().Next(), l.Back()

		require.Nil(t, first.Previous())
		require.Same(t, second, first.Next())
		require.Same(t, first, second.Previous())
		require.Same(t, third, second.Next())
		require.Same(t, second, third.Previous())
		require.Nil(t, third.Next())
	})

	t.Run("InsertBefore and InsertAfter", func(t *testing.T) {
		l := linked.NewList(2)

		front := l.InsertBefore(1, l.Front())
		back := l.InsertAfter(3, l.Back())
		l.InsertAfter(4, back)
		l.InsertBefore(0, front)

		require.Equal(t, []int{0, 1, 2, 3, 4}, l.ToSlice())
		require.Equal(t, []int{4, 3, 2, 1, 0}, l.ToSliceBackward())
		require.Equal(t, 5, l.Len())
		require.Equal(t, 0, l.Front().Value())
		require.Equal(t, 4, l.Back().Value())
	})

	t.Run("Insert relative to a nil or foreign mark is rejected", func(t *testing.T) {
		l := linked.NewList(1)
		other := linked.NewList(9)

		require.Nil(t, l.InsertBefore(0, nil))
		require.Nil(t, l.InsertAfter(0, nil))
		require.Nil(t, l.InsertBefore(0, other.Front()))
		require.Nil(t, l.InsertAfter(0, other.Front()))
		require.Equal(t, []int{1}, l.ToSlice())
		require.Equal(t, 1, l.Len())
	})

	t.Run("Remove from the front, the back, and the middle", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)

		l.Remove(l.Front())
		require.Equal(t, []int{2, 3, 4}, l.ToSlice())

		l.Remove(l.Back())
		require.Equal(t, []int{2, 3}, l.ToSlice())

		l.Remove(l.Front().Next())
		require.Equal(t, []int{2}, l.ToSlice())
		require.Equal(t, 1, l.Len())
	})

	t.Run("Removing the only node empties the list", func(t *testing.T) {
		var l linked.List[int]
		node := l.PushBack(1)

		l.Remove(node)

		require.True(t, l.IsEmpty())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		l.PushBack(2)
		require.Equal(t, []int{2}, l.ToSlice())
		require.Same(t, l.Front(), l.Back())
	})

	t.Run("A removed node is detached", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		node := l.Front().Next()

		l.Remove(node)

		require.Nil(t, node.Previous())
		require.Nil(t, node.Next())
		require.Equal(t, 2, node.Value())
	})

	t.Run("Removing twice or removing a foreign node is a no-op", func(t *testing.T) {
		l := linked.NewList(1, 2)
		node := l.Front()

		l.Remove(node)
		l.Remove(node)
		l.Remove(linked.NewList(9).Front())
		l.Remove(nil)

		require.Equal(t, []int{2}, l.ToSlice())
		require.Equal(t, 1, l.Len())
	})

	t.Run("Clear empties the list and detaches every node", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		first, last := l.Front(), l.Back()

		l.Clear()

		require.True(t, l.IsEmpty())
		require.Zero(t, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
		require.Empty(t, l.ToSlice())

		//no-op
		l.Remove(first)
		l.MoveToFront(last)
		require.Zero(t, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.InsertAfter(0, last))

		l.PushBack(4)
		require.Equal(t, []int{4}, l.ToSlice())
	})

	t.Run("Clear on an empty list is a no-op", func(t *testing.T) {
		var l linked.List[int]

		require.NotPanics(t, l.Clear)
		require.True(t, l.IsEmpty())
	})

	t.Run("A copy of a used list panics when used", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		c := copyList(l)
		node := l.Front()

		const msg = "List must not be copied after first use"
		require.PanicsWithValue(t, msg, func() { c.PushBack(4) })
		require.PanicsWithValue(t, msg, func() { c.PushFront(0) })
		require.PanicsWithValue(t, msg, func() { c.PopFront() })
		require.PanicsWithValue(t, msg, func() { c.PopBack() })
		require.PanicsWithValue(t, msg, func() { c.PushBackList(l) })
		require.PanicsWithValue(t, msg, func() { c.PushFrontList(l) })
		require.PanicsWithValue(t, msg, func() { c.Remove(node) })
		require.PanicsWithValue(t, msg, func() { c.Clear() })
		require.PanicsWithValue(t, msg, func() { c.MoveToFront(node) })
		require.PanicsWithValue(t, msg, func() { c.MoveToBack(node) })
		require.PanicsWithValue(t, msg, func() { c.MoveBefore(node, node) })
		require.PanicsWithValue(t, msg, func() { c.MoveAfter(node, node) })
		require.PanicsWithValue(t, msg, func() { c.InsertBefore(0, node) })
		require.PanicsWithValue(t, msg, func() { c.InsertAfter(0, node) })
		require.PanicsWithValue(t, msg, func() { c.ToSlice() })
		require.PanicsWithValue(t, msg, func() { c.ToSliceBackward() })
		require.PanicsWithValue(t, msg, func() {
			for range c.Nodes() {
			}
		})
		require.PanicsWithValue(t, msg, func() {
			for range c.NodesBackward() {
			}
		})

		require.PanicsWithValue(t, msg, func() { c.Front() })
		require.PanicsWithValue(t, msg, func() { c.Back() })
		require.PanicsWithValue(t, msg, func() { c.Len() })
		require.PanicsWithValue(t, msg, func() { c.IsEmpty() })

		require.Equal(t, []int{1, 2, 3}, l.ToSlice())
	})

	t.Run("A list may be copied while it holds no nodes", func(t *testing.T) {
		c := copyList(linked.NewList[int]())
		c.PushBack(1)
		require.Equal(t, []int{1}, c.ToSlice())

		used := linked.NewList(1, 2)
		used.Clear()
		d := copyList(used)
		d.PushBack(3)
		require.Equal(t, []int{3}, d.ToSlice())
		require.True(t, used.IsEmpty())
	})

	t.Run("MoveToFront and MoveToBack", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		middle := l.Front().Next()

		l.MoveToFront(middle)
		require.Equal(t, []int{2, 1, 3}, l.ToSlice())

		l.MoveToBack(middle)
		require.Equal(t, []int{1, 3, 2}, l.ToSlice())
		require.Equal(t, []int{2, 3, 1}, l.ToSliceBackward())
		require.Equal(t, 3, l.Len())
	})

	t.Run("Moving a node already at the target end or a foreign node is a no-op", func(t *testing.T) {
		l := linked.NewList(1, 2)
		other := linked.NewList(9)

		l.MoveToFront(l.Front())
		l.MoveToBack(l.Back())
		l.MoveToFront(other.Front())
		l.MoveToBack(other.Front())

		require.Equal(t, []int{1, 2}, l.ToSlice())
		require.Equal(t, []int{9}, other.ToSlice())
	})

	t.Run("Moving the only node keeps the list intact", func(t *testing.T) {
		var l linked.List[int]
		node := l.PushBack(1)

		l.MoveToFront(node)
		l.MoveToBack(node)

		require.Equal(t, []int{1}, l.ToSlice())
		require.Same(t, node, l.Front())
		require.Same(t, node, l.Back())
	})

	t.Run("MoveBefore and MoveAfter reposition a node", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)
		first, second, third, fourth := l.Front(), l.Front().Next(), l.Back().Previous(), l.Back()

		l.MoveBefore(fourth, second)
		require.Equal(t, []int{1, 4, 2, 3}, l.ToSlice())

		l.MoveAfter(first, third)
		require.Equal(t, []int{4, 2, 3, 1}, l.ToSlice())
		require.Same(t, fourth, l.Front())
		require.Same(t, first, l.Back())

		l.MoveBefore(first, fourth)
		require.Equal(t, []int{1, 4, 2, 3}, l.ToSlice())
		require.Same(t, first, l.Front())
		require.Same(t, third, l.Back())

		require.Equal(t, []int{3, 2, 4, 1}, l.ToSliceBackward())
		require.Equal(t, 4, l.Len())
	})

	t.Run("Moving a node next to its current neighbour keeps the order", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		first, second := l.Front(), l.Front().Next()

		l.MoveBefore(first, second)
		l.MoveAfter(second, first)

		require.Equal(t, []int{1, 2, 3}, l.ToSlice())
		require.Equal(t, []int{3, 2, 1}, l.ToSliceBackward())
	})

	t.Run("MoveBefore and MoveAfter with the node itself, a foreign node, or nil are no-ops", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		middle := l.Front().Next()
		foreign := linked.NewList(9).Front()

		l.MoveBefore(middle, middle)
		l.MoveAfter(middle, middle)
		l.MoveBefore(middle, foreign)
		l.MoveAfter(foreign, middle)
		l.MoveBefore(middle, nil)
		l.MoveAfter(nil, middle)

		require.Equal(t, []int{1, 2, 3}, l.ToSlice())
		require.Equal(t, 3, l.Len())
		require.Same(t, middle, l.Front().Next())
		require.NotSame(t, middle, middle.Next())

		var single linked.List[int]
		only := single.PushBack(1)
		single.MoveAfter(only, only)
		require.Equal(t, []int{1}, single.ToSlice())
		require.Same(t, only, single.Front())
	})

	t.Run("PushBackList and PushFrontList append another list", func(t *testing.T) {
		l := linked.NewList(3, 4)

		l.PushBackList(linked.NewList(5, 6))
		l.PushFrontList(linked.NewList(1, 2))

		require.Equal(t, []int{1, 2, 3, 4, 5, 6}, l.ToSlice())
		require.Equal(t, 6, l.Len())
	})

	t.Run("PushBackList and PushFrontList tolerate nil and empty lists", func(t *testing.T) {
		l := linked.NewList(1)

		l.PushBackList(nil)
		l.PushFrontList(nil)
		l.PushBackList(linked.NewList[int]())
		l.PushFrontList(linked.NewList[int]())

		require.Equal(t, []int{1}, l.ToSlice())
	})

	t.Run("A list can be appended to itself", func(t *testing.T) {
		l := linked.NewList(1, 2)
		l.PushBackList(l)
		require.Equal(t, []int{1, 2, 1, 2}, l.ToSlice())

		l = linked.NewList(2, 1)
		l.PushFrontList(l)
		require.Equal(t, []int{2, 1, 2, 1}, l.ToSlice())
		require.Equal(t, 4, l.Len())
	})

	t.Run("Iteration stops early on break", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)

		var seen []int
		for v := range l.All() {
			seen = append(seen, v)
			if v == 2 {
				break
			}
		}
		require.Equal(t, []int{1, 2}, seen)

		seen = nil
		for v := range l.Backward() {
			seen = append(seen, v)
			break
		}
		require.Equal(t, []int{3}, seen)
	})

	t.Run("Iteration survives removing the current node", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)
		nodes := map[int]*linked.Node[int]{}
		for n := range l.Nodes() {
			nodes[n.Value()] = n
		}

		var seen []int
		for v := range l.All() {
			seen = append(seen, v)
			l.Remove(nodes[v])
		}
		require.Equal(t, []int{1, 2, 3, 4}, seen)
		require.True(t, l.IsEmpty())

		l = linked.NewList(1, 2, 3, 4)
		nodes = map[int]*linked.Node[int]{}
		for n := range l.Nodes() {
			nodes[n.Value()] = n
		}

		seen = nil
		for v := range l.Backward() {
			seen = append(seen, v)
			l.Remove(nodes[v])
		}
		require.Equal(t, []int{4, 3, 2, 1}, seen)
		require.True(t, l.IsEmpty())
	})

	t.Run("Iteration reflects removals and insertions ahead of the current node", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)
		second := l.Front().Next()

		var seen []int
		for v := range l.All() {
			seen = append(seen, v)
			if v == 1 {
				l.Remove(second)
				l.InsertAfter(9, l.Front())
			}
		}
		require.Equal(t, []int{1, 9, 3, 4}, seen)

		l = linked.NewList(1, 2, 3, 4)
		third := l.Back().Previous()

		seen = nil
		for v := range l.Backward() {
			seen = append(seen, v)
			if v == 4 {
				l.Remove(third)
				l.InsertBefore(9, l.Back())
			}
		}
		require.Equal(t, []int{4, 9, 2, 1}, seen)
	})

	t.Run("Iteration never yields a node removed together with the current one", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4)
		first, second := l.Front(), l.Front().Next()

		var seen []int
		for v := range l.All() {
			seen = append(seen, v)
			if v == 1 {
				l.Remove(first)
				l.Remove(second)
			}
		}
		require.Equal(t, []int{1}, seen)

		l = linked.NewList(1, 2, 3, 4)
		last, penultimate := l.Back(), l.Back().Previous()

		seen = nil
		for v := range l.Backward() {
			seen = append(seen, v)
			if v == 4 {
				l.Remove(last)
				l.Remove(penultimate)
			}
		}
		require.Equal(t, []int{4}, seen)
	})

	t.Run("Nodes yields the nodes themselves in both directions", func(t *testing.T) {
		l := linked.NewList(1, 2, 3)
		first, second, third := l.Front(), l.Front().Next(), l.Back()

		var nodes []*linked.Node[int]
		for n := range l.Nodes() {
			nodes = append(nodes, n)
		}
		require.Equal(t, []*linked.Node[int]{first, second, third}, nodes)

		nodes = nil
		for n := range l.NodesBackward() {
			nodes = append(nodes, n)
		}
		require.Equal(t, []*linked.Node[int]{third, second, first}, nodes)
	})

	t.Run("The list stays consistent after removing nodes during iteration", func(t *testing.T) {
		l := linked.NewList(1, 2, 3, 4, 5)
		for n := range l.Nodes() {
			if n.Value()%2 == 0 {
				l.Remove(n)
			}
		}

		require.Equal(t, []int{1, 3, 5}, l.ToSlice())
		require.Equal(t, []int{5, 3, 1}, l.ToSliceBackward())
		require.Equal(t, 3, l.Len())
		require.Equal(t, 1, l.Front().Value())
		require.Equal(t, 5, l.Back().Value())

		for n := range l.NodesBackward() {
			if n.Value() != 3 {
				l.Remove(n)
			}
		}

		require.Equal(t, []int{3}, l.ToSlice())
		require.Same(t, l.Front(), l.Back())
	})
}
