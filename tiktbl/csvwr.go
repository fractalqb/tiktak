package tiktbl

import (
	"fmt"
	"io"
)

type CSV struct {
	FS             string
	SkipEmptyLines bool
}

func (tw *CSV) Write(w io.Writer, t *Data) error {
	for _, row := range t.rows {
		empty := true
		for _, cell := range row {
			if cell.text != "" {
				empty = false
				break
			}
		}
		if !empty || !tw.SkipEmptyLines {
			c := 0
			for c < len(t.cols) {
				if c > 0 {
					io.WriteString(w, tw.FS)
				}
				if c < len(row) {
					io.WriteString(w, row[c].text)
				}
				c++
			}
			fmt.Fprintln(w)
		}
	}
	return nil
}
