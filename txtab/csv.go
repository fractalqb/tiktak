package txtab

import (
	"fmt"
	"io"
)

type Csv struct {
	FSep  string
	Cols  []string
	first bool
}

func (csv *Csv) AddColumn(title string, props ...interface{}) error {
	csv.Cols = append(csv.Cols, title)
	if len(props) == 0 {
		return nil
	}
	var err UnknownPropError
	for i, p := range props {
		err = append(err, UnknownProp{i, p})
	}
	return err
}

func (csv *Csv) ColumnNo() int { return len(csv.Cols) }

func (csv *Csv) RowStart(wr io.Writer) error {
	csv.first = true
	return nil
}

func (csv *Csv) RowEnd(wr io.Writer) error {
	_, err := fmt.Fprintln(wr)
	return err
}

func (csv *Csv) Header(wr io.Writer) (err error) {
	if csv.ColumnNo() == 0 {
		return nil
	}
	if _, err = fmt.Fprint(wr, csv.Cols[0]); err != nil {
		return err
	}
	for _, c := range csv.Cols[1:] {
		if _, err = fmt.Fprintf(wr, "%s%s", csv.FSep, c); err != nil {
			return err
		}
	}
	return csv.RowEnd(wr)
}

func (csv *Csv) Hrule(wr io.Writer) error { return nil }

func (csv *Csv) Cell(wr io.Writer, width int, v interface{}, props ...interface{}) (err error) {
	if !csv.first {
		if _, err = fmt.Fprint(wr, csv.FSep); err != nil {
			return err
		}
	} else {
		csv.first = false
	}
	if v != nil {
		_, err = fmt.Fprint(wr, v)
	}
	return err
}

func (csv *Csv) Cells(wr io.Writer, colFrom, colTo int, v interface{}, props ...interface{}) error {
	for colFrom <= colTo {
		if err := csv.Cell(wr, 0, v, props...); err != nil {
			return err
		}
		colFrom++
	}
	return nil
}
