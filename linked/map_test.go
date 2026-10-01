package linked_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AndreySenov/meridian/v2/linked"
)

type pair struct {
	key   string
	value int
}

func pairs(m *linked.Map[string, int]) []pair {
	var out []pair
	for k, v := range m.All() {
		out = append(out, pair{k, v})
	}
	return out
}

func newMap(values ...pair) *linked.Map[string, int] {
	m := linked.NewMap[string, int]()
	for _, p := range values {
		m.Store(p.key, p.value)
	}
	return m
}

func pairSeq(values ...pair) iter.Seq2[string, int] {
	return func(yield func(string, int) bool) {
		for _, p := range values {
			if !yield(p.key, p.value) {
				return
			}
		}
	}
}

func TestMap(t *testing.T) {
	t.Run("Zero value is an empty map", func(t *testing.T) {
		var m linked.Map[string, int]

		v, loaded := m.Load("a")
		require.False(t, loaded)
		require.Zero(t, v)

		require.False(t, m.Delete("a"))
		require.Zero(t, m.DeleteFunc(func(string, int) bool { return true }))
		require.Empty(t, pairs(&m))
		require.Empty(t, slices.Collect(m.Keys()))
		require.Empty(t, slices.Collect(m.Values()))
	})

	t.Run("NewMapSeq collects an iterator in order", func(t *testing.T) {
		m := linked.NewMapSeq(pairSeq(pair{"c", 3}, pair{"a", 1}, pair{"b", 2}))

		require.Equal(t, []pair{{"c", 3}, {"a", 1}, {"b", 2}}, pairs(m))
		require.Equal(t, 3, m.Len())

		indexed := linked.NewMapSeq(slices.All([]string{"x", "y"}))
		require.Equal(t, []int{0, 1}, slices.Collect(indexed.Keys()))
		require.Equal(t, []string{"x", "y"}, slices.Collect(indexed.Values()))

		require.True(t, linked.NewMapSeq(pairSeq()).IsEmpty())
	})

	t.Run("NewMapSeq keeps the first position and the last value of a repeated key", func(t *testing.T) {
		m := linked.NewMapSeq(pairSeq(pair{"a", 1}, pair{"b", 2}, pair{"a", 3}))

		require.Equal(t, []pair{{"a", 3}, {"b", 2}}, pairs(m))
		require.Equal(t, 2, m.Len())
	})

	t.Run("NewMapSeq over All copies a map", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2})

		c := linked.NewMapSeq(m.All())
		c.Store("c", 3)
		m.Delete("a")

		require.Equal(t, []pair{{"a", 1}, {"b", 2}, {"c", 3}}, pairs(c))
		require.Equal(t, []pair{{"b", 2}}, pairs(m))
	})

	t.Run("NewMap returns an empty insertion-ordered map", func(t *testing.T) {
		m := linked.NewMap[string, int]()

		require.True(t, m.IsEmpty())

		m.Store("b", 2)
		m.Store("a", 1)
		require.Equal(t, []pair{{"b", 2}, {"a", 1}}, pairs(m))
	})

	t.Run("Store and Load", func(t *testing.T) {
		var m linked.Map[string, int]

		previous, replaced := m.Store("a", 1)
		require.False(t, replaced)
		require.Zero(t, previous)

		v, loaded := m.Load("a")
		require.True(t, loaded)
		require.Equal(t, 1, v)

		_, loaded = m.Load("missing")
		require.False(t, loaded)
	})

	t.Run("Store replaces the value of an existing key in place", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		previous, replaced := m.Store("b", 20)
		require.True(t, replaced)
		require.Equal(t, 2, previous)

		v, loaded := m.Load("b")
		require.True(t, loaded)
		require.Equal(t, 20, v)

		require.Equal(t, []string{"a", "b", "c"}, slices.Collect(m.Keys()))
		require.Equal(t, []int{1, 20, 3}, slices.Collect(m.Values()))
	})

	t.Run("Len and IsEmpty track the entries", func(t *testing.T) {
		var m linked.Map[string, int]

		require.Zero(t, m.Len())
		require.True(t, m.IsEmpty())

		m.Store("a", 1)
		m.Store("b", 2)
		m.Store("c", 3)
		require.Equal(t, 3, m.Len())
		require.False(t, m.IsEmpty())

		m.Store("b", 20)
		require.Equal(t, 3, m.Len(), "replacing a value adds no entry")

		m.Delete("a")
		require.Equal(t, 2, m.Len())

		m.DeleteFunc(func(string, int) bool { return true })
		require.Zero(t, m.Len())
		require.True(t, m.IsEmpty())
	})

	t.Run("Iteration follows insertion order", func(t *testing.T) {
		m := newMap(pair{"c", 3}, pair{"a", 1}, pair{"b", 2})

		want := []pair{{"c", 3}, {"a", 1}, {"b", 2}}
		for range 20 {
			require.Equal(t, want, pairs(m))
		}
		require.Equal(t, []string{"c", "a", "b"}, slices.Collect(m.Keys()))
		require.Equal(t, []int{3, 1, 2}, slices.Collect(m.Values()))
	})

	t.Run("Delete removes the key and keeps the order of the rest", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		require.True(t, m.Delete("b"))

		_, loaded := m.Load("b")
		require.False(t, loaded)
		require.Equal(t, []pair{{"a", 1}, {"c", 3}}, pairs(m))

		require.False(t, m.Delete("b"), "deleting twice reports nothing to delete")
		require.False(t, m.Delete("missing"))
		require.Equal(t, []pair{{"a", 1}, {"c", 3}}, pairs(m))
	})

	t.Run("A key stored again after deletion goes to the end", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		m.Delete("a")
		_, replaced := m.Store("a", 10)

		require.False(t, replaced)
		require.Equal(t, []pair{{"b", 2}, {"c", 3}, {"a", 10}}, pairs(m))
	})

	t.Run("Clear removes every entry", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		m.Clear()

		require.True(t, m.IsEmpty())
		require.Zero(t, m.Len())
		require.Empty(t, pairs(m))
		require.Empty(t, slices.Collect(m.Keys()))

		_, loaded := m.Load("a")
		require.False(t, loaded)
		require.False(t, m.Delete("a"))

		m.Store("c", 30)
		_, replaced := m.Store("a", 10)
		require.False(t, replaced)
		require.Equal(t, []pair{{"c", 30}, {"a", 10}}, pairs(m))
	})

	t.Run("DeleteFirst and DeleteLast take entries from the ends of the order", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		value, deleted := m.DeleteFirst()
		require.True(t, deleted)
		require.Equal(t, 1, value)

		value, deleted = m.DeleteLast()
		require.True(t, deleted)
		require.Equal(t, 3, value)

		require.Equal(t, []pair{{"b", 2}}, pairs(m))

		_, loaded := m.Load("a")
		require.False(t, loaded)
	})

	t.Run("DeleteFirst and DeleteLast report an empty map", func(t *testing.T) {
		var m linked.Map[string, int]

		value, deleted := m.DeleteFirst()
		require.False(t, deleted)
		require.Zero(t, value)

		value, deleted = m.DeleteLast()
		require.False(t, deleted)
		require.Zero(t, value)
	})

	t.Run("Clear on an empty map is a no-op", func(t *testing.T) {
		var m linked.Map[string, int]

		require.NotPanics(t, m.Clear)
		require.True(t, m.IsEmpty())
	})

	t.Run("A copy of a used map shares its entries", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2})
		c := *m

		c.Store("c", 3)
		m.Delete("a")
		c.Store("b", 20)

		require.Equal(t, []pair{{"b", 20}, {"c", 3}}, pairs(m))
		require.Equal(t, []pair{{"b", 20}, {"c", 3}}, pairs(&c))
		require.Equal(t, 2, m.Len())
		require.Equal(t, 2, c.Len())
	})

	t.Run("A copy of a zero-value map is independent, like a nil map", func(t *testing.T) {
		var m linked.Map[string, int]
		c := m

		m.Store("a", 1)
		c.Store("b", 2)

		require.Equal(t, []pair{{"a", 1}}, pairs(&m))
		require.Equal(t, []pair{{"b", 2}}, pairs(&c))
	})

	t.Run("DeleteFunc removes the matching entries in order", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3}, pair{"d", 4})

		var seen []string
		deleted := m.DeleteFunc(func(k string, v int) bool {
			seen = append(seen, k)
			return v%2 == 0
		})

		require.Equal(t, 2, deleted)
		require.Equal(t, []string{"a", "b", "c", "d"}, seen)
		require.Equal(t, []pair{{"a", 1}, {"c", 3}}, pairs(m))

		_, loaded := m.Load("b")
		require.False(t, loaded)
	})

	t.Run("DeleteFunc reports zero when nothing matches", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2})

		require.Zero(t, m.DeleteFunc(func(string, int) bool { return false }))
		require.Equal(t, []pair{{"a", 1}, {"b", 2}}, pairs(m))
	})

	t.Run("Iteration stops early on break", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		var keys []string
		for k := range m.All() {
			keys = append(keys, k)
			if k == "b" {
				break
			}
		}
		require.Equal(t, []string{"a", "b"}, keys)

		var values []int
		for v := range m.Values() {
			values = append(values, v)
			break
		}
		require.Equal(t, []int{1}, values)
	})

	t.Run("Entries may be deleted and stored during iteration", func(t *testing.T) {
		m := newMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		var seen []string
		for k := range m.All() {
			seen = append(seen, k)
			switch k {
			case "a":
				m.Delete("b") // the entry ahead is skipped
			case "c":
				m.Store("d", 4) // appended, so it is visited
			case "d":
				m.Delete("d") // the current entry
			}
		}

		require.Equal(t, []string{"a", "c", "d"}, seen)
		require.Equal(t, []pair{{"a", 1}, {"c", 3}}, pairs(m))
	})
}

