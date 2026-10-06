package async

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var errGoexit = errors.New("runtime.Goexit was called in task")

// SingleFlight deduplicates concurrent work by key: while a task for a key
// is in flight, every [SingleFlight.Do] call with that key joins it and
// receives the same result instead of running its own task. The zero value
// is ready to use.
//
// It is a typed alternative to [golang.org/x/sync/singleflight]: keys and
// values are generic rather than string and interface{}, so results need no
// type assertions. Each consumer gets a [Future] that it can await under its
// own context. A panicking task is reported to every consumer as a regular error.
type SingleFlight[K comparable, V any] struct {
	mu      sync.Mutex
	flights map[K]*flight[V]
}

// SingleFlightFunc is a task called by [SingleFlight.Do]. Its context
// belongs to the call rather than to any consumer, and only
// [SingleFlight.Cancel] cancels it: the consumers then get
// [context.Canceled], and whatever the task returns is discarded.
// Implementations should honor the cancellation and return promptly.
type SingleFlightFunc[V any] func(ctx context.Context) (V, error)

// Do returns a [Future] for the result of the task, either starting the task
// in a new goroutine or joining the in-flight call for the key if one
// exists. [Future.IsShared] indicates whether the result is shared by multiple consumers.
// The task runs to completion even if every consumer stops waiting; only
// [SingleFlight.Cancel] asks it to stop.
// A panic inside the task is recovered and delivered to all consumers as an
// error; a task that terminates its goroutine with [runtime.Goexit] also
// yields an error instead of blocking the consumers.
func (s *SingleFlight[K, V]) Do(key K, task SingleFlightFunc[V]) Future[V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.flights == nil {
		s.flights = make(map[K]*flight[V])
	}

	if f, ok := s.flights[key]; ok {
		return f.promise.Future()
	}

	ctx, cancel := context.WithCancel(context.Background())
	f := &flight[V]{
		promise: NewPromise[V](),
		cancel:  cancel,
	}
	s.flights[key] = f

	go func() {
		var (
			value     V
			err       error
			completed bool
		)

		defer cancel()
		defer func() {
			if !completed {
				err = errGoexit
			}

			s.mu.Lock()
			if s.flights[key] == f {
				delete(s.flights, key)
			}
			s.mu.Unlock()

			f.promise.Complete(value, err)
		}()

		value, err = runTask(ctx, task)
		completed = true
	}()

	return f.promise.Future()
}

// Cancel cancels the context of the current call for the key and detaches
// it, so later [SingleFlight.Do] calls for the key start a fresh task. Every
// consumer receives [context.Canceled], unless the call completes first. It
// reports whether a call was in flight.
//
// The task itself is only asked to stop: one that honors its context returns
// early, while one that ignores it runs to completion with its result
// discarded.
func (s *SingleFlight[K, V]) Cancel(key K) bool {
	s.mu.Lock()
	f, ok := s.flights[key]
	if ok {
		delete(s.flights, key)
	}
	s.mu.Unlock()

	if !ok {
		return false
	}

	// Rejection must be outside the lock to keep OnComplete handlers free to call
	// back into this SingleFlight. It also precedes cancellation, so a task that
	// stops on its context cannot complete with its own error first.
	f.promise.Reject(context.Canceled)
	f.cancel()

	return true
}

// Forget detaches the current call for the key, if any, without
// interrupting it: consumers already waiting still receive its result, while
// later [SingleFlight.Do] calls for the key start a fresh task.
func (s *SingleFlight[K, V]) Forget(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.flights, key)
}

type flight[V any] struct {
	promise *Promise[V]
	cancel  context.CancelFunc
}

func runTask[V any](ctx context.Context, task SingleFlightFunc[V]) (value V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in task: %v", r)
		}
	}()
	return task(ctx)
}
