package meridian

import (
	"iter"
	"slices"

	"github.com/AndreySenov/meridian/internal"
)

// LinkedList represents a doubly linked list.
// It is a typed alternative to [container/list.List].
// The zero value for LinkedList is an empty list ready to use.
// A LinkedList must not be copied after first use, and it is not safe for
// concurrent use.
type LinkedList[E any] struct {
	noCopy     internal.NoCopy
	head, tail *LinkedListNode[E]
	len        int
}

// NewLinkedList returns a new list holding values in the given order.
func NewLinkedList[E any](values ...E) *LinkedList[E] {
	l := new(LinkedList[E])
	for _, v := range values {
		l.PushBack(v)
	}
	return l
}

// NewLinkedListSeq returns a new list holding values from seq.
func NewLinkedListSeq[E any](seq iter.Seq[E]) *LinkedList[E] {
	l := new(LinkedList[E])
	for v := range seq {
		l.PushBack(v)
	}
	return l
}

// Front returns the first node, or nil if the list is empty.
func (l *LinkedList[E]) Front() *LinkedListNode[E] {
	l.check()
	return l.head
}

// Back returns the last node, or nil if the list is empty.
func (l *LinkedList[E]) Back() *LinkedListNode[E] {
	l.check()
	return l.tail
}

// Len returns the number of elements in the list.
func (l *LinkedList[E]) Len() int {
	l.check()
	return l.len
}

// IsEmpty reports whether the list has no elements.
func (l *LinkedList[E]) IsEmpty() bool {
	return l.Len() == 0
}

// All returns an iterator over the values from front to back. It follows
// the same rules under mutation as [LinkedList.Nodes].
func (l *LinkedList[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		for node := range l.Nodes() {
			if !yield(node.value) {
				return
			}
		}
	}
}

