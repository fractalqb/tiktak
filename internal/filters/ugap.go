package filters

import (
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

type MicroGap struct {
	Gap time.Duration `tikf:",Minimum duration to no be a µ-gap"`
}

func (f *MicroGap) Filter(tl *tiktak.TimeLine) error {
	i := 0
	for i < len(*tl) {
		sw := (*tl)[i]
		d := sw.Duration()
		if d < 0 || d >= f.Gap {
			i++
			continue
		}
		tl.DelSwitch(i)
		tl.Switch(sw.When().Add(d/2), (*tl)[i].Task())
	}
	return nil
}
