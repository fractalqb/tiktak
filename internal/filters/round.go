package filters

import (
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type Round struct {
	Precision time.Duration `tikf:"p,Precision to round timestamps to"`
}

func (f *Round) Filter(tl *tiktak.TimeLine) error {
	for i, sw := range *tl {
		rt := sw.When().Round(f.Precision)
		if err := tl.Reschedule(i, rt); err != nil {
			return err
		}
	}
	return nil
}
