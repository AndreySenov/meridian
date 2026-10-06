# Meridian

[![Build Status](https://github.com/AndreySenov/meridian/actions/workflows/default.yml/badge.svg)](https://github.com/AndreySenov/meridian/actions)
[![Latest Release](https://img.shields.io/github/v/release/AndreySenov/meridian?color=00ADD8)](https://github.com/AndreySenov/meridian/releases)
[![License](https://img.shields.io/github/license/AndreySenov/meridian?color=00ADD8)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/AndreySenov/meridian/v4.svg)](https://pkg.go.dev/github.com/AndreySenov/meridian/v4)

A Go library of concurrency and collection utilities.

## Features

- Package `async`: **Promise** and **Future**, **SingleFlight**
- Package `linked`: **List**, **Map**

## Installation

Run the `go get` command to install Meridian:

```sh
go get github.com/AndreySenov/meridian/v4
```

Use the `-u` flag to update Meridian to the latest version:

```sh
go get -u github.com/AndreySenov/meridian/v4
```

Then import the package you need:

```go
import (
	"github.com/AndreySenov/meridian/v4/async"
	"github.com/AndreySenov/meridian/v4/linked"
)
```

## Promise and Future

`async.Promise` and `async.Future` are companion constructs used to handle results of asynchronous tasks.
The `async.Promise` produces the result at most once. The `async.Future` provides a read-only interface to consume the result.
Any number of `async.Future` handles can observe the outcome of the same `async.Promise`.

Usage example:
```go
func GetProfile(ctx context.Context, id string) (*Profile, error) {
	p := async.NewPromise[*Profile]()

	go func() {
		r, err := fetchProfileFromDB(id)
		p.Complete(r, err) // or p.Resolve(r) / p.Reject(err)
	}()

	f := p.Future()
	return f.Get(ctx) // blocks until completed or ctx is done
}
```

An alternative way to consume the result is to register an `OnComplete` handler.
The handler runs on the goroutine that completes the `async.Promise`, or immediately
on the calling goroutine if the `async.Promise` is already completed.
Multiple `OnComplete` handlers can be registered, including on different
`async.Future` handles of the same `async.Promise`;
the order of execution matches the order of registration.

Usage example:
```go
f.OnComplete(func(profile *Profile, err error) {
	if err != nil {
		log.Printf("profile %s failed: %v", id, err)
		return
	}
	cache.Put(id, profile)
})
```

## SingleFlight

`async.SingleFlight` is an alternative to
[golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
with generic keys and values.

While a task for a key is in flight, every `Do` call with that key joins it and
receives the same result instead of running its own task:

```go
var flights async.SingleFlight[string, *Profile]

func LoadProfile(ctx context.Context, id string) (*Profile, error) {
	future := flights.Do(id, func(taskCtx context.Context) (*Profile, error) {
		return fetchProfileFromDB(taskCtx, id) // runs once per key, no matter how many callers
	})
	return future.Get(ctx) // each consumer waits with its own context
}
```

## List and Map

A `linked.List` is a doubly linked list, a typed alternative to `container/list`.
A `linked.Map` is a map that iterates over its entries in a predictable order.
Both are ready to use as zero values.

Usage example:
```go
l := linked.NewList("a", "b", "c")
l.MoveToFront(l.Back())
fmt.Println(l.ToSlice()) // [c a b]

m := linked.NewMap[string, int]()
m.Store("first", 1)
m.Store("second", 2)
for name, n := range m.All() {
	fmt.Println(name, n) // first 1, then second 2
}
```

A `linked.Map` from `linked.NewAccessOrderedMap` iterates from the most recently
used entry to the least recently used one, where both `Load` and `Store` count
as a use. Together with `DeleteLast`, which removes and returns the entry at the
end of the order, that makes an LRU cache:

```go
cache := linked.NewAccessOrderedMap[string, int]()

put := func(key string, value int) {
	cache.Store(key, value)
	if cache.Len() > capacity {
		// The least recently used entry, key included.
		evictedKey, evictedValue, _ := cache.DeleteLast()
		log.Println("evicted", evictedKey, evictedValue)
	}
}
```

## Documentation

See the [package documentation](https://pkg.go.dev/github.com/AndreySenov/meridian/v4) for the full API reference.

## License

Meridian is licensed under the Apache License, Version 2.0. See [NOTICE](NOTICE) and [LICENSE](LICENSE) for details.
