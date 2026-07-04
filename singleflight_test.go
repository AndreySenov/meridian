package meridian

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSingleFlight(t *testing.T) {
	t.Run("Deduplicates concurrent calls", func(t *testing.T) {
		var sf SingleFlight[string, int]
		var calls int32

		const n = 50
		start := make(chan struct{})
		results := make([]int, n)
		errs := make([]error, n)

		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				f := sf.Do("key", func() (int, error) {
					atomic.AddInt32(&calls, 1)
					time.Sleep(30 * time.Millisecond)
					return 42, nil
				})
				results[i], errs[i] = f.Get(context.Background())
			}()
		}
		close(start)
		wg.Wait()

		require.Equal(t, int32(1), atomic.LoadInt32(&calls))

		for i := range results {
			require.NoError(t, errs[i])
			require.Equal(t, 42, results[i])
		}
	})

	t.Run("Runs again after completion", func(t *testing.T) {
		var sf SingleFlight[string, int]
		var calls int32

		for i := 0; i < 2000; i++ {
			f1 := sf.Do("key", func() (int, error) {
				return int(atomic.AddInt32(&calls, 1)), nil
			})
			v1, err1 := f1.Get(context.Background())
			require.NoError(t, err1)
			require.Equal(t, int(2*i+1), v1)

			f2 := sf.Do("key", func() (int, error) {
				return int(atomic.AddInt32(&calls, 1)), nil
			})
			v2, err2 := f2.Get(context.Background())
			require.NoError(t, err2)
			require.Equal(t, int(2*i+2), v2)
		}
	})

	t.Run("Different keys are independent", func(t *testing.T) {
		var sf SingleFlight[string, int]
		var calls int32

		f1 := sf.Do("a", func() (int, error) {
			atomic.AddInt32(&calls, 1)
			return 1, nil
		})
		f2 := sf.Do("b", func() (int, error) {
			atomic.AddInt32(&calls, 1)
			return 2, nil
		})

		v1, err1 := f1.Get(context.Background())
		v2, err2 := f2.Get(context.Background())

		require.NoError(t, err1)
		require.Equal(t, 1, v1)

		require.Equal(t, 2, v2)
		require.NoError(t, err2)

		require.Equal(t, int32(2), atomic.LoadInt32(&calls))
	})

	t.Run("Error is shared by all waiters", func(t *testing.T) {
		var sf SingleFlight[string, int]
		wantErr := errors.New("boom")

		const n = 10
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				f := sf.Do("key", func() (int, error) {
					return 0, wantErr
				})
				_, errs[i] = f.Get(context.Background())
			}()
		}
		wg.Wait()

		for _, err := range errs {
			require.ErrorIs(t, err, wantErr)
		}
	})

	t.Run("Panic is recovered and returned as error", func(t *testing.T) {
		var sf SingleFlight[string, int]

		f := sf.Do("key", func() (int, error) {
			panic("boom")
		})

		_, err := f.Get(context.Background())

		require.NotNil(t, err)
		require.Contains(t, err.Error(), "boom")
	})

	t.Run("Panic does not crash other waiters", func(t *testing.T) {
		var sf SingleFlight[string, int]

		const n = 10
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				f := sf.Do("key", func() (int, error) {
					panic("boom")
				})
				_, errs[i] = f.Get(context.Background())
			}()
		}
		wg.Wait()

		for _, err := range errs {
			require.NotNil(t, err)
		}
	})

	t.Run("Key is usable after panic", func(t *testing.T) {
		var sf SingleFlight[string, int]

		f1 := sf.Do("key", func() (int, error) {
			panic("boom")
		})

		_, err := f1.Get(context.Background())
		require.NotNil(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		f2 := sf.Do("key", func() (int, error) {
			return 7, nil
		})

		v2, err2 := f2.Get(ctx)

		require.NoError(t, err2)
		require.Equal(t, 7, v2)
	})

	t.Run("Goexit in task unblocks waiters and frees the key", func(t *testing.T) {
		var sf SingleFlight[string, int]

		f1 := sf.Do("key", func() (int, error) {
			runtime.Goexit()
			return 1, nil // unreachable
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := f1.Get(ctx)
		require.Error(t, err)
		require.NotErrorIs(t, err, context.DeadlineExceeded, "waiter must get the task's error, not hang until its own ctx")

		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
		defer cancel2()

		f2 := sf.Do("key", func() (int, error) {
			return 7, nil
		})

		v2, err2 := f2.Get(ctx2)
		require.NoError(t, err2)
		require.Equal(t, 7, v2)
	})

	// Run with -race
	t.Run("Concurrent mix of panics and successes", func(t *testing.T) {
		var sf SingleFlight[string, int]
		keys := []string{"a", "b", "c"}

		var wg sync.WaitGroup
		for round := 0; round < 20; round++ {
			for _, k := range keys {
				k := k
				round := round
				wg.Add(1)
				go func() {
					defer wg.Done()
					f := sf.Do(k, func() (int, error) {
						if round%2 == 0 {
							panic("fail")
						}
						return round, nil
					})
					_, _ = f.Get(context.Background())
				}()
			}
		}

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.Fail(t, "timed out waiting for calls to finish - possible deadlock or leaked flight")
		}
	})

	t.Run("Forget lets a new caller bypass a still in-flight call", func(t *testing.T) {
		var sf SingleFlight[string, int]

		aStarted := make(chan struct{})
		aRelease := make(chan struct{})
		fA := sf.Do("key", func() (int, error) {
			close(aStarted)
			<-aRelease
			return 1, nil
		})
		<-aStarted

		sf.Forget("key")

		fB := sf.Do("key", func() (int, error) {
			return 2, nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		vB, errB := fB.Get(ctx)
		require.NoError(t, errB)
		require.Equal(t, 2, vB)

		close(aRelease)
		vA, errA := fA.Get(context.Background())
		require.NoError(t, errA)
		require.Equal(t, 1, vA)
	})

	t.Run("Forget on unknown key is a no-op", func(t *testing.T) {
		var sf SingleFlight[string, int]

		require.NotPanics(t, func() {
			sf.Forget("never-seen")
		})

		f := sf.Do("key", func() (int, error) {
			return 1, nil
		})
		v, err := f.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, v)
	})

	t.Run("Forget during an in-flight call does not corrupt a newer flight", func(t *testing.T) {
		var sf SingleFlight[string, int]

		aStarted := make(chan struct{})
		aRelease := make(chan struct{})
		fA := sf.Do("key", func() (int, error) {
			close(aStarted)
			<-aRelease
			return 1, nil
		})
		<-aStarted

		sf.Forget("key")

		bStarted := make(chan struct{})
		bRelease := make(chan struct{})
		fB := sf.Do("key", func() (int, error) {
			close(bStarted)
			<-bRelease
			return 2, nil
		})
		<-bStarted

		close(aRelease)
		_, errA := fA.Get(context.Background())
		require.NoError(t, errA)

		time.Sleep(30 * time.Millisecond)

		var freshTaskRan int32
		fC := sf.Do("key", func() (int, error) {
			atomic.AddInt32(&freshTaskRan, 1)
			return 3, nil
		})

		close(bRelease)
		vB, errB := fB.Get(context.Background())
		vC, errC := fC.Get(context.Background())

		require.NoError(t, errB)
		require.NoError(t, errC)
		require.Equal(t, 2, vB)
		require.Equal(t, vB, vC, "fC should have joined the still in-flight B, not started a fresh task")
		require.Zero(t, atomic.LoadInt32(&freshTaskRan), "a fresh task should not run for a key with an in-flight flight")
	})

	t.Run("IsShared is false when nobody joins the call", func(t *testing.T) {
		var sf SingleFlight[string, int]

		f := sf.Do("key", func() (int, error) {
			return 1, nil
		})
		v, err := f.Get(context.Background())

		require.NoError(t, err)
		require.Equal(t, 1, v)
		require.False(t, f.IsShared())
	})

	t.Run("IsShared is true once a caller joins an in-flight call", func(t *testing.T) {
		var sf SingleFlight[string, int]

		started := make(chan struct{})
		release := make(chan struct{})
		f1 := sf.Do("key", func() (int, error) {
			close(started)
			<-release
			return 1, nil
		})
		<-started

		f2 := sf.Do("key", func() (int, error) {
			return 2, nil
		})

		close(release)
		v1, err1 := f1.Get(context.Background())
		v2, err2 := f2.Get(context.Background())

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, 1, v1)
		require.Equal(t, 1, v2)

		require.True(t, f1.IsShared())
		require.True(t, f2.IsShared())
	})

	t.Run("IsShared reflects concurrent joiners under -race", func(t *testing.T) {
		var sf SingleFlight[string, int]

		started := make(chan struct{})
		release := make(chan struct{})
		f0 := sf.Do("key", func() (int, error) {
			close(started)
			<-release
			return 1, nil
		})
		<-started

		const joiners = 20
		futures := make([]Future[int], joiners)
		var wg sync.WaitGroup
		for i := 0; i < joiners; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				futures[i] = sf.Do("key", func() (int, error) {
					return 0, nil
				})
			}()
		}
		wg.Wait()
		close(release)

		_, err := f0.Get(context.Background())
		require.NoError(t, err)
		require.True(t, f0.IsShared())

		for _, f := range futures {
			_, err := f.Get(context.Background())

			require.NoError(t, err)
			require.True(t, f.IsShared())
		}
	})

	t.Run("A fresh call after the previous one completed is not shared", func(t *testing.T) {
		var sf SingleFlight[string, int]

		f1 := sf.Do("key", func() (int, error) { return 1, nil })
		_, _ = f1.Get(context.Background())

		require.False(t, f1.IsShared())

		f2 := sf.Do("key", func() (int, error) { return 2, nil })
		v2, err2 := f2.Get(context.Background())

		require.NoError(t, err2)
		require.Equal(t, 2, v2)
		require.False(t, f2.IsShared())
	})
}
