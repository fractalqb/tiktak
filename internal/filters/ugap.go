package filters

import (
	"fmt"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type MicroGap struct {
	Gap time.Duration `tikf:",Minimum duration to no be a µ-gap"`
}

func (f *MicroGap) Filter(tl *tiktak.TimeLine) error {
	for _, sw := range *tl {
		d := sw.Duration()
		if d >= 0 && d < f.Gap {
			sw.FilterNotes(func(n tiktak.Note) bool { return n.Sym != 'µ' })
			sw.AddWarning('µ', fmt.Sprintf("µ-Gap %s < %s", d, f.Gap))
		}
	}
	return nil
}
