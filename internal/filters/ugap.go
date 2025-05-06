package filters

import (
	"flag"
	"fmt"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type MicroGap struct {
	Gap time.Duration
}

func (f *MicroGap) Flags(args []string) ([]string, error) {
	fs := flag.NewFlagSet("ugap", flag.ContinueOnError)
	fs.DurationVar(&f.Gap, "gap", f.Gap, "Minimum duration to no be a µ-gap")
	err := fs.Parse(args)
	return fs.Args(), err
}

func (f *MicroGap) Filter(tl *tiktak.TimeLine, _ time.Time) error {
	for _, sw := range *tl {
		d := sw.Duration()
		if d >= 0 && d < f.Gap {
			sw.FilterNotes(func(n tiktak.Note) bool { return n.Sym != 'µ' })
			sw.AddWarning('µ', fmt.Sprintf("µ-Gap %s < %s", d, f.Gap))
		}
	}
	return nil
}
