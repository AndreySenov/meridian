package async_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/AndreySenov/meridian/v2/async"
)

func ExamplePromise() {
	p := async.NewPromise[int]()

	// The producer completes the Promise on its own goroutine, while the
	// consumer waits for the result on another one.
	go func() {
		p.Resolve(42) // or p.Reject(err) / p.Complete(value, err)
	}()

	value, err := p.Future().Get(context.Background())

	fmt.Println(value, err)
	// Output: 42 <nil>
}

func ExamplePromise_Resolve() {
	p := async.NewPromise[int]()

	// Resolve is the success shorthand: the result carries no error.
	p.Resolve(42)

	value, err := p.Future().Get(context.Background())

	fmt.Println(value, err)
	// Output: 42 <nil>
}

func ExamplePromise_Reject() {
	p := async.NewPromise[int]()

	// Reject is the failure shorthand: the result carries the zero value
	// along with the error.
	p.Reject(errors.New("not found"))

	value, err := p.Future().Get(context.Background())

	fmt.Println(value, err)
	// Output: 0 not found
}

func ExamplePromise_Complete() {
	p := async.NewPromise[int]()

	p.Complete(42, nil)

	// Only the first completion takes effect; later ones are no-ops.
	p.Complete(0, errors.New("too late"))

	value, err := p.Future().Get(context.Background())

	fmt.Println(value, err)
	// Output: 42 <nil>
}