func newAccessOrderedMap(values ...pair) *linked.Map[string, int] {
	m := linked.NewAccessOrderedMap[string, int]()
	for _, p := range values {
		m.Store(p.key, p.value)
	}
	return m
}

func TestAccessOrderedMap(t *testing.T) {
	t.Run("Iteration starts from the most recently used entry", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		require.Equal(t, []pair{{"c", 3}, {"b", 2}, {"a", 1}}, pairs(m))
		require.Equal(t, []string{"c", "b", "a"}, slices.Collect(m.Keys()))
		require.Equal(t, []int{3, 2, 1}, slices.Collect(m.Values()))
		require.Equal(t, 3, m.Len())
	})

	t.Run("Load counts as a use and reorders the entries", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		value, loaded := m.Load("a")

		require.True(t, loaded)
		require.Equal(t, 1, value)
		require.Equal(t, []string{"a", "c", "b"}, slices.Collect(m.Keys()))
	})

	t.Run("Loading a missing key leaves the order alone", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2})

		_, loaded := m.Load("missing")

		require.False(t, loaded)
		require.Equal(t, []string{"b", "a"}, slices.Collect(m.Keys()))
	})

	t.Run("Store counts as a use for an existing key", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		previous, replaced := m.Store("a", 10)

		require.True(t, replaced)
		require.Equal(t, 1, previous)
		require.Equal(t, []pair{{"a", 10}, {"c", 3}, {"b", 2}}, pairs(m))
		require.Equal(t, 3, m.Len())
	})

	t.Run("NewAccessOrderedMapSeq matches storing the entries one by one", func(t *testing.T) {
		values := []pair{{"a", 1}, {"b", 2}, {"c", 3}}

		fromSeq := linked.NewAccessOrderedMapSeq(pairSeq(values...))

		require.Equal(t, pairs(newAccessOrderedMap(values...)), pairs(fromSeq))
		require.True(t, linked.NewAccessOrderedMapSeq(pairSeq()).IsEmpty())
	})

	t.Run("Delete keeps the order of the rest", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		require.True(t, m.Delete("b"))

		require.Equal(t, []pair{{"c", 3}, {"a", 1}}, pairs(m))
		require.False(t, m.Delete("b"))
	})

	t.Run("DeleteFunc sees the entries most recently used first", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})

		var seen []string
		deleted := m.DeleteFunc(func(key string, _ int) bool {
			seen = append(seen, key)
			return key == "c"
		})

		require.Equal(t, 1, deleted)
		require.Equal(t, []string{"c", "b", "a"}, seen)
		require.Equal(t, []pair{{"b", 2}, {"a", 1}}, pairs(m))
	})

	t.Run("DeleteLast evicts the least recently used entry", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2}, pair{"c", 3})
		m.Load("a") // a is now the most recently used, b the least

		value, deleted := m.DeleteLast()

		require.True(t, deleted)
		require.Equal(t, 2, value)
		require.Equal(t, []pair{{"a", 1}, {"c", 3}}, pairs(m))

		_, loaded := m.Load("b")
		require.False(t, loaded)
	})

	t.Run("DeleteFirst removes the most recently used entry", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2})

		value, deleted := m.DeleteFirst()

		require.True(t, deleted)
		require.Equal(t, 2, value)
		require.Equal(t, []pair{{"a", 1}}, pairs(m))
	})

	t.Run("Clear empties the map and keeps its order", func(t *testing.T) {
		m := newAccessOrderedMap(pair{"a", 1}, pair{"b", 2})

		m.Clear()
		require.True(t, m.IsEmpty())

		m.Store("c", 3)
		m.Store("d", 4)
		require.Equal(t, []pair{{"d", 4}, {"c", 3}}, pairs(m), "the map stays access-ordered")
	})
}
