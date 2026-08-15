package meridian

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
func (f Future[T]) OnComplete(handler func(value T, err error)) {
	f.check()

	f.state.completeMu.Lock()
	if f.state.completed {
		f.state.completeMu.Unlock()
		handler(f.state.value, f.state.err)
		return
	}
	f.state.onCompleteHandlers = append(f.state.onCompleteHandlers, handler)
	f.state.completeMu.Unlock()
}

// IsShared reports whether other Future handles exist for the same Promise.
// When true, the result value is shared: if T is a pointer, slice, or map,
// mutating the value affects the other holders, so treat it as read-only.
func (f Future[T]) IsShared() bool {
	f.check()
	return f.state.joinerCount.Load() > 1
}

func (f Future[T]) check() {
	if f.state == nil {
		panic("Future is not initialized")
	}
}
