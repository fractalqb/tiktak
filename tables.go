package main

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type tableCol struct {
	title string
	width int
}

func (tc tableCol) Width() int {
	if tc.width == 0 {
		return utf8.RuneCountInString(tc.title)
	}
	return tc.width
}

func tableHRule(wr io.Writer, prefix string, cols ...tableCol) {
	fmt.Fprint(wr, prefix, "+")
	for _, c := range cols {
		w := c.Width()
		fmt.Fprint(wr, strings.Repeat("-", w+2))
		fmt.Fprint(wr, "+")
	}
	fmt.Fprintln(wr)
}

func tableStartRow(wr io.Writer, prefix string) {
	fmt.Fprint(wr, prefix, "|")
}

func tableCell(wr io.Writer, width int, text string) {
	algr := width < 0
	if algr {
		width = -width
	}
	str := []rune(text)
	wr.Write([]byte{' '})
	if len(str) > width {
		fmt.Fprint(wr, string(str[:width-1])+"…")
	} else if algr {
		fmt.Fprint(wr, strings.Repeat(" ", width-len(str)))
		fmt.Fprint(wr, text)
	} else {
		fmt.Fprint(wr, text)
		fmt.Fprint(wr, strings.Repeat(" ", width-len(str)))
	}
	fmt.Fprintf(wr, " |")
}

func tableSpanWidth(widths ...int) int {
	l := len(widths)
	switch l {
	case 0:
		return 0
	case 1:
		return widths[0]
	}
	res := widths[0] + widths[l-1] + 3
	for _, w := range widths[1 : l-1] {
		res += w + 3
	}
	return res
}

func tableColsWidth(cols ...tableCol) int {
	ws := make([]int, len(cols))
	for i, c := range cols {
		ws[i] = c.Width()
	}
	return tableSpanWidth(ws...)
}

func tableHead(wr io.Writer, prefix string, cols ...tableCol) {
	tableStartRow(wr, prefix)
	for _, c := range cols {
		tableCell(wr, c.Width(), c.title)
	}
	fmt.Fprintln(wr)
}
