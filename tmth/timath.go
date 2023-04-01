package tmth

import (
	"time"
)

func AddDay(t time.Time, add int, loc *time.Location) time.Time {
	if loc == nil {
		loc = t.Location()
	}
	y, m, d := t.Date()
	ch, cm, cs := t.Clock()
	nano := t.Nanosecond()
	return time.Date(y, m, d+add, ch, cm, cs, nano, loc)
}

func StartDay(t time.Time, add int, loc *time.Location) time.Time {
	if loc == nil {
		loc = t.Location()
	}
	y, m, d := t.Date()
	return time.Date(y, m, d+add, 0, 0, 0, 0, loc)
}

// LastDay retuns Time from shifted to the last Weekday wd not after from.
func LastDay(wd time.Weekday, from time.Time, loc *time.Location) time.Time {
	dd := int(wd - from.Weekday())
	if dd > 0 {
		dd -= 7
	}
	return AddDay(from, dd, loc)
}

// NextDay returns Time from shited to the next Weekday wd after from.
func NextDay(wd time.Weekday, from time.Time, loc *time.Location) time.Time {
	dd := int(wd - from.Weekday())
	if dd <= 0 {
		dd += 7
	}
	return AddDay(from, dd, loc)
}

func StartMonth(t time.Time, add int, loc *time.Location) time.Time {
	if loc == nil {
		loc = t.Location()
	}
	y, m, _ := t.Date()
	return time.Date(y, m+time.Month(add), 1, 0, 0, 0, 0, loc)
}
