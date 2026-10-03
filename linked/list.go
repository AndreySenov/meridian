package linked

import (
	"iter"
	"slices"

	"github.com/AndreySenov/meridian/v2/internal"
)

// List represents a doubly linked list.
// It is a typed alternative to [container/list.List].
// The zero value for List is an empty list ready to use.
// A List must not be copied after first use, and it is not safe for
// concurrent use.
type List[E any] struct {
	noCopy     internal.NoCopy
	head, tail *Node[E]
	len        int
}

// NewList returns a new list holding values in the given order.
func NewList[E any](values ...E) *List[E] {
	l := new(List[E])
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

// NewListSeq returns a new list holding values from seq.
func NewListSeq[E any](seq iter.Seq[E]) *List[E] {
	l := NewList[E]()
	for v := range seq {
		l.PushBack(v)
	}
	return l
}

// Front returns the first node, or nil if the list is empty.
func (l *List[E]) Front() *Node[E] {
	l.check()
	return l.head
}

// Back returns the last node, or nil if the list is empty.
func (l *List[E]) Back() *Node[E] {
	l.check()
	return l.tail
}

// Len returns the number of elements in the list.
func (l *List[E]) Len() int {
	l.check()
	return l.len
}

// IsEmpty reports whether the list has no elements.
func (l *List[E]) IsEmpty() bool {
	return l.Len() == 0
}

// All returns an iterator over the values from front to back. It follows
// the same rules under mutation as [List.Nodes].
func (l *List[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		for node := range l.Nodes() {
			if !yield(node.value) {
				return
			}
		}
	}
}

// Backward returns an iterator over the values from back to front. It
// follows the same rules under mutation as [List.NodesBackward].
func (l *List[E]) Backward() iter.Seq[E] {
	return func(yield func(E) bool) {
		for node := range l.NodesBackward() {
			if !yield(node.value) {
				return
			}
		}
	}
}

// Nodes returns an iterator over the nodes from front to back.
//
// The list may be modified during iteration: any node may be removed,
// including the current one, and nodes may be inserted; a removed node is
// never yielded, and a node inserted ahead of the current one is. Moving
// the current node changes where the iteration continues from.
func (l *List[E]) Nodes() iter.Seq[*Node[E]] {
	l.check()
	return func(yield func(*Node[E]) bool) {
		for node := l.head; node != nil; {
			next := node.next
			if !yield(node) {
				return
			}
			if l.retains(node) {
				node = node.next
			} else if l.retains(next) {
				node = next
			} else {
				return
			}
		}
	}
}

// NodesBackward returns an iterator over the nodes from back to front,
// under the same rules for modification as [List.Nodes].
func (l *List[E]) NodesBackward() iter.Seq[*Node[E]] {
	l.check()
	return func(yield func(*Node[E]) bool) {
		for node := l.tail; node != nil; {
			previous := node.previous
			if !yield(node) {
				return
			}
			if l.retains(node) {
				node = node.previous
			} else if l.retains(previous) {
				node = previous
			} else {
				return
			}
		}
	}
}

// ToSlice returns the values from front to back, or nil if the list is
// empty.
func (l *List[E]) ToSlice() []E {
	return slices.Collect(l.All())
}

// ToSliceBackward returns the values from back to front, or nil if the
// list is empty.
func (l *List[E]) ToSliceBackward() []E {
	return slices.Collect(l.Backward())
}

// PushFront inserts value at the front and returns its node.
func (l *List[E]) PushFront(value E) *Node[E] {
	l.check()
	node := l.newNode(value)
	l.linkHead(node)
	l.len++
	return node
}

// PushBack inserts value at the back and returns its node.
func (l *List[E]) PushBack(value E) *Node[E] {
	l.check()
	node := l.newNode(value)
	l.linkTail(node)
	l.len++
	return node
}

// PopFront removes the first element and returns its value. It reports
// false, with the zero value, if the list is empty.
func (l *List[E]) PopFront() (value E, ok bool) {
	l.check()
	node := l.Front()
	if node == nil {
		return
	}
	l.Remove(node)
	return node.value, true
}

// PopBack removes the last element and returns its value. It reports
// false, with the zero value, if the list is empty.
func (l *List[E]) PopBack() (value E, ok bool) {
	l.check()
	node := l.Back()
	if node == nil {
		return
	}
	l.Remove(node)
	return node.value, true
}

// PushFrontList inserts copies of the values of other at the front, keeping
// their order, so that other's first value becomes the first value of l.
// Other is left unchanged and may be l itself; nil is a no-op.
func (l *List[E]) PushFrontList(other *List[E]) {
	l.check()
	if other != nil {
		length := other.len
		pushed := 0
		for value := range other.Backward() {
			if pushed == length {
				break
			}
			l.PushFront(value)
			pushed++
		}
	}
}

// PushBackList inserts copies of the values of other at the back, keeping
// their order. Other is left unchanged and may be l itself; nil is a no-op.
func (l *List[E]) PushBackList(other *List[E]) {
	l.check()
	if other != nil {
		length := other.len
		pushed := 0
		for value := range other.All() {
			if pushed == length {
				break
			}
			l.PushBack(value)
			pushed++
		}
	}
}

// Remove removes target from the list. The node is detached afterwards:
// its Previous and Next return nil, and the list treats it as foreign, so
// removing it again is a no-op. Its Value stays readable.
func (l *List[E]) Remove(target *Node[E]) {
	l.check()
	if l.retains(target) {
		l.unlink(target)
		l.len--
	}
}

// Clear removes every element. Nodes obtained before Clear are detached
// and no longer accepted by the list.
func (l *List[E]) Clear() {
	for node := range l.Nodes() {
		l.Remove(node)
	}
}

// MoveToFront moves target to the front of the list.
func (l *List[E]) MoveToFront(target *Node[E]) {
	l.check()
	if l.retains(target) && l.head != target {
		l.unlink(target)
		l.linkHead(target)
	}
}

// MoveToBack moves target to the back of the list.
func (l *List[E]) MoveToBack(target *Node[E]) {
	l.check()
	if l.retains(target) && l.tail != target {
		l.unlink(target)
		l.linkTail(target)
	}
}

// MoveBefore moves target to the position immediately before mark. It is a
// no-op if target and mark are the same node.
func (l *List[E]) MoveBefore(target *Node[E], mark *Node[E]) {
	l.check()
	if target != mark && l.retains(target) && l.retains(mark) {
		l.unlink(target)
		l.linkBefore(target, mark)
	}
}

// MoveAfter moves target to the position immediately after mark. It is a
// no-op if target and mark are the same node.
func (l *List[E]) MoveAfter(target *Node[E], mark *Node[E]) {
	l.check()
	if target != mark && l.retains(target) && l.retains(mark) {
		l.unlink(target)
		l.linkAfter(target, mark)
	}
}

// InsertBefore inserts value immediately before mark and returns its node,
// or nil if mark is not a node of the list.
func (l *List[E]) InsertBefore(value E, mark *Node[E]) *Node[E] {
	l.check()
	if l.retains(mark) {
		node := l.newNode(value)
		l.linkBefore(node, mark)
		l.len++
		return node
	}
	return nil
}

// InsertAfter inserts value immediately after mark and returns its node,
// or nil if mark is not a node of the list.
func (l *List[E]) InsertAfter(value E, mark *Node[E]) *Node[E] {
	l.check()
	if l.retains(mark) {
		node := l.newNode(value)
		l.linkAfter(node, mark)
		l.len++
		return node
	}
	return nil
}

func (l *List[E]) check() {
	if l.head != nil && l.head.list != l {
		panic("List must not be copied after first use")
	}
}

func (l *List[E]) retains(target *Node[E]) bool {
	return target != nil && target.list == l
}

func (l *List[E]) linkHead(target *Node[E]) {
	if target != nil {
		if l.head == nil {
			l.linkSole(target)
		} else {
			l.linkBefore(target, l.head)
		}
	}
}

func (l *List[E]) linkTail(target *Node[E]) {
	if target != nil {
		if l.tail == nil {
			l.linkSole(target)
		} else {
			l.linkAfter(target, l.tail)
		}
	}
}

func (l *List[E]) linkSole(target *Node[E]) {
	if target != nil {
		target.list = l
		l.head = target
		l.tail = target
	}
}

func (l *List[E]) linkBefore(target *Node[E], mark *Node[E]) {
	if target != nil {
		target.list = l
		target.previous = mark.previous
		target.next = mark
		if mark.previous != nil {
			mark.previous.next = target
		}
		mark.previous = target
		if mark == l.head {
			l.head = target
		}
	}
}

func (l *List[E]) linkAfter(target *Node[E], mark *Node[E]) {
	if target != nil {
		target.list = l
		target.previous = mark
		target.next = mark.next
		if mark.next != nil {
			mark.next.previous = target
		}
		mark.next = target
		if mark == l.tail {
			l.tail = target
		}
	}
}

func (l *List[E]) unlink(target *Node[E]) {
	if target != nil {
		if target.previous != nil {
			target.previous.next = target.next
		}
		if target.next != nil {
			target.next.previous = target.previous
		}
		if l.head == target {
			l.head = target.next
		}
		if l.tail == target {
			l.tail = target.previous
		}
		target.previous = nil
		target.next = nil
		target.list = nil
	}
}

func (l *List[E]) newNode(value E) *Node[E] {
	return &Node[E]{
		value: value,
	}
}

// Node is an element of a [List]. Nodes are created by the
// list and belong to it until removed, after which they belong to no list.
// A list ignores nodes that are nil or not its own: such operations are no-ops.
type Node[E any] struct {
	value          E
	previous, next *Node[E]
	list           *List[E]
}

// Value returns the element's value.
func (l *Node[E]) Value() E {
	return l.value
}

// Previous returns the preceding node, or nil if this is the first node or
// the node has been removed from its list.
func (l *Node[E]) Previous() *Node[E] {
	return l.previous
}

// Next returns the following node, or nil if this is the last node or the
// node has been removed from its list.
func (l *Node[E]) Next() *Node[E] {
	return l.next
}
