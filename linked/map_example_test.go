package linked_test

import (
	"fmt"
	"slices"

	"github.com/AndreySenov/meridian/v2/linked"
)

func ExampleMap() {
	// The zero value is an empty map ready to use.
	var m linked.Map[string, int]

	m.Store("c", 3)
	m.Store("a", 1)
	m.Store("b", 2)

	// Iteration follows the order the keys were inserted in, not the order
	// of a Go map, which is random.
	for key, value := range m.All() {
		fmt.Println(key, value)
	}

	// Output:
	// c 3
	// a 1
	// b 2
}

func ExampleNewMapSeq() {
	// Any iter.Seq2 works, including another map's.
	m := linked.NewMapSeq(slices.All([]string{"zero", "one"}))

	fmt.Println(slices.Collect(m.Keys()))
	fmt.Println(slices.Collect(m.Values()))
	// Output:
	// [0 1]
	// [zero one]
}

func ExampleMap_Store() {
	m := linked.NewMap[string, int]()

	m.Store("a", 1)
	m.Store("b", 2)

	// Storing an existing key replaces the value and keeps the position.
	previous, replaced := m.Store("a", 10)

	fmt.Println(previous, replaced)
	fmt.Println(slices.Collect(m.Keys()))
	// Output:
	// 1 true
	// [a b]
}

func ExampleMap_Load() {
	m := linked.NewMap[string, int]()
	m.Store("a", 1)

	value, loaded := m.Load("a")
	fmt.Println(value, loaded)

	value, loaded = m.Load("missing")
	fmt.Println(value, loaded)

	// Output:
	// 1 true
	// 0 false
}

func ExampleMap_Delete() {
	m := linked.NewMapSeq(slices.All([]string{"a", "b", "c"}))

	value, deleted := m.Delete(1)
	fmt.Println(value, deleted)

	// Deleting a missing key yields the zero value.
	value, deleted = m.Delete(1)
	fmt.Printf("%q %v\n", value, deleted)

	// A key stored again goes to the end of the order.
	m.Store(1, "b")

	fmt.Println(slices.Collect(m.Values()))
	// Output:
	// b true
	// "" false
	// [a c b]
}

func ExampleMap_DeleteFunc() {
	m := linked.NewMap[string, int]()
	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	deleted := m.DeleteFunc(func(_ string, value int) bool {
		return value%2 == 0
	})

	fmt.Println(deleted)
	fmt.Println(slices.Collect(m.Keys()))
	// Output:
	// 1
	// [a c]
}

func ExampleMap_Keys() {
	m := linked.NewMapSeq(slices.All([]string{"a", "b"}))

	for key := range m.Keys() {
		fmt.Println("key:", key)
	}

	for value := range m.Values() {
		fmt.Println("value:", value)
	}

	// Output:
	// key: 0
	// key: 1
	// value: a
	// value: b
}

func ExampleMap_Last() {
	cache := linked.NewAccessOrderedMap[string, int]()
	cache.Store("a", 1)
	cache.Store("b", 2)

	// The eviction candidate can be inspected before deciding to drop it,
	// because neither First nor Last counts as a use.
	key, value, _ := cache.Last()
	fmt.Println("next to evict:", key, value)

	firstKey, _, _ := cache.First()
	fmt.Println("most recently used:", firstKey)

	fmt.Println(slices.Collect(cache.Keys()))
	// Output:
	// next to evict: a 1
	// most recently used: b
	// [b a]
}

func ExampleMap_DeleteFirst() {
	m := linked.NewMapSeq(slices.All([]string{"a", "b", "c"}))

	firstKey, firstValue, _ := m.DeleteFirst()
	lastKey, lastValue, _ := m.DeleteLast()

	fmt.Println(firstKey, firstValue)
	fmt.Println(lastKey, lastValue)
	fmt.Println(slices.Collect(m.Values()))
	// Output:
	// 0 a
	// 2 c
	// [b]
}

func ExampleNewAccessOrderedMap() {
	m := linked.NewAccessOrderedMap[string, int]()

	m.Store("a", 1)
	m.Store("b", 2)
	m.Store("c", 3)

	// Both Load and Store count as a use, so the entry just touched moves
	// to the front and iteration runs from the most recently used one.
	m.Load("a")

	fmt.Println(slices.Collect(m.Keys()))
	// Output: [a c b]
}

func ExampleNewAccessOrderedMapSeq() {
	m := linked.NewAccessOrderedMapSeq(slices.All([]string{"a", "b", "c"}))

	// The entry stored last is the most recently used one.
	fmt.Println(slices.Collect(m.Values()))
	// Output: [c b a]
}

func ExampleMap_DeleteLast() {
	// A least-recently-used cache of two entries.
	cache := linked.NewAccessOrderedMap[string, int]()

	put := func(key string, value int) {
		cache.Store(key, value)
		if cache.Len() > 2 {
			deletedKey, deletedValue, _ := cache.DeleteLast()
			fmt.Println("evicted:", deletedKey, deletedValue)
		}
	}

	put("a", 1)
	put("b", 2)
	cache.Load("a") // a is used again, so b becomes the eviction target
	put("c", 3)     // over capacity: b goes

	fmt.Println(slices.Collect(cache.Keys()))
	// Output:
	// evicted: b 2
	// [c a]
}
