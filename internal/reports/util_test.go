package reports

import "fmt"

func ExampleSpanID() {
	for i := 2; i < 65537; i = i * i {
		fmt.Println(i, SpanID(i))
	}
	// Output:
	// 2 2
	// 4 4
	// 16 G
	// 256 74
	// 65536 1EKG
}

func ExampleParseSpanID() {
	for _, id := range []string{"2", "4", "G", "74", "1EkG"} {
		fmt.Println(ParseSpanID(id))
	}
	// Output:
	// 2 <nil>
	// 4 <nil>
	// 16 <nil>
	// 256 <nil>
	// 65536 <nil>
}
