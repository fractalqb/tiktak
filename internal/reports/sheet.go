package reports

import (
	"fmt"
	"io"
	"time"

	"git.fractalqb.de/fractalqb/tetrta"
	"git.fractalqb.de/fractalqb/tiktak"
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

	var tbl tetrta.Table
	crsr := tbl.At(0, 0).
		SetString(fmt.Sprintf("SHEET: %s – %s",
			fmts.Date(day),
			fmts.Date(tiktak.StartDay(end, -1, loc)),
		), tetrta.SpanAll, Bold()).NextRow().
		SetString("", tetrta.SpanAll, tetrta.CellPad('-')).NextRow().
		With(tetrta.Left, Bold()).SetStrings("Day", "Start", "Stop", "Break", "Work")
	for _, t := range sht.Tasks {
		crsr.SetString(t.String(), tetrta.Left, Bold())
	}
	if len(sht.Tasks) > 0 {
		crsr.SetString("Rest", tetrta.Left, Bold())
	}
	weekSep := func(t time.Time) {
		_, w := t.ISOWeek()
		crsr.SetString(
			fmt.Sprintf(" Week %d ", w),
			tetrta.SpanAll,
			tetrta.Center,
			tetrta.CellPad('-'),
			Muted(),
		).NextRow()
	}
	if day.Weekday() != sht.WeekStart {
		crsr.NextRow()
		weekSep(day)
	} else {
		crsr.NextRow()
	}
	tsums, tmap := accounts(sht.Tasks)
	tsumw := make([]time.Duration, len(tsums))
	count, stopCount, weekCount := 0, 0, 0
	var workSum, breakSum, restSum time.Duration
	var weekWork, weekBreak, weekRest time.Duration
	var starts, stops time.Duration
	var notes []int
	weekSums := func() {
		if weekWork == 0 {
			return
		}
		crsr.SetString("Week count:", Muted()).Set(weekCount, Muted()).
			SetString("Sum:", Muted())
		crsr.With(Muted()).SetStrings(
			fmts.Duration(weekBreak),
			fmts.Duration(weekWork),
		)
		for i, td := range tsumw {
			if td > 0 {
				crsr.SetString(fmts.Duration(td), Muted())
				tsumw[i] = 0
			} else {
				crsr.SetString("-", tetrta.Center, Muted())
			}
		}
		if weekRest > 0 {
			crsr.SetString(fmts.Duration(weekRest), Muted())
		} else if len(tsumw) > 0 {
			crsr.SetString("-", tetrta.Center, Muted())
		}
		crsr.NextRow()
		weekWork, weekBreak, weekRest, weekCount = 0, 0, 0, 0
	}
	for day.Before(end) {
		if day.Weekday() == sht.WeekStart {
			weekSums()
			weekSep(day)
		}

		style := tetrta.NoStyle()
		next := tiktak.StartDay(day, 1, loc)
		dayWork, ds, de := tl.Duration(day, next, now, tiktak.AnyTask)
		if dayWork == 0 {
			day = next
			continue
		}
		stop := "..."
		if !de.IsZero() {
			stop = fmts.Clock(de)
			stops += tiktak.ClockOf(de).Dur
			stopCount++
		} else {
			style = Bold()
		}
		var dayBreak time.Duration
		if de.IsZero() {
			dayBreak, _, _ = tl.Duration(ds, now, now, tiktak.IsATask(nil))
		} else {
			dayBreak, _, _ = tl.Duration(ds, de, now, tiktak.IsATask(nil))
		}

		if hasWarning(tl, day, next, tiktak.AnyTask) {
			crsr.SetString(fmts.ShortDate(day), tetrta.AddStyles(style, Warn()))
		} else {
			crsr.SetString(fmts.ShortDate(day), style)
		}

		crsr.With(style).SetStrings(fmts.Clock(ds), stop)
		if dayBreak > 0 {
			crsr.SetString(fmts.Duration(dayBreak), style)
		} else {
			crsr.SetString("-", style, tetrta.Center)
		}
		crsr.SetString(fmts.Duration(dayWork), style)

		rest := dayWork
		for i, t := range sht.Tasks {
			warns := false
			td, _, _ := tl.Duration(day, next, now, func(s *tiktak.Switch) bool {
				st := s.Task()
				if st == nil {
					if t == nil {
						notes := s.SelectNotes(notes[:0], tiktak.Warning)
						warns = warns || len(notes) > 0
						return true
					}
				} else if tmap[st] == t {
					notes := s.SelectNotes(notes[:0], tiktak.Warning)
					warns = warns || len(notes) > 0
					return true
				}
				return false
			})
			if td == 0 {
				crsr.SetString("-", style, tetrta.Center)
			} else {
				if warns {
					crsr.SetString(fmts.Duration(td), style, Warn())
				} else {
					crsr.SetString(fmts.Duration(td), style)
				}
				tsumw[i] += td
				tsums[i].n++
				tsums[i].d += td
				rest -= td
			}
		}
		if len(tsums) > 0 {
			crsr.SetString(fmts.Duration(rest), style)
			weekRest += rest
		}

		count++
		starts += tiktak.ClockOf(ds).Dur
		weekWork += dayWork
		weekBreak += dayBreak
		workSum += dayWork
		breakSum += dayBreak
		restSum += rest
		if workSum > 0 {
			weekCount++
		}

		crsr.NextRow()
		day = next
	}
	weekSums()
	var startAvg, stopAvg tiktak.Clock
	if count > 0 {
		startAvg = tiktak.Clock{Dur: starts / time.Duration(count), Location: now.Location()}
	}
	if stopCount > 0 {
		stopAvg = tiktak.Clock{Dur: stops / time.Duration(stopCount), Location: now.Location()}
	}
	crsr.SetString("", tetrta.SpanAll, tetrta.CellPad('-')).NextRow().
		SetString("Average:", tetrta.Right, Bold()).
		SetStrings(fmts.Clock(startAvg.On(now)), fmts.Clock(stopAvg.On(now))).
		SetStrings(fmts.Duration(breakSum/time.Duration(count)), fmts.Duration(workSum/time.Duration(count)))
	for _, ts := range tsums {
		if ts.n > 0 {
			crsr.SetString(fmts.Duration(ts.d / time.Duration(ts.n)))
		} else {
			crsr.SetString("-", tetrta.Center)
		}
	}
	if len(tsums) > 0 {
		if count > 0 {
			crsr.SetString(fmts.Duration(restSum / time.Duration(count)))
		} else {
			crsr.SetString("-", tetrta.Center)
		}
	}
	crsr.NextRow().
		SetString("Count:", tetrta.Right, Bold()).Set(count).
		SetString("Sum:", Bold()).
		With(Underline()).SetStrings(fmts.Duration(breakSum), fmts.Duration(workSum))
	for _, ts := range tsums {
		crsr.SetString(fmts.Duration(ts.d), Underline())
	}
	if len(tsums) > 0 {
		if count > 0 {
			crsr.SetString(fmts.Duration(restSum), Underline())
		} else {
			crsr.SetString("-", tetrta.Center)
		}
	}

	for i := 1; i < tbl.Columns(); i++ {
		tbl.Align(tetrta.Right, i)
	}
	sht.Layout.Write(w, &tbl)
}
