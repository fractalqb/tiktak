package filters

import (
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type DEArbZG struct{}

func (f *DEArbZG) Flags(args []string) ([]string, error) {
	return args, nil
}

func (f *DEArbZG) Filter(tl *tiktak.TimeLine, now time.Time) error {
	return nil
}
