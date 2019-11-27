package txtab

import (
	"os"
)

func ExampleTable() {
	tab := Table{
		Cols: []Column{
			Column{Title: "Foo", With: 6},
			Column{Title: "#", With: 5, Align: Center},
			Column{Title: "Baz", With: 6, Align: Right},
		},
	}
	tab.Header(os.Stdout)
	tab.Hrule(os.Stdout)
	tab.RowStart(os.Stdout)
	tab.Cell(os.Stdout, 0, "This is too long")
	tab.Cell(os.Stdout, 1, "so is this")
	tab.Cell(os.Stdout, 2, "This is too long")
	// Output:
	// | Foo    |   #   |    Baz |
	// +--------+-------+--------+
	// | This … | so i… | … long |
}
