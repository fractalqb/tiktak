package reports

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"git.fractalqb.de/fractalqb/tetrta"
	"git.fractalqb.de/fractalqb/tiktak"
)

type Report struct {
	Layout tetrta.TableWriter
	Fmts   Formats
}

type Formats interface {
	Date(time.Time) string
	ShortDate(time.Time) string
	Clock(time.Time) string
	Duration(time.Duration) string
}

const (
	DateFmt      = "Mon, 02 Jan 2006"
	ShortDateFmt = "Mon, 02 Jan"
)

var (
	MinutesFmts = cfgFmts{
		fullDateFmt:  DateFmt,
		shortDateFmt: ShortDateFmt,
		clockFmt:     "15:04",
		durFmtr:      minutes,
	}

	SecondsFmts = cfgFmts{
		fullDateFmt:  DateFmt,
		shortDateFmt: ShortDateFmt,
		clockFmt:     "15:04:05",
		durFmtr:      fmtDuration,
	}

	FracCFmts = cfgFmts{
		fullDateFmt:  DateFmt,
		shortDateFmt: ShortDateFmt,
		clockFmt:     "15:04",
		durFmtr: func(d time.Duration) string {
			return fmt.Sprintf("%.2f", float64(d.Seconds())/(60*60))
		},
	}
)

type cfgFmts struct {
	fullDateFmt  string
	shortDateFmt string
	clockFmt     string
	durFmtr      func(time.Duration) string
}

func (f cfgFmts) Date(t time.Time) string { return t.Format(f.fullDateFmt) }

func (f cfgFmts) ShortDate(t time.Time) string { return t.Format(f.shortDateFmt) }

func (f cfgFmts) Clock(t time.Time) string { return t.Format(f.clockFmt) }

func (f cfgFmts) Duration(d time.Duration) string { return f.durFmtr(d) }

func minutes(d time.Duration) string {
	d = d.Round(time.Minute)
	h, m, s, f := tiktak.HMSF(d)
	sub := (float64(s) + f) / 60
	m = int(math.Round(float64(m) + sub))
	res := fmt.Sprintf("%02d:%02d", h, m)
	return res
}

func fmtDuration(d time.Duration) (res string) {
	if d < 0 {
		return "∞"
	}
	h, m, s, f := tiktak.HMSF(d)
	days := h / 24
	h %= 24
	if f != 0 {
		fstr := strconv.FormatFloat(f, 'f', 3, 64)
		res = fmt.Sprintf(`%02d'%02d.%s"`, m, s, fstr[2:])
	} else if s != 0 {
		res = fmt.Sprintf(`%02d'%02d"`, m, s)
	} else {
		res = fmt.Sprintf(`%02d'`, m)
	}
	if h != 0 {
		res = fmt.Sprintf("%02d:%s", h, res)
	}
	if days != 0 {
		res = fmt.Sprintf("%dd%s", days, res)
	}
	return res
}
