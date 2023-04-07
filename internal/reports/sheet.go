package reports

import (
	"fmt"
	"io"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
)

type Sheet struct {
	Report
	WeekStart time.Weekday
	Tasks     []*tiktak.Task
}

type tsum struct {
	n int
	d time.Duration
}

func accounts(tasks []*tiktak.Task) (sums []tsum, tmap map[*tiktak.Task]*tiktak.Task) {
	if len(tasks) == 0 {
		return
	}
	sums = make([]tsum, len(tasks))
	tmap = make(map[*tiktak.Task]*tiktak.Task)
	tasks[0].Root().Visit(true, func(at *tiktak.Task) error {
		var accon *tiktak.Task
		for _, tt := range tasks {
			if !at.Is(tt) {
				continue
			}
			if accon == nil || tt.Is(accon) {
				accon = tt
			}
		}
		if accon != nil {
			tmap[at] = accon
		}
		return nil
	})
	return
}

func (sht *Sheet) Write(w io.Writer, tl tiktak.TimeLine, now time.Time) {
	if len(tl) == 0 {
		return
	}
	loc := time.Local
	fmts := sht.Fmts
	if fmts == nil {
		fmts = MinutesFmts
	}

	end := tiktak.StartDay(tl[len(tl)-1].When(), 1, loc)
	day := tiktak.StartDay(tl[0].When(), 0, loc)

	var tbl tiktbl.Data
	crsr := tbl.At(0, 0).
		SetString(fmt.Sprintf("SHEET: %s – %s",
			fmts.Date(day),
			fmts.Date(tiktak.StartDay(end, -1, loc)),
		), tiktbl.SpanAll, Bold()).NextRow().
		SetString("", tiktbl.SpanAll, tiktbl.Pad('-')).NextRow().
		With(tiktbl.Left, Bold()).SetStrings("Day", "Start", "Stop", "Break", "Work")
	for _, t := range sht.Tasks {
		crsr.SetString(t.String(), tiktbl.Left, Bold())
	}
	if len(sht.Tasks) > 0 {
		crsr.SetString("Rest", tiktbl.Left, Bold())
	}
	if day.Weekday() == sht.WeekStart {
		crsr.NextRow()
	} else {
		crsr.NextRow().SetString("", tiktbl.SpanAll, tiktbl.Pad('-')).NextRow()
	}
	tsums, tmap := accounts(sht.Tasks)
	count, stopCount := 0, 0
	var workSum, breakSum, restSum time.Duration
	var starts, stops time.Duration
	for day.Before(end) {
		style := tiktbl.NoStyle()
		next := tiktak.StartDay(day, 1, loc)
		d, ds, de := tl.Duration(day, next, now, func(s *tiktak.Switch) bool { return s.Task() != nil })
		stop := "..."
		if !de.IsZero() {
			stop = fmts.Clock(de)
			stops += tiktak.ClockOf(de).D
			stopCount++
		} else {
			style = Bold()
		}
		p, _, _ := tl.Duration(ds, de, now, func(s *tiktak.Switch) bool { return s.Task() == nil })

		if day.Weekday() == sht.WeekStart {
			_, w := day.ISOWeek()
			crsr.SetString(
				fmt.Sprintf("-- Week %d ", w),
				tiktbl.SpanAll,
				tiktbl.Pad('-'),
			).NextRow()
		}

		crsr.With(style).SetStrings(
			fmts.ShortDate(day),
			fmts.Clock(ds),
			stop,
			fmts.Duration(p),
			fmts.Duration(d),
		)

		rest := d
		for i, t := range sht.Tasks {
			td, _, _ := tl.Duration(day, next, now, func(s *tiktak.Switch) bool {
				st := s.Task()
				if st == nil {
					return t == nil
				}
				return tmap[st] == t
			})
			if td == 0 {
				crsr.SetString("-", tiktbl.Center)
			} else {
				crsr.SetString(fmts.Duration(td))
				tsums[i].n++
				tsums[i].d += td
				rest -= td
			}
		}
		if len(tsums) > 0 {
			crsr.SetString(fmts.Duration(rest))
		}

		count++
		starts += tiktak.ClockOf(ds).D
		workSum += d
		breakSum += p
		restSum += rest

		crsr.NextRow()
		day = next
	}
	var startAvg, stopAvg tiktak.Clock
	if count > 0 {
		startAvg = tiktak.Clock{D: starts / time.Duration(count), Location: now.Location()}
	}
	if stopCount > 0 {
		stopAvg = tiktak.Clock{D: stops / time.Duration(stopCount), Location: now.Location()}
	}
	crsr.SetString("", tiktbl.SpanAll, tiktbl.Pad('-')).NextRow().
		SetString("Average:", tiktbl.Right, Bold()).
		SetStrings(fmts.Clock(startAvg.On(now)), fmts.Clock(stopAvg.On(now))).
		SetStrings(fmts.Duration(breakSum/time.Duration(count)), fmts.Duration(workSum/time.Duration(count)))
	for _, ts := range tsums {
		if ts.n > 0 {
			crsr.SetString(fmts.Duration(ts.d / time.Duration(ts.n)))
		} else {
			crsr.SetString("-", tiktbl.Center)
		}
	}
	if len(tsums) > 0 {
		if count > 0 {
			crsr.SetString(fmts.Duration(restSum / time.Duration(count)))
		} else {
			crsr.SetString("-", tiktbl.Center)
		}
	}
	crsr.NextRow().
		SetString("Count:", tiktbl.Right, Bold()).Set(count).
		SetString("Sum:", Bold()).
		SetStrings(fmts.Duration(breakSum), fmts.Duration(workSum))
	for _, ts := range tsums {
		crsr.SetString(fmts.Duration(ts.d))
	}
	if len(tsums) > 0 {
		if count > 0 {
			crsr.SetString(fmts.Duration(restSum))
		} else {
			crsr.SetString("-", tiktbl.Center)
		}
	}

	for i := 1; i < tbl.Columns(); i++ {
		tbl.Align(tiktbl.Right, i)
	}
	sht.Layout.Write(w, &tbl)
}
