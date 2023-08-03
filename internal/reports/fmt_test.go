package reports

import (
	"fmt"
	"time"
)

// Fix https://github.com/fractalqb/tiktak/issues/3
func Example_dur() {
	d := 48*time.Minute + 5*time.Second + 714286*time.Microsecond
	fmt.Println(fmtDuration(d))
	d = 8*time.Hour + 30*time.Minute + 8*time.Second + 571429*time.Microsecond
	fmt.Println(fmtDuration(d))
	// Output:
	// 48'05.714"
	// 08:30'08.571"
}
