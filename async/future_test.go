package async_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/AndreySenov/meridian/async"
)

func TestFuture(t *testing.T) {
	t.Run("Get blocks until completion", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		go func() {
			time.Sleep(30 * time.Millisecond)
			p.Resolve(9)
		}()

		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.Equal(t, 9, v)
	})

	t.Run("Get returns ctx err on timeout", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		_, err := f.Get(ctx)

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Get prefers result over already cancelled ctx", func(t *testing.T) {
		for range 200 {
			p := async.NewPromise[int]()
			f := p.Future()
			p.Resolve(5)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			v, err := f.Get(ctx)

			require.NoError(t, err)
			require.Equal(t, 5, v)
		}
	})

	t.Run("Done blocks until completion then is closed", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		select {
		case <-f.Done():
			require.FailNow(t, "Done must not be closed before completion")
		default:
		}

		p.Resolve(9)

		select {
		case <-f.Done():
		case <-time.After(time.Second):
			require.FailNow(t, "Done must be closed after completion")
		}

		v, err := f.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, 9, v)
	})

	t.Run("IsDone reports completion without blocking", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		require.False(t, f.IsDone())

		p.Resolve(9)

		require.True(t, f.IsDone())
	})

	t.Run("IsShared is false for a single handle", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()
		p.Resolve(1)
		_, _ = f.Get(context.Background())

		require.False(t, f.IsShared())
	})

	t.Run("IsShared is true once a second handle is created", func(t *testing.T) {
		p := async.NewPromise[int]()
		f1 := p.Future()
		f2 := p.Future()
		p.Resolve(1)
		_, _ = f1.Get(context.Background())

		require.True(t, f1.IsShared())
		require.True(t, f2.IsShared())
	})

	t.Run("IsShared reflects the number of Future handles, not calls to Get", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()
		p.Resolve(1)

		_, _ = f.Get(context.Background())
		_, _ = f.Get(context.Background())
		_, _ = f.Get(context.Background())

		require.False(t, f.IsShared())
	})

	t.Run("OnComplete runs the handler on completion", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var (
			gotValue int
			gotErr   error
			calls    int
		)
		f.OnComplete(func(value int, err error) {
			gotValue, gotErr, calls = value, err, calls+1
		})

		require.Zero(t, calls)

		p.Resolve(9)

		require.Equal(t, 1, calls)
		require.Equal(t, 9, gotValue)
		require.NoError(t, gotErr)
	})

	t.Run("OnComplete passes the error to the handler", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()
		wantErr := errors.New("boom")

		var gotErr error
		f.OnComplete(func(_ int, err error) {
			gotErr = err
		})

		p.Reject(wantErr)

		require.ErrorIs(t, gotErr, wantErr)
	})

	t.Run("OnComplete runs the handler immediately if already completed", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()
		p.Resolve(9)

		var gotValue int
		f.OnComplete(func(value int, _ error) {
			gotValue = value
		})

		require.Equal(t, 9, gotValue)
	})

	t.Run("OnComplete runs every handler once, in registration order", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var order []int
		for i := range 5 {
			f.OnComplete(func(_ int, _ error) {
				order = append(order, i)
			})
		}

		p.Resolve(1)
		p.Resolve(2) // a second completion must not run the handlers again

		require.Equal(t, []int{0, 1, 2, 3, 4}, order)
	})

	t.Run("OnComplete cancel unregisters the handler and is idempotent", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var cancelled, kept int
		cancel := f.OnComplete(func(_ int, _ error) {
			cancelled++
		})
		f.OnComplete(func(_ int, _ error) {
			kept++
		})

		require.True(t, cancel())
		require.False(t, cancel())

		p.Resolve(9)

		require.Zero(t, cancelled)
		require.Equal(t, 1, kept)
	})

	t.Run("OnComplete cancel removes only its own handler, keeping the order", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var order []int
		for i := range 5 {
			cancel := f.OnComplete(func(_ int, _ error) {
				order = append(order, i)
			})
			if i == 1 || i == 3 {
				cancel()
			}
		}

		p.Resolve(9)

		require.Equal(t, []int{0, 2, 4}, order)
	})

	t.Run("OnComplete cancel is a no-op when called repeatedly or after completion", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var calls int
		cancel := f.OnComplete(func(_ int, _ error) {
			calls++
		})

		p.Resolve(9)
		require.Equal(t, 1, calls)

		require.False(t, cancel(), "the handler has already run, so nothing was stopped")
		require.False(t, cancel(), "a repeated cancel stops nothing either")

		cancelLate := f.OnComplete(func(_ int, _ error) {
			calls++
		})
		require.False(t, cancelLate())
		require.Equal(t, 2, calls)
	})

	// Run with -race
	t.Run("OnComplete cancel is safe against concurrent completion", func(t *testing.T) {
		for range 100 {
			p := async.NewPromise[int]()
			f := p.Future()

			var calls, stopped atomic.Int64
			cancels := make([]func() bool, 20)
			for i := range cancels {
				cancels[i] = f.OnComplete(func(_ int, _ error) {
					calls.Add(1)
				})
			}

			var wg sync.WaitGroup
			for _, cancel := range cancels {
				wg.Go(func() {
					if cancel() {
						stopped.Add(1)
					}
				})
			}
			wg.Go(func() {
				p.Resolve(1)
			})
			wg.Wait()

			require.Equal(t, int64(len(cancels)), calls.Load()+stopped.Load())
		}
	})

	t.Run("OnComplete handler runs after done is closed", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		var (
			doneClosed bool
			isDone     bool
			gotValue   int
			gotErr     error
		)
		f.OnComplete(func(_ int, _ error) {
			select {
			case <-f.Done():
				doneClosed = true
			default:
			}
			isDone = f.IsDone()
			gotValue, gotErr = f.Get(context.Background())
		})

		p.Resolve(9)

		require.True(t, doneClosed, "Done must already be closed inside the handler")
		require.True(t, isDone, "IsDone must report completion inside the handler")
		require.NoError(t, gotErr)
		require.Equal(t, 9, gotValue)
	})

	t.Run("OnComplete handler may complete the same Promise again", func(t *testing.T) {
		p := async.NewPromise[int]()
		f := p.Future()

		f.OnComplete(func(_ int, _ error) {
			p.Resolve(555) // must be a no-op
			f.OnComplete(func(_ int, _ error) {})
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			p.Resolve(5)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			require.FailNow(t, "re-entrant completion from a handler must not deadlock")
		}

		v, err := f.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, 5, v)
	})

	// Run with -race
	t.Run("OnComplete is safe against concurrent completion", func(t *testing.T) {
		for range 100 {
			p := async.NewPromise[int]()
			f := p.Future()

			const handlers = 20
			var calls atomic.Int64

			var wg sync.WaitGroup
			for range handlers {
				wg.Go(func() {
					f.OnComplete(func(_ int, _ error) {
						calls.Add(1)
					})
				})
			}
			wg.Go(func() {
				p.Resolve(1)
			})
			wg.Wait()

			require.Equal(t, int64(handlers), calls.Load())
		}
	})

	t.Run("A zero-value Future panics instead of misbehaving silently", func(t *testing.T) {
		var f async.Future[int]

		require.PanicsWithValue(t, "Future is not initialized", func() {
			_, _ = f.Get(context.Background())
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.Done()
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.IsDone()
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.OnComplete(func(_ int, _ error) {})
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.IsShared()
		})
	})
}
