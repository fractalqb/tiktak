package reports

import (
	"os"

	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
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

func Bold() tiktbl.Styler {
	return tiktbl.Style(func(s string) string { return color.InBold(s) })
}

func Underline() tiktbl.Styler {
	return tiktbl.Style(func(s string) string { return color.InUnderline(s) })
}

func Muted() tiktbl.Styler {
	return tiktbl.Style(func(s string) string { return color.InBlue(s) })
}

func Warn() tiktbl.Styler {
	return tiktbl.Style(func(s string) string { return color.OverYellow(s) })
}
