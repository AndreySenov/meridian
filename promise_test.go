package meridian

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromise(t *testing.T) {
	t.Run("Resolve", func(t *testing.T) {
		p := NewPromise[string]()
		f := p.Future()

		p.Resolve("hello")

		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.Equal(t, "hello", v)
	})

	t.Run("Reject", func(t *testing.T) {
		p := NewPromise[string]()
		f := p.Future()
		wantErr := errors.New("boom")

		p.Reject(wantErr)

		v, err := f.Get(context.Background())

		require.ErrorIs(t, err, wantErr)
		require.Empty(t, v)
	})

	t.Run("Complete", func(t *testing.T) {
		p := NewPromise[int]()
		f := p.Future()

		p.Complete(7, nil)

		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.Equal(t, 7, v)
	})

	t.Run("Only first completion wins", func(t *testing.T) {
		p := NewPromise[int]()
		f := p.Future()

		p.Resolve(1)
		p.Resolve(2)
		p.Reject(errors.New("ignored"))

		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.Equal(t, 1, v)
	})

	t.Run("Multiple futures share state", func(t *testing.T) {
		p := NewPromise[int]()
		f1 := p.Future()
		f2 := p.Future()

		p.Resolve(11)

		v1, _ := f1.Get(context.Background())
		v2, _ := f2.Get(context.Background())

		require.Equal(t, 11, v1)
		require.Equal(t, 11, v2)
	})

	// Run with -race: exercises sync.Once under contention to make sure only one
	// writer ever touches promiseState.value/err.
	t.Run("Concurrent complete has no race", func(t *testing.T) {
		p := NewPromise[int]()
		f := p.Future()

		const n = 50
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.Resolve(i)
			}()
		}
		wg.Wait()

		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.GreaterOrEqual(t, v, 0)
		require.Less(t, v, n)
	})
}
