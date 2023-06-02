package tiktbl

import (
	"os"
	"testing"
	"testing/quick"
)

func Example() {
	var tbl Data
	tbl.Set(0, 0, "foo")
	tbl.Set(0, 1, "bar")
	tbl.Set(1, 0, "", Span(-1), Pad('-'))
	tbl.Set(2, 0, "baz", Span(2), Center, Pad('.'))
	(&Terminal{CellPad: " | "}).Write(os.Stdout, &tbl)
	// Output:
	// foo | bar
	// ---------
	// ...baz...
}

func ExampleCrsr() {
	var tbl Data
	crsr := tbl.At(0, 0).SetString("2-column heading", Span(2), Pad('.')).NextRow()
	crsr.SetString("", SpanAll, Pad('-')).NextRow()
	crsr.SetStrings("col1", "col2")
	(&Terminal{CellPad: " | "}).Write(os.Stdout, &tbl)
	// Output:
	// 2-column heading
	// ----------------
	// col1    | col2
}

func ExampleStretchEmpty() {
	var tbl Data
	crsr := tbl.At(0, 0).
		SetString("head").
		SetString("2-column heading", Span(2), Pad('.')).NextRow()
	crsr.SetString("", SpanAll, Pad('-')).NextRow()
	crsr.SetStrings("col1")
	(&Terminal{CellPad: " | "}).Write(os.Stdout, &tbl)
	// Output:
	// head | 2-column heading
	// -----------------------
	// col1 |         |
}

func Test_stretch(t *testing.T) {
	proto := []int{4, 7, 5}
	cols := make([]int, len(proto))
	f := func(w int) bool {
		if w %= 100; w < 0 {
			w = -w
		}
		copy(cols, proto)
		stretch(cols, w)
		sum := 0
		for _, c := range cols {
			sum += c
		}
		return sum == w
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
