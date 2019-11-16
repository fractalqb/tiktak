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

func colsWidth(tw tableWriter, cols ...tableCol) int {
	ws := make([]int, len(cols))
	for i, c := range cols {
		ws[i] = c.Width()
	}
	return tw.SpanWidth(ws...)
}

type tableWriter interface {
	Head(cols ...tableCol)
	HRule(cols ...tableCol)
	StartRow()
	Cell(width int, text string)
	SpanWidth(widths ...int) int
}

type borderedWriter struct {
	wr     io.Writer
	prefix string
}

func (tw borderedWriter) Head(cols ...tableCol) {
	tw.StartRow()
	for _, c := range cols {
		tw.Cell(c.Width(), c.title)
	}
	fmt.Fprintln(tw.wr)
}

func (tw borderedWriter) HRule(cols ...tableCol) {
	fmt.Fprint(tw.wr, tw.prefix, "+")
	for _, c := range cols {
		w := c.Width()
		fmt.Fprint(tw.wr, strings.Repeat("-", w+2))
		fmt.Fprint(tw.wr, "+")
	}
	fmt.Fprintln(tw.wr)
}

func (tw borderedWriter) StartRow() {
	fmt.Fprint(tw.wr, tw.prefix, "|")
}

func (tw borderedWriter) Cell(width int, text string) {
	algr := width < 0
	if algr {
		width = -width
	}
	str := []rune(text)
	tw.wr.Write([]byte{' '})
	if len(str) > width {
		fmt.Fprint(tw.wr, string(str[:width-1])+"…")
	} else if algr {
		fmt.Fprint(tw.wr, strings.Repeat(" ", width-len(str)))
		fmt.Fprint(tw.wr, text)
	} else {
		fmt.Fprint(tw.wr, text)
		fmt.Fprint(tw.wr, strings.Repeat(" ", width-len(str)))
	}
	fmt.Fprintf(tw.wr, " |")
}

func (tw borderedWriter) SpanWidth(widths ...int) int {
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
