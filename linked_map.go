package meridian

import "iter"

// LinkedMap is a map that iterates over its entries in insertion order.
// The zero value for LinkedMap is an empty map ready to use. Like a map,
// a LinkedMap is a reference to shared data once in use: copies of it
// refer to the same entries.
// A LinkedMap is not safe for concurrent use.
type LinkedMap[K comparable, V any] struct {
	entries map[K]linkedMapEntry[K, V]
	keys    *LinkedList[K]
}

// NewLinkedMapSeq returns a new map holding the entries from seq.
func NewLinkedMapSeq[K comparable, V any](seq iter.Seq2[K, V]) *LinkedMap[K, V] {
	l := new(LinkedMap[K, V])
	for k, v := range seq {
		l.Store(k, v)
	}
	return l
}

// Load returns the value stored for key and reports whether the key was
// present.
func (l *LinkedMap[K, V]) Load(key K) (value V, loaded bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		value = existing.value
		loaded = true
	}

	return
}

// Store sets the value for key; an existing key keeps its position,
// and Store returns the value it replaced.
func (l *LinkedMap[K, V]) Store(key K, value V) (previous V, replaced bool) {
	l.init()

	if existing, ok := l.entries[key]; ok {
		l.entries[key] = linkedMapEntry[K, V]{
			value: value,
			node:  existing.node,
		}
		return existing.value, true
	}

	l.entries[key] = linkedMapEntry[K, V]{
		value: value,
		node:  l.keys.PushBack(key),
	}

	return
}

// Delete removes key and reports whether it was present.
func (l *LinkedMap[K, V]) Delete(key K) (deleted bool) {
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
func (l *LinkedMap[K, V]) DeleteFunc(del func(key K, value V) bool) (deleted int) {
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
func (l *LinkedMap[K, V]) Clear() {
	l.init()
	clear(l.entries)
	l.keys.Clear()
}

// Len returns the number of entries in the map.
func (l *LinkedMap[K, V]) Len() int {
	l.init()
	return l.keys.Len()
}

// IsEmpty reports whether the map has no entries.
func (l *LinkedMap[K, V]) IsEmpty() bool {
	l.init()
	return l.Len() == 0
}

// All returns an iterator over the entries in insertion order. The map may
// be modified during iteration: an entry deleted before it is reached is skipped,
// and one stored during iteration is visited.
func (l *LinkedMap[K, V]) All() iter.Seq2[K, V] {
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
func (l *LinkedMap[K, V]) Keys() iter.Seq[K] {
	l.init()
	return l.keys.All()
}

// Values returns an iterator over the values in insertion order.
func (l *LinkedMap[K, V]) Values() iter.Seq[V] {
	l.init()

	return func(yield func(V) bool) {
		for _, v := range l.All() {
			if !yield(v) {
				return
			}
		}
	}
}

func (l *LinkedMap[K, V]) init() {
	if l.entries == nil {
		l.entries = make(map[K]linkedMapEntry[K, V])
		l.keys = new(LinkedList[K])
	}
}

type linkedMapEntry[K comparable, V any] struct {
	value V
	node  *LinkedListNode[K]
}
