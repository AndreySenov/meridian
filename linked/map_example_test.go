package linked_test

import (
	"fmt"
	"slices"

	"github.com/AndreySenov/meridian/linked"
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
	var m linked.Map[string, int]

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
	var m linked.Map[string, int]
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

	fmt.Println(m.Delete(1))
	fmt.Println(m.Delete(1))

	// A key stored again goes to the end of the order.
	m.Store(1, "b")

	fmt.Println(slices.Collect(m.Values()))
	// Output:
	// true
	// false
	// [a c b]
}

func ExampleMap_DeleteFunc() {
	var m linked.Map[string, int]
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
