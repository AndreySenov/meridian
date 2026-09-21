package async

import (
	"context"
)

// Future is a read handle for a Promise's eventual result, created by
// Promise.Future. Futures are small values, safe to copy and to share
// between goroutines. The zero value has no Promise behind it, and its
// methods panic.
type Future[T any] struct {
	state *promiseState[T]
}

// Get returns the result, blocking until the Promise is completed or ctx
// is done. If ctx ends first, Get returns the zero value and ctx.Err();
// if both are ready, the result wins. Get may be called repeatedly.
func (f Future[T]) Get(ctx context.Context) (T, error) {
	f.check()

	select {
	case <-f.Done():
		return f.state.value, f.state.err
	case <-ctx.Done():
		select {
		case <-f.Done():
			return f.state.value, f.state.err
		default:
			var zero T
			return zero, ctx.Err()
		}
	}
}

// Done returns a channel that is closed when the Promise is completed.
func (f Future[T]) Done() <-chan struct{} {
	f.check()
	return f.state.done
}

// OnComplete registers a handler for the result. If the Promise is already
// completed, the handler runs immediately on the calling goroutine;
// otherwise OnComplete returns at once and the handler runs on the
// goroutine that completes the Promise, after Done is closed. Handlers run
// one after another in registration order, so a handler should be quick and
// must not panic: it holds up the handlers behind it, and its panic
// surfaces on whichever goroutine runs it. Start a goroutine inside the
// handler for slow work.
//
// Calling the returned cancel function unregisters the handler and releases
// it. It reports whether it stopped the handler from being run: false means
// the handler has already been picked up for execution, or was cancelled
// before. Cancelling is safe at any time and any number of times, but it
// does not wait for a handler that is already running.
func (f Future[T]) OnComplete(handler func(value T, err error)) (cancel func() bool) {
	f.check()

	f.state.completeMu.Lock()
	if f.state.completed {
		f.state.completeMu.Unlock()
		handler(f.state.value, f.state.err)
		return func() bool { return false }
	}

	id := f.state.nextHandlerID
	f.state.nextHandlerID++
	f.state.onCompleteHandlers.Store(id, handler)
	f.state.completeMu.Unlock()

	return func() bool {
		f.state.completeMu.Lock()
		defer f.state.completeMu.Unlock()

		return f.state.onCompleteHandlers.Delete(id)
	}
}

// IsShared reports whether other Future handles exist for the same Promise.
// When true, the result value is shared: if T is a pointer, slice, or map,
// mutating the value affects the other holders, so treat it as read-only.
func (f Future[T]) IsShared() bool {
	f.check()
	return f.state.joinerCount.Load() > 1
}

// IsDone reports whether the Promise has been completed.
func (f Future[T]) IsDone() bool {
	f.check()
	select {
	case <-f.Done():
		return true
	default:
		return false
	}
}

func (f Future[T]) check() {
	if f.state == nil {
		panic("Future is not initialized")
	}
}
