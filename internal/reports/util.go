package reports

import (
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

const empty = "-"

func hasWarning(tl tiktak.TimeLine, s, e time.Time, f func(*tiktak.Switch) bool) bool {
	if len(tl) == 0 {
		return false
	}
	_, sw := tl.Pick(s)
	if sw == nil && s.Before(tl[0].When()) {
		sw = tl[0]
	}
	if sw.When().Before(s) {
		sw = sw.Next()
	}
	var nis []int
	for sw != nil && sw.When().Before(e) {
		if f(sw) {
			nis = sw.SelectNotes(nis[:0], func(n tiktak.Note) bool { return n.Sym != 0 })
			if len(nis) > 0 {
				return true
			}
		}
		sw = sw.Next()
	}
	return false
}
