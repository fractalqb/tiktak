package txtab

import (
	"os"
)

func ExampleWriter() {
	wr := Writer{
		F: &Table{
			Cols: []Column{
				Column{Title: "Foo", With: 6},
				Column{Title: "#", With: 5, Align: Center},
				Column{Title: "Baz", With: 6, Align: Right},
			},
		},
		W: os.Stdout,
	}
	wr.Header()
	wr.Hrule()
	wr.RowStart()
	wr.Cell("This is too long")
	wr.Cell("is")
	wr.Cell("This is too long")
	wr.RowEnd()
	// Output:
	// | Foo    |   #   |    Baz |
	// +--------+-------+--------+
	// | This … |  is   | … long |
}
