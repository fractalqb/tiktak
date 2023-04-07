package tiktak

import (
	"math"
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

// LastDay returns Time from shifted to the last Weekday wd not after from.
func LastDay(wd time.Weekday, from time.Time, loc *time.Location) time.Time {
	dd := int(wd - from.Weekday())
	if dd > 0 {
		dd -= 7
	}
	return AddDay(from, dd, loc)
}

// NextDay returns Time from shifted to the next Weekday wd after from.
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

func HMSF(d time.Duration) (h, m, s int, f float64) {
	h = int(math.Trunc(d.Hours()))
	d -= time.Duration(h) * time.Hour
	m = int(math.Trunc(d.Minutes()))
	d -= time.Duration(m) * time.Minute
	f = d.Seconds()
	s = int(math.Trunc(f))
	f -= float64(s)
	return
}

type Date struct {
	Year     int
	Month    time.Month
	Day      int
	Location *time.Location
}

func DateOf(t time.Time) (res Date) {
	res.Year, res.Month, res.Day = t.Date()
	res.Location = t.Location()
	return
}

func (l *Date) Compare(r *Date) (d int) {
	if d = l.Year - r.Year; d != 0 {
		return d
	}
	if d = int(l.Month - r.Month); d != 0 {
		return d
	}
	return l.Day - r.Day
}

func (d *Date) Start() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, d.Location)
}

type Clock struct {
	D        time.Duration
	Location *time.Location
}

func ClockOf(t time.Time) Clock {
	s := StartDay(t, 0, nil)
	return Clock{
		D:        t.Sub(s),
		Location: t.Location(),
	}
}

func (c Clock) On(day time.Time) time.Time {
	s := StartDay(day, 0, day.Location())
	return s.Add(c.D) // TODO c.Location ?
}

func (c Clock) HMSF() (h, m, s int, f float64) { return HMSF(c.D) }
