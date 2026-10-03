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
	m := NewMap[K, V]()
	for k, v := range seq {
		m.Store(k, v)
	}
	return m
}

// NewAccessOrderedMap returns a new access-ordered map.
func NewAccessOrderedMap[K comparable, V any]() *Map[K, V] {
	m := NewMap[K, V]()
	m.accessOrder = true
	return m
}

// NewAccessOrderedMapSeq returns a new access-ordered map holding the entries from seq.
func NewAccessOrderedMapSeq[K comparable, V any](seq iter.Seq2[K, V]) *Map[K, V] {
	m := NewAccessOrderedMap[K, V]()
	for k, v := range seq {
		m.Store(k, v)
	}
	return m
}

// Load returns the value stored for key and reports whether the key was
// present. In an access-ordered map Load reorders the entries,
// so it modifies the map despite reading a value.
func (m *Map[K, V]) Load(key K) (value V, loaded bool) {
	m.init()

	if existing, ok := m.entries[key]; ok {
		value = existing.value
		loaded = true

		if m.accessOrder {
			m.keys.MoveToFront(existing.node)
		}
	}

	return
}

// Store sets the value for key and returns the value it replaced. In an
// insertion-ordered map a new key is appended and an existing one keeps its
// position; in an access-ordered map the key becomes the most recently used
// entry either way.
func (m *Map[K, V]) Store(key K, value V) (previous V, replaced bool) {
	m.init()

	if existing, ok := m.entries[key]; ok {
		m.entries[key] = mapEntry[K, V]{
			value: value,
			node:  existing.node,
		}

		if m.accessOrder {
			m.keys.MoveToFront(existing.node)
		}

		return existing.value, true
	}

	var node *Node[K]
	if m.accessOrder {
		node = m.keys.PushFront(key)
	} else {
		node = m.keys.PushBack(key)
	}

	m.entries[key] = mapEntry[K, V]{
		value: value,
		node:  node,
	}

	return
}

// Delete removes key and reports whether it was present.
func (m *Map[K, V]) Delete(key K) (deleted bool) {
	m.init()

	if existing, ok := m.entries[key]; ok {
		m.delete(existing.node)
		deleted = true
	}

	return
}

// DeleteFunc removes every entry for which del returns true, calling del
// in the map's order, and returns the number of entries removed.
func (m *Map[K, V]) DeleteFunc(del func(key K, value V) bool) (deleted int) {
	m.init()

	for k := range m.Keys() {
		v := m.entries[k]
		if del(k, v.value) {
			m.delete(v.node)
			deleted++
		}
	}

	return
}

// DeleteFirst removes the first entry according to the map's order - the
// entry stored first, or the most recently used one in an access-ordered
// map. It reports false with the zero value if the map is empty.
func (m *Map[K, V]) DeleteFirst() (value V, deleted bool) {
	m.init()

	first := m.keys.Front()
	if first == nil {
		return
	}

	if existing, ok := m.entries[first.Value()]; ok {
		value = existing.value
		deleted = true
	}
	m.delete(first)

	return
}

// DeleteLast removes the last entry according to the map's order - the
// entry stored last, or the least recently used one, which makes it the
// eviction target of an LRU cache. It reports false with the zero value if
// the map is empty.
func (m *Map[K, V]) DeleteLast() (value V, deleted bool) {
	m.init()

	back := m.keys.Back()
	if back == nil {
		return
	}

	if existing, ok := m.entries[back.Value()]; ok {
		value = existing.value
		deleted = true
	}
	m.delete(back)

	return
}

// Clear removes all entries from the map.
func (m *Map[K, V]) Clear() {
	m.init()
	clear(m.entries)
	m.keys.Clear()
}

// Len returns the number of entries in the map.
func (m *Map[K, V]) Len() int {
	m.init()
	return m.keys.Len()
}

// IsEmpty reports whether the map has no entries.
func (m *Map[K, V]) IsEmpty() bool {
	m.init()
	return m.Len() == 0
}

// All returns an iterator over the entries in the map's order.
// The map may be modified during iteration: an entry deleted
// before it is reached is skipped, and one stored during iteration is
// visited.
func (m *Map[K, V]) All() iter.Seq2[K, V] {
	m.init()

	return func(yield func(K, V) bool) {
		for k := range m.Keys() {
			v := m.entries[k]
			if !yield(k, v.value) {
				return
			}
		}
	}
}

// Keys returns an iterator over the keys in the map's order.
func (m *Map[K, V]) Keys() iter.Seq[K] {
	m.init()
	return m.keys.All()
}

// Values returns an iterator over the values in the map's order.
func (m *Map[K, V]) Values() iter.Seq[V] {
	m.init()

	return func(yield func(V) bool) {
		for _, v := range m.All() {
			if !yield(v) {
				return
			}
		}
	}
}

func (m *Map[K, V]) init() {
	if m.entries == nil || m.keys == nil {
		m.entries = make(map[K]mapEntry[K, V])
		m.keys = NewList[K]()
	}
}

func (m *Map[K, V]) delete(node *Node[K]) {
	m.keys.Remove(node)
	delete(m.entries, node.Value())
}

type mapEntry[K comparable, V any] struct {
	value V
	node  *Node[K]
}
