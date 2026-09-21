package linked

import "iter"

// Map is a map that iterates over its entries in insertion order.
// The zero value for Map is an empty map ready to use. Like a map,
// a Map is a reference to shared data once in use: copies of it
// refer to the same entries.
// A Map is not safe for concurrent use.
type Map[K comparable, V any] struct {
	entries map[K]mapEntry[K, V]
	keys    *List[K]
}

// NewMapSeq returns a new map holding the entries from seq.
func NewMapSeq[K comparable, V any](seq iter.Seq2[K, V]) *Map[K, V] {
	l := new(Map[K, V])
	for k, v := range seq {
		l.Store(k, v)
	}
	return l
}

// Load returns the value stored for key and reports whether the key was
// present.
func (l *Map[K, V]) Load(key K) (value V, loaded bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		value = existing.value
		loaded = true
	}

	return
}

// Store sets the value for key; an existing key keeps its position,
// and Store returns the value it replaced.
func (l *Map[K, V]) Store(key K, value V) (previous V, replaced bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		l.entries[key] = mapEntry[K, V]{
			value: value,
			node:  existing.node,
		}
		return existing.value, true
	}

	l.entries[key] = mapEntry[K, V]{
		value: value,
		node:  l.keys.PushBack(key),
	}

	return
}

// Delete removes key and reports whether it was present.
func (l *Map[K, V]) Delete(key K) (deleted bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		l.keys.Remove(existing.node)
		delete(l.entries, key)
		deleted = true
	}

	return
}

// DeleteFunc removes every entry for which del returns true, calling del in
// iteration order, and returns the number of entries removed.
func (l *Map[K, V]) DeleteFunc(del func(key K, value V) bool) (deleted int) {
	l.init()

	for k := range l.Keys() {
		v := l.entries[k]
		if del(k, v.value) {
			delete(l.entries, k)
			l.keys.Remove(v.node)
			deleted++
		}
	}

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

// All returns an iterator over the entries in insertion order. The map may
// be modified during iteration: an entry deleted before it is reached is skipped,
// and one stored during iteration is visited.
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

// Keys returns an iterator over the keys in insertion order.
func (l *Map[K, V]) Keys() iter.Seq[K] {
	l.init()
	return l.keys.All()
}

// Values returns an iterator over the values in insertion order.
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
	if l.entries == nil {
		l.entries = make(map[K]mapEntry[K, V])
		l.keys = new(List[K])
	}
}

type mapEntry[K comparable, V any] struct {
	value V
	node  *Node[K]
}
