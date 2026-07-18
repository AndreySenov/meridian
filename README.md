# Meridian

[![Build Status](https://github.com/AndreySenov/meridian/actions/workflows/default.yml/badge.svg)](https://github.com/AndreySenov/meridian/actions)
[![Latest Release](https://img.shields.io/github/v/release/AndreySenov/meridian?color=00ADD8)](https://github.com/AndreySenov/meridian/releases)
[![License](https://img.shields.io/github/license/AndreySenov/meridian?color=00ADD8)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/AndreySenov/meridian.svg)](https://pkg.go.dev/github.com/AndreySenov/meridian)

A Go library of concurrency utilities.

## Features

- **Promise** and **Future**
- **SingleFlight**

## Installation

Run the `go get` command to install Meridian:

```sh
go get github.com/AndreySenov/meridian
```

Use the `-u` flag to update Meridian to the latest version:

```sh
go get -u github.com/AndreySenov/meridian
```

## Promise and Future

A `Promise` and a `Future` are companion constructs used to handle results of asynchronous tasks.
The `Promise` produces the result at most once. The `Future` provides a read-only interface to consume the result.
Any number of `Future` handles can observe the outcome of the same `Promise`.

Usage example:
```go
func GetProfile(ctx context.Context, id string) (*Profile, error) {
	p := meridian.NewPromise[*Profile]()

	go func() {
		r, err := fetchProfileFromDB(id)
		p.Complete(r, err) // or p.Resolve(r) / p.Reject(err)
	}()

	f := p.Future()
	return f.Get(ctx) // blocks until completed or ctx is done
}
```

## SingleFlight

SingleFlight is an alternative to
[golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
with generic keys and values.

While a task for a key is in flight, every `Do` call with that key joins it and
receives the same result instead of running its own task:

```go
var flights meridian.SingleFlight[string, *Profile]

func LoadProfile(ctx context.Context, id string) (*Profile, error) {
	future := flights.Do(id, func() (*Profile, error) {
		return fetchProfileFromDB(id) // runs once per key, no matter how many callers
	})
	return future.Get(ctx) // each caller waits with its own context
}
```

## Documentation

See the [package documentation](https://pkg.go.dev/github.com/AndreySenov/meridian) for the full API reference.

## License

Meridian is licensed under the Apache License, Version 2.0. See [NOTICE](NOTICE) and [LICENSE](LICENSE) for details.
