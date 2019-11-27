package txtab

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Alignment int

const (
	Left Alignment = iota
	Right
	Center
)

type ClipFunc = func(s string, width int) string

type Column struct {
	Title string
	With  int
	Align Alignment
	Clip  ClipFunc
}

func (c *Column) Write(wr io.Writer, s string, props ...interface{}) error {
	alng, clip := c.evalProps(props)
	return wrCells(wr, s, c.With, alng, clip)
}

func (c *Column) evalProps(props []interface{}) (Alignment, ClipFunc) {
	alng, clip := c.Align, c.Clip
	for _, p := range props {
		switch prop := p.(type) {
		case Alignment:
			alng = prop
		case ClipFunc:
			clip = prop
		}
	}
	return alng, clip
}

func ClipLeft(s string, w int) string {
	l := utf8.RuneCountInString(s)
	if l > w {
		rs := []rune(s)
		var sb strings.Builder
		sb.WriteRune('…')
		for _, r := range rs[l-w+1:] {
			sb.WriteRune(r)
		}
		s = sb.String()
	}
	return s
}

func ClipRight(s string, w int) string {
	l := utf8.RuneCountInString(s)
	if l > w {
		rs := []rune(s)
		var sb strings.Builder
		for _, r := range rs[:w-1] {
			sb.WriteRune(r)
		}
		sb.WriteRune('…')
		s = sb.String()
	}
	return s
}

type Table struct {
	Prefix string
	Cols   []Column
}

func (tab *Table) AddColumn(title string, props ...interface{}) error {
	var unkp UnknownPropError
	col := Column{Title: title}
	for i, p := range props {
		switch prop := p.(type) {
		case int:
			col.With = prop
		case Alignment:
			col.Align = prop
		case ClipFunc:
			col.Clip = prop
		default:
			unkp = append(unkp, UnknownProp{i, p})
		}
	}
	if col.With <= 0 {
		col.With = utf8.RuneCountInString(title)
	}
	tab.Cols = append(tab.Cols, col)
	if len(unkp) > 0 {
		return unkp
	}
	return nil
}

func (tab *Table) ColumnNo() int { return len(tab.Cols) }

func (tab *Table) ColsWidth(from, to int) int {
	num := to - from + 1
	if num < 1 {
		return 0
	}
	res := 3 * (num - 1)
	for idx := from; idx <= to; idx++ {
		res += tab.Cols[idx].With
	}
	return res
}

func (tab *Table) RowStart(wr io.Writer) error {
	_, err := fmt.Fprintf(wr, "%s|", tab.Prefix)
	return err
}

func (tab *Table) RowEnd(wr io.Writer) error {
	_, err := fmt.Fprintln(wr)
	return err
}

func (tab *Table) Header(wr io.Writer) (err error) {
	if len(tab.Cols) == 0 {
		return nil
	}
	if err = tab.RowStart(wr); err != nil {
		return err
	}
	for i := range tab.Cols {
		if err = tab.Cell(wr, i, tab.Cols[i].Title); err != nil {
			return err
		}
	}
	if err = tab.RowEnd(wr); err != nil {
		return err
	}
	return err
}

func (tab *Table) Hrule(wr io.Writer) (err error) {
	if len(tab.Cols) == 0 {
		return nil
	}
	if _, err = fmt.Fprintf(wr, tab.Prefix); err != nil {
		return err
	}
	_, err = fmt.Fprintf(wr, "+-%s", strings.Repeat("-", tab.Cols[0].With))
	if err != nil {
		return err
	}
	for _, col := range tab.Cols[1:] {
		_, err = fmt.Fprintf(wr, "-+-%s", strings.Repeat("-", col.With))
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(wr, "-+")
	return err
}

func (tab *Table) Cell(wr io.Writer, i int, v interface{}, props ...interface{}) (err error) {
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case fmt.Stringer:
		s = t.String()
	default:
		s = fmt.Sprint(v)
	}
	if _, err = fmt.Fprint(wr, " "); err != nil {
		return err
	}
	if err = tab.Cols[i].Write(wr, s, props...); err != nil {
		return err
	}
	_, err = fmt.Fprint(wr, " |")
	return err
}

func (tab *Table) Cells(wr io.Writer, colFrom, colTo int, v interface{}, props ...interface{}) (err error) {
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case fmt.Stringer:
		s = t.String()
	default:
		s = fmt.Sprint(v)
	}
	if _, err = fmt.Fprint(wr, " "); err != nil {
		return err
	}
	alng, clip := tab.Cols[colFrom].evalProps(props)
	err = wrCells(wr, s,
		tab.ColsWidth(colFrom, colTo),
		alng, clip)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(wr, " |")
	return err
}

func wrCells(wr io.Writer, s string, w int, a Alignment, clip ClipFunc) error {
	if clip != nil {
		s = clip(s, w)
	} else {
		if a == Right {
			s = ClipLeft(s, w)
		} else {
			s = ClipRight(s, w)
		}
	}
	l := utf8.RuneCountInString(s)
	dw := w - l
	switch a {
	case Left:
		if _, err := fmt.Fprint(wr, s); err != nil {
			return err
		}
		for dw > 0 {
			if _, err := fmt.Fprint(wr, " "); err != nil {
				return err
			}
			dw--
		}
	case Right:
		for dw > 0 {
			if _, err := fmt.Fprint(wr, " "); err != nil {
				return err
			}
			dw--
		}
		if _, err := fmt.Fprint(wr, s); err != nil {
			return err
		}
	case Center:
		dr := dw - (dw / 2)
		dw -= dr
		for dw > 0 {
			if _, err := fmt.Fprint(wr, " "); err != nil {
				return err
			}
			dw--
		}
		if _, err := fmt.Fprint(wr, s); err != nil {
			return err
		}
		for dr > 0 {
			if _, err := fmt.Fprint(wr, " "); err != nil {
				return err
			}
			dr--
		}
	default:
		return fmt.Errorf("unknonwn alignment: %v", a)
	}
	return nil
}
