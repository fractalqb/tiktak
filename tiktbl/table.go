package tiktbl

import (
	"fmt"
	"io"
	"math"
	"strings"
	"unicode/utf8"
)

type Writer interface {
	Write(w io.Writer, t *Data) error
}

type CellOpt interface{ applyToCell(*cell) error }

type Align int

func (a Align) applyToCell(c *cell) error {
	c.align = a
	return nil
}

const (
	Left Align = 1 + iota
	Center
	Right
)

type Span int

const SpanAll Span = -1

func (s Span) applyToCell(c *cell) error {
	c.span = int(s)
	return nil
}

type Pad rune

func (p Pad) applyToCell(c *cell) error {
	c.pad = rune(p)
	return nil
}

type Data struct {
	cols []Align
	rows []row
}

func (t *Data) Align(a Align, cs ...int) {
	for _, c := range cs {
		if c >= len(t.cols) {
			if c < cap(t.cols) {
				t.cols = t.cols[:c+1]
			} else {
				tmp := make([]Align, c+1)
				copy(tmp, t.cols)
				t.cols = tmp
			}
		}
		t.cols[c] = a
	}
}

func (t *Data) Set(r, c int, data any, opts ...CellOpt) error {
	switch data := data.(type) {
	case string:
		return t.SetString(r, c, data, opts...)
	case fmt.Stringer:
		return t.SetString(r, c, data.String(), opts...)
	default:
		return t.SetString(r, c, fmt.Sprint(data), opts...)
	}
}

func (t *Data) SetString(r, c int, text string, opts ...CellOpt) error {
	if r >= len(t.rows) {
		if r < cap(t.rows) {
			t.rows = t.rows[:r+1]
		} else {
			tmp := make([]row, r+1)
			copy(tmp, t.rows)
			t.rows = tmp
		}
	}
	rou := t.rows[r]
	if c >= len(rou) {
		if c < cap(rou) {
			rou = rou[:c+1]
		} else {
			tmp := make(row, c+1)
			copy(tmp, rou)
			rou = tmp
		}
		t.rows[r] = rou
	}
	cl := &rou[c]
	cl.text = text
	for _, opt := range opts {
		if err := opt.applyToCell(cl); err != nil {
			return err
		}
	}
	if cl.span == 0 {
		cl.span = 1
	}
	t.width(c, cl.span)
	if cl.align != 0 && t.cols[c] == 0 {
		t.cols[c] = cl.align
	}
	return nil
}

func (t *Data) Columns() int { return len(t.cols) }

func (t *Data) ColumnWidths(pad string) (cws []int) {
	cws = make([]int, len(t.cols))
	for _, row := range t.rows {
		for c := range row {
			cell := &row[c]
			if cell.span < 1 || cell.text == "" {
				continue
			}
			w := utf8.RuneCountInString(cell.text)
			if w > cws[c] {
				cws[c] = w
			}
		}
	}
	// Heuristics to avoid linear optimization
	cpw := utf8.RuneCountInString(pad)
	for _, row := range t.rows {
		for c := range row {
			cell := &row[c]
			if cell.span < 0 {
				w := utf8.RuneCountInString(cell.text)
				if l := len(cws) - c; l > 1 {
					w -= (l - 1) * cpw
				}
				s := 0
				for _, c := range cws[c:] {
					s += c
				}
				if w > s {
					stretch(cws[c:], w)
				}
			} else if sp := cell.span; sp > 1 {
				w := utf8.RuneCountInString(cell.text)
				w -= (sp - 1) * cpw
				s := 0
				for _, c := range cws[c : c+sp] {
					s += c
				}
				if w > s {
					stretch(cws[c:c+sp], w)
				}
			}
		}
	}
	return cws
}
func stretch(cws []int, total int) {
	l := len(cws)
	switch l {
	case 0:
		return
	case 1:
		cws[0] = total
		return
	}
	var scale float64
	for _, cw := range cws {
		scale += float64(cw)
	}
	scale = float64(total) / scale
	breaks := make([]float64, l)
	breaks[0] = scale * float64(cws[0])
	for i := 1; i < l; i++ {
		breaks[i] = breaks[i-1] + scale*float64(cws[i])
	}
	if t := math.Trunc(breaks[l-1]); t < breaks[l-1] {
		breaks[l-1] = t + 1
	}
	last := 0.0
	for i, b := range breaks {
		b = math.Round(b)
		cws[i] = int(b - last)
		last = b
	}
}

func (t *Data) width(col, span int) {
	if span > 1 {
		col += span
	} else {
		col++
	}
	if col > len(t.cols) {
		tmp := make([]Align, col)
		copy(tmp, t.cols)
		t.cols = tmp
	}
}

type cell struct {
	text   string
	span   int
	align  Align
	pad    rune
	styler Styler
}

func (c *cell) write(w io.Writer, cw int, a Align) error {
	if cw == 0 {
		return nil
	}
	rn := utf8.RuneCountInString(c.text)
	if rn > cw {
		txt := []rune(c.text)
		_, err := io.WriteString(w, string(txt[:cw-1])+"…")
		return err
	}
	if c.align != 0 {
		a = c.align
	}
	pad := " "
	if c.pad != 0 {
		pad = string(c.pad)
	}
	cw -= rn
	txt := c.text
	if c.styler != nil {
		txt = c.styler.style(txt)
	}
	switch a {
	case Right:
		if _, err := io.WriteString(w, strings.Repeat(pad, cw)); err != nil {
			return err
		}
		if _, err := io.WriteString(w, txt); err != nil {
			return err
		}
	case Center:
		rn := cw / 2
		cw -= rn
		if _, err := io.WriteString(w, strings.Repeat(pad, rn)); err != nil {
			return err
		}
		if _, err := io.WriteString(w, txt); err != nil {
			return err
		}
		if _, err := io.WriteString(w, strings.Repeat(pad, cw)); err != nil {
			return err
		}
	default:
		if _, err := io.WriteString(w, txt); err != nil {
			return err
		}
		if _, err := io.WriteString(w, strings.Repeat(pad, cw)); err != nil {
			return err
		}
	}
	return nil
}

type row []cell
