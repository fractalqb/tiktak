package reports

import (
	"os"

	"git.fractalqb.de/fractalqb/tetrta"
	"github.com/TwiN/go-color"
)

func init() {
	ostat, err := os.Stdout.Stat()
	if err != nil {
		color.Toggle(false)
		return
	}
	if ostat.Mode()&os.ModeCharDevice == 0 {
		color.Toggle(false)
	} else {
		color.Toggle(true)
	}
}

func Bold() tetrta.Styler {
	return tetrta.Style(func(s string) string { return color.InBold(s) })
}

func Underline() tetrta.Styler {
	return tetrta.Style(func(s string) string { return color.InUnderline(s) })
}

func Muted() tetrta.Styler {
	return tetrta.Style(func(s string) string { return color.InCyan(s) })
}

func Warn() tetrta.Styler {
	return tetrta.Style(func(s string) string { return color.OverYellow(s) })
}
