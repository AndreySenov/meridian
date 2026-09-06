package meridian_test

import (
	"context"
	"fmt"

	"github.com/AndreySenov/meridian"
)

func ExampleSingleFlight_Do() {
	var flights meridian.SingleFlight[string, int]

	calls := 0
	task := func() (int, error) {
		calls++
		return 42, nil
	}

	// The second call joins the first one instead of running the task again.
	f1 := flights.Do("key", task)
	f2 := flights.Do("key", task)

	value, err := f1.Get(context.Background())

	fmt.Println("result:", value, err)
	fmt.Println("task runs:", calls)
	fmt.Println("shared:", f2.IsShared())

	// Output:
	// result: 42 <nil>
	// task runs: 1
	// shared: true
}

func ExampleSingleFlight_Forget() {
	var flights meridian.SingleFlight[string, int]

	calls := 0
	task := func() (int, error) {
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
