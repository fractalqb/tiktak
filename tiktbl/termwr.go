package tiktbl

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Terminal struct {
	CellPad string
}

func (tw *Terminal) Write(w io.Writer, t *Data) error {
	cws := t.ColumnWidths(tw.CellPad)
	cpw := utf8.RuneCountInString(tw.CellPad)
	for _, row := range t.rows {
		c := 0
		for c < len(t.cols) {
			if c > 0 {
				if _, err := io.WriteString(w, tw.CellPad); err != nil {
					return err
				}
			}
			if c >= len(row) {
				if _, err := io.WriteString(w, strings.Repeat(" ", cws[c])); err != nil {
					return err
				}
				c++
				continue
			}
			cell := &row[c]
			cw := tw.cellWidth(t, cell, c, cws, cpw)
			if err := cell.write(w, cw, t.cols[c]); err != nil {
				return err
			}
			if cell.span > 0 {
				c += cell.span
			} else if cell.span < 0 {
				c = len(t.cols)
			} else {
				c++
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

func (tw *Terminal) cellWidth(t *Data, c *cell, ci int, cws []int, padWidth int) (w int) {
	if c.span == 0 || c.span == 1 {
		return cws[ci]
	}
	if c.span < 0 {
		w = (len(t.cols) - 1) * padWidth
		for _, cw := range cws {
			w += cw
		}
		return w
	}
	w = (c.span - 1) * padWidth
	for i := 0; i < c.span; i++ {
		w += cws[ci+i]
	}
	return w
}
