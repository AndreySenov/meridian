package async_test

import (
	"context"
	"fmt"

	"github.com/AndreySenov/meridian/v2/async"
)

func ExampleFuture_Get() {
	p := async.NewPromise[int]()
	f := p.Future()

	// Get reports the context error if the Promise is still pending.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.Get(ctx)
	fmt.Println("pending:", err)

	p.Resolve(42)

	value, err := f.Get(context.Background())
	fmt.Println("completed:", value, err)

	// Output:
	// pending: context canceled
	// completed: 42 <nil>
}

func ExampleFuture_Done() {
	p := async.NewPromise[int]()
	f := p.Future()

	p.Resolve(42)

	// Done costs nothing to wait on: it starts no goroutine and allocates
	// nothing, which makes it a natural fit for a select.
	select {
	case <-f.Done():
		value, _ := f.Get(context.Background())
		fmt.Println("completed:", value)
	case <-context.Background().Done():
		fmt.Println("cancelled")
	}

	// Output: completed: 42
}

func ExampleFuture_IsDone() {
	p := async.NewPromise[int]()
	f := p.Future()

	// IsDone checks the result without blocking.
	fmt.Println("pending:", f.IsDone())

	p.Resolve(42)

	fmt.Println("completed:", f.IsDone())

	// Output:
	// pending: false
	// completed: true
}

func ExampleFuture_OnComplete() {
	p := async.NewPromise[int]()
	f := p.Future()

	// Handlers run in registration order once the Promise is completed.
	f.OnComplete(func(value int, err error) {
		fmt.Println("first:", value, err)
	})
	f.OnComplete(func(value int, err error) {
		fmt.Println("second:", value, err)
	})

	p.Resolve(42)

	// A handler registered after completion runs immediately.
	f.OnComplete(func(value int, err error) {
		fmt.Println("third:", value, err)
	})

	// Output:
	// first: 42 <nil>
	// second: 42 <nil>
	// third: 42 <nil>
}

func ExampleFuture_OnComplete_cancel() {
	p := async.NewPromise[int]()
	f := p.Future()

	cancel := f.OnComplete(func(value int, _ error) {
		fmt.Println("never runs:", value)
	})

	f.OnComplete(func(value int, _ error) {
		fmt.Println("still registered:", value)
	})

	// Unregister the first handler while the Promise is still pending.
	// cancel reports whether it stopped the handler from being run.
	fmt.Println("stopped:", cancel())

	p.Resolve(42)

	// Output:
	// stopped: true
	// still registered: 42
}

func ExampleFuture_IsShared() {
	p := async.NewPromise[int]()
	f1 := p.Future()

	fmt.Println("one handle:", f1.IsShared())

	f2 := p.Future()

	fmt.Println("two handles:", f1.IsShared(), f2.IsShared())

	// Output:
	// one handle: false
	// two handles: true true
}
