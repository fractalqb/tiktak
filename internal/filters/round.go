package filters

import (
	"flag"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type Round struct {
	Precision time.Duration
}

func (f *Round) Flags(args []string) ([]string, error) {
	fs := flag.NewFlagSet("round", flag.ContinueOnError)
	fs.DurationVar(&f.Precision, "p", f.Precision, "Precision to round timestamps to")
	err := fs.Parse(args)
	return fs.Args(), err
}

func (f *Round) Filter(tl *tiktak.TimeLine, _ time.Time) error {
	for i, sw := range *tl {
		rt := sw.When().Round(f.Precision)
		if err := tl.Reschedule(i, rt); err != nil {
			return err
		}
	}
	return nil
}