// Backward returns an iterator over the values from back to front. It
// follows the same rules under mutation as [LinkedList.NodesBackward].
func (l *LinkedList[E]) Backward() iter.Seq[E] {
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
func (l *LinkedList[E]) Nodes() iter.Seq[*LinkedListNode[E]] {
	l.check()
	return func(yield func(*LinkedListNode[E]) bool) {
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
// under the same rules for modification as [LinkedList.Nodes].
func (l *LinkedList[E]) NodesBackward() iter.Seq[*LinkedListNode[E]] {
	l.check()
	return func(yield func(*LinkedListNode[E]) bool) {
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
func (l *LinkedList[E]) ToSlice() []E {
	return slices.Collect(l.All())
}

// ToSliceBackward returns the values from back to front, or nil if the
// list is empty.
func (l *LinkedList[E]) ToSliceBackward() []E {
	return slices.Collect(l.Backward())
}

// PushFront inserts value at the front and returns its node.
func (l *LinkedList[E]) PushFront(value E) *LinkedListNode[E] {
	l.check()
	node := l.newLinkedListNode(value)
	l.linkHead(node)
	l.len++
	return node
}

// PushBack inserts value at the back and returns its node.
func (l *LinkedList[E]) PushBack(value E) *LinkedListNode[E] {
	l.check()
	node := l.newLinkedListNode(value)
	l.linkTail(node)
	l.len++
	return node
}

// PopFront removes the first element and returns its value. It reports
// false, with the zero value, if the list is empty.
func (l *LinkedList[E]) PopFront() (value E, ok bool) {
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
func (l *LinkedList[E]) PopBack() (value E, ok bool) {
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
func (l *LinkedList[E]) PushFrontList(other *LinkedList[E]) {
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
func (l *LinkedList[E]) PushBackList(other *LinkedList[E]) {
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
func (l *LinkedList[E]) Remove(target *LinkedListNode[E]) {
	l.check()
	if l.retains(target) {
		l.unlink(target)
		l.len--
	}
}

// Clear removes every element. Nodes obtained before Clear are detached
// and no longer accepted by the list.
func (l *LinkedList[E]) Clear() {
	for node := range l.Nodes() {
		l.Remove(node)
	}
}

// MoveToFront moves target to the front of the list.
func (l *LinkedList[E]) MoveToFront(target *LinkedListNode[E]) {
	l.check()
	if l.retains(target) && l.head != target {
		l.unlink(target)
		l.linkHead(target)
	}
}

// MoveToBack moves target to the back of the list.
func (l *LinkedList[E]) MoveToBack(target *LinkedListNode[E]) {
	l.check()
	if l.retains(target) && l.tail != target {
		l.unlink(target)
		l.linkTail(target)
	}
}

// MoveBefore moves target to the position immediately before mark. It is a
// no-op if target and mark are the same node.
func (l *LinkedList[E]) MoveBefore(target *LinkedListNode[E], mark *LinkedListNode[E]) {
	l.check()
	if target != mark && l.retains(target) && l.retains(mark) {
		l.unlink(target)
		l.linkBefore(target, mark)
	}
}

// MoveAfter moves target to the position immediately after mark. It is a
// no-op if target and mark are the same node.
func (l *LinkedList[E]) MoveAfter(target *LinkedListNode[E], mark *LinkedListNode[E]) {
	l.check()
	if target != mark && l.retains(target) && l.retains(mark) {
		l.unlink(target)
		l.linkAfter(target, mark)
	}
}

// InsertBefore inserts value immediately before mark and returns its node,
// or nil if mark is not a node of the list.
func (l *LinkedList[E]) InsertBefore(value E, mark *LinkedListNode[E]) *LinkedListNode[E] {
	l.check()
	if l.retains(mark) {
		node := l.newLinkedListNode(value)
		l.linkBefore(node, mark)
		l.len++
		return node
	}
	return nil
}

// InsertAfter inserts value immediately after mark and returns its node,
// or nil if mark is not a node of the list.
func (l *LinkedList[E]) InsertAfter(value E, mark *LinkedListNode[E]) *LinkedListNode[E] {
	l.check()
	if l.retains(mark) {
		node := l.newLinkedListNode(value)
		l.linkAfter(node, mark)
		l.len++
		return node
	}
	return nil
}

func (l *LinkedList[E]) check() {
	if l.head != nil && l.head.list != l {
		panic("LinkedList must not be copied after first use")
	}
}

func (l *LinkedList[E]) retains(target *LinkedListNode[E]) bool {
	return target != nil && target.list == l
}

func (l *LinkedList[E]) linkHead(target *LinkedListNode[E]) {
	if target != nil {
		if l.head == nil {
			l.linkSole(target)
		} else {
			l.linkBefore(target, l.head)
		}
	}
}

func (l *LinkedList[E]) linkTail(target *LinkedListNode[E]) {
	if target != nil {
		if l.tail == nil {
			l.linkSole(target)
		} else {
			l.linkAfter(target, l.tail)
		}
	}
}

func (l *LinkedList[E]) linkSole(target *LinkedListNode[E]) {
	if target != nil {
		target.list = l
		l.head = target
		l.tail = target
	}
}

func (l *LinkedList[E]) linkBefore(target *LinkedListNode[E], mark *LinkedListNode[E]) {
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

func (l *LinkedList[E]) linkAfter(target *LinkedListNode[E], mark *LinkedListNode[E]) {
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

func (l *LinkedList[E]) unlink(target *LinkedListNode[E]) {
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

func (l *LinkedList[E]) newLinkedListNode(value E) *LinkedListNode[E] {
	return &LinkedListNode[E]{
		value: value,
	}
}

// LinkedListNode is an element of a [LinkedList]. Nodes are created by the
// list and belong to it until removed, after which they belong to no list.
// A list ignores nodes that are nil or not its own: such operations are no-ops.
type LinkedListNode[E any] struct {
	value          E
	previous, next *LinkedListNode[E]
	list           *LinkedList[E]
}

// Value returns the element's value.
func (l *LinkedListNode[E]) Value() E {
	return l.value
}

// Previous returns the preceding node, or nil if this is the first node or
// the node has been removed from its list.
func (l *LinkedListNode[E]) Previous() *LinkedListNode[E] {
	return l.previous
}

// Next returns the following node, or nil if this is the last node or the
// node has been removed from its list.
func (l *LinkedListNode[E]) Next() *LinkedListNode[E] {
	return l.next
}
