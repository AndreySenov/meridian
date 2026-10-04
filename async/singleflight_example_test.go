package async_test

import (
	"context"
	"fmt"

	"github.com/AndreySenov/meridian/v3/async"
)

func ExampleSingleFlight_Do() {
	var flights async.SingleFlight[string, int]

	calls := 0
	release := make(chan struct{})
	task := func(context.Context) (int, error) {
		<-release // keep the call in flight until the second Do joins it
		calls++
		return 42, nil
	}

	// The second call joins the first one instead of running the task again.
	f1 := flights.Do("key", task)
	f2 := flights.Do("key", task)
	close(release)

	value, err := f1.Get(context.Background())

	fmt.Println("result:", value, err)
	fmt.Println("task runs:", calls)
	fmt.Println("shared:", f2.IsShared())

	// Output:
	// result: 42 <nil>
	// task runs: 1
	// shared: true
}

func ExampleSingleFlight_Cancel() {
	var flights async.SingleFlight[string, int]

	started := make(chan struct{})
	f := flights.Do("key", func(ctx context.Context) (int, error) {
		close(started)
		<-ctx.Done() // the task stops as soon as its context is canceled

		return 0, ctx.Err()
	})
	<-started

	fmt.Println("canceled:", flights.Cancel("key"))

	// Every consumer is released at once, regardless of whether the task cooperates.
	value, err := f.Get(context.Background())

	fmt.Println("result:", value, err)
	// Output:
	// canceled: true
	// result: 0 context canceled
}

func ExampleSingleFlight_Forget() {
	var flights async.SingleFlight[string, int]

	calls := 0
	task := func(context.Context) (int, error) {
		calls++
		return calls, nil
	}

	f1 := flights.Do("key", task)
	value1, _ := f1.Get(context.Background())

	// Forget detaches the current call, so the next Do starts a fresh task.
	flights.Forget("key")

	f2 := flights.Do("key", task)
	value2, _ := f2.Get(context.Background())

	fmt.Println(value1, value2)
	// Output: 1 2
}
