package txtab

import (
	"io"
)

type UnknownProp struct {
	Index int
	Prop  interface{}
}

type UnknownPropError []UnknownProp

func (upe UnknownPropError) Error() string {
	return "unknown properties" // TODO
}

type Formatter interface {
	AddColumn(title string, props ...interface{}) error
	ColumnNo() int
	RowStart(wr io.Writer) error
	RowEnd(wr io.Writer) error
	Header(wr io.Writer) error
	Hrule(wr io.Writer) error
	Cell(wr io.Writer, width int, v interface{}, props ...interface{}) error
	Cells(wr io.Writer, colFrom, colTo int, v interface{}, props ...interface{}) error
}

type Writer struct {
	F   Formatter
	W   io.Writer
	col int
}

func (w *Writer) AddColumn(title string, props ...interface{}) error {
	return w.F.AddColumn(title, props...)
}

func (w *Writer) RowStart() error {
	w.col = 0
	return w.F.RowStart(w.W)
}

func (w *Writer) RowEnd() error { return w.F.RowEnd(w.W) }
func (w *Writer) Header() error { return w.F.Header(w.W) }
func (w *Writer) Hrule() error  { return w.F.Hrule(w.W) }

func (w *Writer) Cell(v interface{}, props ...interface{}) error {
	err := w.F.Cell(w.W, w.col, v, props...)
	if err != nil {
		return err
	}
	w.col++
	return nil
}

func (w *Writer) Cells(merge int, v interface{}, props ...interface{}) error {
	switch merge {
	case 0:
		return nil
	case -1:
		merge = w.F.ColumnNo()
	}
	err := w.F.Cells(w.W, w.col, w.col+merge-1, v, props...)
	if err != nil {
		return err
	}
	w.col += merge
	return nil
}
