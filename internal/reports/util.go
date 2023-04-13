package reports

import (
	"fmt"
	"strings"
	"time"
	"unicode"

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

const numChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func SpanID(idx int) string {
	if idx < 0 {
		return ""
	}
	if idx == 0 {
		return string(numChars[0])
	}
	var ds []byte
	for idx > 0 {
		d := idx % len(numChars)
		ds = append(ds, numChars[d])
		idx /= len(numChars)
	}
	for i, j := 0, len(ds)-1; i < j; i, j = i+1, j-1 {
		ds[i], ds[j] = ds[j], ds[i]
	}
	return string(ds)
}

func ParseSpanID(s string) (int, error) {
	res := 0
	for _, d := range s {
		n := strings.IndexRune(numChars, unicode.ToUpper(d))
		if n < 0 {
			return 0, fmt.Errorf("invalid rune '%c' in span ID", d)
		}
		res = len(numChars)*res + n
	}
	return res, nil
}
