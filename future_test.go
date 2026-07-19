package meridian

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFuture(t *testing.T) {
	t.Run("Get blocks until completion", func(t *testing.T) {
		p := NewPromise[int]()
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
		p := NewPromise[int]()
		f := p.Future()

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		_, err := f.Get(ctx)

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Get prefers result over already cancelled ctx", func(t *testing.T) {
		for range 200 {
			p := NewPromise[int]()
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
		p := NewPromise[int]()
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

	t.Run("IsShared is false for a single handle", func(t *testing.T) {
		p := NewPromise[int]()
		f := p.Future()
		p.Resolve(1)
		_, _ = f.Get(context.Background())

		require.False(t, f.IsShared())
	})

	t.Run("IsShared is true once a second handle is created", func(t *testing.T) {
		p := NewPromise[int]()
		f1 := p.Future()
		f2 := p.Future()
		p.Resolve(1)
		_, _ = f1.Get(context.Background())

		require.True(t, f1.IsShared())
		require.True(t, f2.IsShared())
	})

	t.Run("IsShared reflects the number of Future handles, not calls to Get", func(t *testing.T) {
		p := NewPromise[int]()
		f := p.Future()
		p.Resolve(1)

		_, _ = f.Get(context.Background())
		_, _ = f.Get(context.Background())
		_, _ = f.Get(context.Background())

		require.False(t, f.IsShared())
	})

	t.Run("A zero-value Future panics instead of misbehaving silently", func(t *testing.T) {
		var f Future[int]

		require.PanicsWithValue(t, "Future is not initialized", func() {
			_, _ = f.Get(context.Background())
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.Done()
		})
		require.PanicsWithValue(t, "Future is not initialized", func() {
			f.IsShared()
		})
	})
}
