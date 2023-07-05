package tiktak

import (
	"fmt"
	"os"
	"strings"
)

func ExampleRead() {
	ts, err := Read(strings.NewReader(`/3 Just to test titles
# Sat, 01 Apr 2023
2023-04-01T12:00:00Z /1
2023-04-01T13:00:00Z /2
	. A note
2023-04-01T11:00:00Z /3
	!? A warning with the question symbol '?'
2023-04-01T11:30:00Z /4
2023-04-01T12:30:00Z /5`), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	Write(os.Stdout, ts)
	// Output:
	// v1.0.0	tiktak time tracker
	// /1
	// /2
	// /3 Just to test titles
	// /4
	// /5
	// # Sat, 01 Apr 2023
	// 2023-04-01T11:00:00Z /3
	// 	!? A warning with the question symbol '?'
	// 2023-04-01T11:30:00Z /4
	// 2023-04-01T12:00:00Z /1
	// 2023-04-01T12:30:00Z /5
	// 2023-04-01T13:00:00Z /2
	// 	. A note
}
