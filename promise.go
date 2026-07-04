package meridian

import (
	"sync"
	"sync/atomic"
)

// Promise is the writable side of an asynchronous result: it is
// completed at most once, and the outcome is delivered to every Future
// handle created from it. Create a Promise with NewPromise and use it by
// pointer; copying a Promise is unsafe.
type Promise[T any] struct {
	state        *promiseState[T]
	initOnce     sync.Once
	completeOnce sync.Once
}

// NewPromise returns a new, pending Promise.
func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{}
}

type promiseState[T any] struct {
	done        chan struct{}
	value       T
	err         error
	joinerCount atomic.Int64
}

// Resolve completes the Promise successfully with value.
// It is equivalent to Complete(value, nil).
func (p *Promise[T]) Resolve(value T) {
	p.Complete(value, nil)
}

// Reject completes the Promise with err.
// It is equivalent to Complete with a zero value and err.
func (p *Promise[T]) Reject(err error) {
	var zero T
	p.Complete(zero, err)
}

// Complete completes the Promise with value and err.
// Only the first completion takes effect: later calls to Complete, Resolve,
// or Reject are no-ops. Safe for concurrent use.
func (p *Promise[T]) Complete(value T, err error) {
	p.init()
	p.completeOnce.Do(func() {
		defer close(p.state.done)
		if err == nil {
			p.state.value = value
		} else {
			p.state.err = err
		}
	})
}

// Future returns a new read handle for the Promise's eventual result.
// It may be called any number of times, before or after completion;
// every handle observes the same outcome.
func (p *Promise[T]) Future() Future[T] {
	p.init()
	p.state.joinerCount.Add(1)
	return Future[T]{
		state: p.state,
	}
}

func (p *Promise[T]) init() {
	p.initOnce.Do(func() {
		p.state = &promiseState[T]{
			done: make(chan struct{}),
		}
	})
}
