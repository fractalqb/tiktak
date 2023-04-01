package tmth

import (
	"fmt"
	"time"
)

func ExampleAddDay() {
	t := time.Date(1999, time.June, 25, 12, 33, 9, 0, time.UTC)
	fmt.Println(AddDay(t, 1, nil))
	fmt.Println(AddDay(t, 10, nil))
	fmt.Println(AddDay(t, -25, nil))
	// Output:
	// 1999-06-26 12:33:09 +0000 UTC
	// 1999-07-05 12:33:09 +0000 UTC
	// 1999-05-31 12:33:09 +0000 UTC
}

func ExampleStartDay() {
	t := time.Date(1999, time.June, 25, 12, 33, 9, 345876, time.UTC)
	fmt.Println(StartDay(t, 0, nil))
	fmt.Println(StartDay(t, -1, nil))
	fmt.Println(StartDay(t, 1, nil))
	// Output:
	// 1999-06-25 00:00:00 +0000 UTC
	// 1999-06-24 00:00:00 +0000 UTC
	// 1999-06-26 00:00:00 +0000 UTC
}

func ExampleLastDay() {
	t := time.Date(1999, time.June, 25, 12, 33, 9, 0, time.UTC)
	fmt.Println(LastDay(time.Friday, t, nil))
	fmt.Println(LastDay(time.Monday, t, nil))
	fmt.Println(LastDay(time.Sunday, t, nil))
	fmt.Println(LastDay(time.Saturday, t, nil))
	// Output:
	// 1999-06-25 12:33:09 +0000 UTC
	// 1999-06-21 12:33:09 +0000 UTC
	// 1999-06-20 12:33:09 +0000 UTC
	// 1999-06-19 12:33:09 +0000 UTC
}

func ExampleNextDay() {
	t := time.Date(1999, time.June, 25, 12, 33, 9, 0, time.UTC)
	fmt.Println(NextDay(time.Sunday, t, nil))
	fmt.Println(NextDay(time.Monday, t, nil))
	fmt.Println(NextDay(time.Friday, t, nil))
	// Output:
	// 1999-06-27 12:33:09 +0000 UTC
	// 1999-06-28 12:33:09 +0000 UTC
	// 1999-07-02 12:33:09 +0000 UTC
}

func ExampleStartMonth() {
	t := time.Date(1999, time.June, 25, 12, 33, 9, 4711, time.UTC)
	fmt.Println(StartMonth(t, 0, nil))
	fmt.Println(StartMonth(t, -1, nil))
	fmt.Println(StartMonth(t, 1, nil))
	// Output:
	// 1999-06-01 00:00:00 +0000 UTC
	// 1999-05-01 00:00:00 +0000 UTC
	// 1999-07-01 00:00:00 +0000 UTC
}
