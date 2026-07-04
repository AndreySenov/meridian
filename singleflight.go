package meridian

import (
	"errors"
	"fmt"
	"sync"
)

var errGoexit = errors.New("runtime.Goexit was called in task")

// SingleFlight deduplicates concurrent work by key: while a task for a key
// is in flight, every Do call with that key joins it and receives the same
// result instead of running its own task. The zero value is ready to use.
//
// It is a typed alternative to golang.org/x/sync/singleflight: keys and
// values are generic rather than string and interface{}, so results need no
// type assertions. Each caller gets a Future that it can await under its
// own context. A panicking task is reported to every waiter as a regular error.
type SingleFlight[K comparable, V any] struct {
	mu      sync.Mutex
	flights map[K]*Promise[V]
}

// Do returns a Future for the result of the task, either starting the task
// in a new goroutine or joining the in-flight call for the key if one
// exists. Future.IsShared indicates whether the result is shared by multiple callers.
// The task runs to completion even if every caller stops waiting.
// A panic inside the task is recovered and delivered to all waiters as an
// error; a task that terminates its goroutine with runtime.Goexit also
// yields an error instead of blocking the waiters.
func (s *SingleFlight[K, V]) Do(key K, task func() (V, error)) Future[V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.flights == nil {
		s.flights = make(map[K]*Promise[V])
	}

	if flight, ok := s.flights[key]; ok {
		return flight.Future()
	}

	p := NewPromise[V]()
	s.flights[key] = p

	go func() {
		var (
			value     V
			err       error
			completed bool
		)

		defer func() {
			if !completed {
				err = errGoexit
			}

			s.mu.Lock()
			if s.flights[key] == p {
				delete(s.flights, key)
			}
			s.mu.Unlock()

			p.Complete(value, err)
		}()

		value, err = runTask(task)
		completed = true
	}()

	return p.Future()
}

func runTask[V any](task func() (V, error)) (value V, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in task: %v", r)
		}
	}()
	return task()
}

// Forget detaches the current call for the key, if any, without
// interrupting it: callers already waiting still receive its result, while
// subsequent Do calls for the key start a fresh task.
func (s *SingleFlight[K, V]) Forget(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.flights, key)
}
