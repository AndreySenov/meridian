package linked

import "iter"

// Map is a map that iterates over its entries in predictable order.
//
// An insertion-ordered map, returned by [NewMap], iterates from the entry
// stored first to the entry stored last.
// An access-ordered map, returned by [NewAccessOrderedMap], iterates
// from the most recently used entry to the least recently used one,
// where both Load and Store count as a use; maps with this order suit LRU caches.
//
// The zero value for Map is an empty insertion-ordered map ready to use.
// Like a map, a Map is a reference to shared data once in use: copies of it
// refer to the same entries.
// A Map is not safe for concurrent use.
type Map[K comparable, V any] struct {
	entries     map[K]mapEntry[K, V]
	keys        *List[K]
	accessOrder bool
}

// NewMap returns a new insertion-ordered map.
func NewMap[K comparable, V any]() *Map[K, V] {
	return new(Map[K, V])
}

// NewMapSeq returns a new insertion-ordered map holding the entries from seq.
func NewMapSeq[K comparable, V any](seq iter.Seq2[K, V]) *Map[K, V] {
	l := NewMap[K, V]()
	for k, v := range seq {
		l.Store(k, v)
	}
	return l
}

// NewAccessOrderedMap returns a new access-ordered map.
func NewAccessOrderedMap[K comparable, V any]() *Map[K, V] {
	l := NewMap[K, V]()
	l.accessOrder = true
	return l
}

// NewAccessOrderedMapSeq returns a new access-ordered map holding the entries from seq.
func NewAccessOrderedMapSeq[K comparable, V any](seq iter.Seq2[K, V]) *Map[K, V] {
	l := NewAccessOrderedMap[K, V]()
	for k, v := range seq {
		l.Store(k, v)
	}
	return l
}

// Load returns the value stored for key and reports whether the key was
// present. In an access-ordered map Load reorders the entries,
// so it modifies the map despite reading a value.
func (l *Map[K, V]) Load(key K) (value V, loaded bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		value = existing.value
		loaded = true

		if l.accessOrder {
			l.keys.MoveToFront(existing.node)
		}
	}

	return
}

// Store sets the value for key and returns the value it replaced. In an
// insertion-ordered map a new key is appended and an existing one keeps its
// position; in an access-ordered map the key becomes the most recently used
// entry either way.
func (l *Map[K, V]) Store(key K, value V) (previous V, replaced bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		l.entries[key] = mapEntry[K, V]{
			value: value,
			node:  existing.node,
		}

		if l.accessOrder {
			l.keys.MoveToFront(existing.node)
		}

		return existing.value, true
	}

	var node *Node[K]
	if l.accessOrder {
		node = l.keys.PushFront(key)
	} else {
		node = l.keys.PushBack(key)
	}

	l.entries[key] = mapEntry[K, V]{
		value: value,
		node:  node,
	}

	return
}

// Delete removes key and reports whether it was present.
func (l *Map[K, V]) Delete(key K) (deleted bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		l.delete(existing.node)
		deleted = true
	}

	return
}

// DeleteFunc removes every entry for which del returns true, calling del
// in the map's order, and returns the number of entries removed.
func (l *Map[K, V]) DeleteFunc(del func(key K, value V) bool) (deleted int) {
	l.init()

	for k := range l.Keys() {
		v := l.entries[k]
		if del(k, v.value) {
			l.delete(v.node)
			deleted++
		}
	}

	return
}

// DeleteFirst removes the first entry according to the map's order - the
// entry stored first, or the most recently used one in an access-ordered
// map. It reports false with the zero value if the map is empty.
func (l *Map[K, V]) DeleteFirst() (value V, deleted bool) {
	l.init()

	first := l.keys.Front()
	if first == nil {
		return
	}

	if existing, ok := l.entries[first.Value()]; ok {
		value = existing.value
		deleted = true
	}
	l.delete(first)

	return
}

// DeleteLast removes the last entry according to the map's order - the
// entry stored last, or the least recently used one, which makes it the
// eviction target of an LRU cache. It reports false with the zero value if
// the map is empty.
func (l *Map[K, V]) DeleteLast() (value V, deleted bool) {
	l.init()

	back := l.keys.Back()
	if back == nil {
		return
	}

	if existing, ok := l.entries[back.Value()]; ok {
		value = existing.value
		deleted = true
	}
	l.delete(back)

	return
}

// Clear removes all entries from the map.
func (l *Map[K, V]) Clear() {
	l.init()
	clear(l.entries)
	l.keys.Clear()
}

// Len returns the number of entries in the map.
func (l *Map[K, V]) Len() int {
	l.init()
	return l.keys.Len()
}

// IsEmpty reports whether the map has no entries.
func (l *Map[K, V]) IsEmpty() bool {
	l.init()
	return l.Len() == 0
}

// All returns an iterator over the entries in the map's order.
// The map may be modified during iteration: an entry deleted
// before it is reached is skipped, and one stored during iteration is
// visited.
func (l *Map[K, V]) All() iter.Seq2[K, V] {
	l.init()

	return func(yield func(K, V) bool) {
		for k := range l.Keys() {
			v := l.entries[k]
			if !yield(k, v.value) {
				return
			}
		}
	}
}

// Keys returns an iterator over the keys in the map's order.
func (l *Map[K, V]) Keys() iter.Seq[K] {
	l.init()
	return l.keys.All()
}

// Values returns an iterator over the values in the map's order.
func (l *Map[K, V]) Values() iter.Seq[V] {
	l.init()

	return func(yield func(V) bool) {
		for _, v := range l.All() {
			if !yield(v) {
				return
			}
		}
	}
}

func (l *Map[K, V]) init() {
	if l.entries == nil || l.keys == nil {
		l.entries = make(map[K]mapEntry[K, V])
		l.keys = new(List[K])
	}
}

func (l *Map[K, V]) delete(node *Node[K]) {
	l.keys.Remove(node)
	delete(l.entries, node.Value())
}

type mapEntry[K comparable, V any] struct {
	value V
	node  *Node[K]
}
