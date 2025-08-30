package reports

import (
	"fmt"
	"io"
	"time"

	"git.fractalqb.de/fractalqb/tetrta"
	"git.fractalqb.de/fractalqb/tiktak"
)

type Sums struct {
	Report
	WeekStart time.Weekday
}

func (sm *Sums) Write(w io.Writer, tl tiktak.TimeLine, now time.Time) {
	troot := tl.FirstTask().Root()
	if troot == nil || len(tl) == 0 {
		return
	}
	_, week := now.ISOWeek()
	tsums := NewTaskSums(now, sm.WeekStart)
	ts, te := tl[0].When(), tl[len(tl)-1].When()
	total := ts.Before(tsums.ms)
	if !total {
		sw := tl[len(tl)-1]
		if sw.Task() == nil {
			total = te.After(tsums.me)
		} else {
			total = !te.Before(tsums.me)
		}
	}

	var tbl tetrta.Table
	var caption string
	if total {
		caption = fmt.Sprintf("SUMS: %s; Week %d [%s – %s]:",
			sm.Fmts.Date(now),
			week,
			sm.Fmts.Date(ts),
			sm.Fmts.Date(te),
		)
	} else {
		caption = fmt.Sprintf("SUMS: %s; Week %d:", sm.Fmts.Date(now), week)
	}
	crsr := tbl.At(0, 0).
		SetString(caption, tetrta.SpanAll, Bold()).NextRow().
		SetString("", tetrta.SpanAll, tetrta.CellPad('-')).NextRow().
		With(tetrta.Left).SetStrings("", "Task", "Today.", "Today/", "Week.", "Week/", "Month.", "Month/")
	if total {
		crsr.With(tetrta.Left).SetStrings("Total.", "Total/")
	}
	crsr.NextRow().
		SetString("", tetrta.SpanAll, tetrta.CellPad('-')).NextRow()

	troot.Visit(false, func(t *tiktak.Task) error {
		var markers string
		tsums.Of(tl, t, sm.Fmts)
		style1 := tetrta.NoStyle()
		if tsums.Open {
			style1 = Bold()
			markers = ">"
		}
		styleSub := style1
		if t.Root() == t {
			styleSub = tetrta.AddStyles(styleSub, Underline())
		}
		crsr.With(style1).SetStrings(markers, t.String())

		warn1 := hasWarning(tl, tsums.ds, tsums.de, tiktak.SameTask(t))
		if tsums.Day1 == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warn1 {
			crsr.SetString(tsums.Day1, tetrta.AddStyles(style1, Warn()))
		} else {
			crsr.SetString(tsums.Day1, style1)
		}
		warnSub := hasWarning(tl, tsums.ds, tsums.de, tiktak.IsATask(t))
		if tsums.DaySub == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warnSub {
			crsr.SetString(tsums.DaySub, tetrta.AddStyles(styleSub, Warn()))
		} else {
			crsr.SetString(tsums.DaySub, styleSub)
		}
		warn1 = warn1 || hasWarning(tl, tsums.ws, tsums.we, tiktak.SameTask(t))
		if tsums.Week1 == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warn1 {
			crsr.SetString(tsums.Week1, tetrta.AddStyles(style1, Warn()))
		} else {
			crsr.SetString(tsums.Week1, style1)
		}
		warnSub = warnSub || hasWarning(tl, tsums.ws, tsums.we, tiktak.IsATask(t))
		if tsums.WeekSub == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warnSub {
			crsr.SetString(tsums.WeekSub, tetrta.AddStyles(styleSub, Warn()))
		} else {
			crsr.SetString(tsums.WeekSub, styleSub)
		}
		warn1 = warn1 || hasWarning(tl, tsums.ms, tsums.me, tiktak.SameTask(t))
		if tsums.Month1 == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warn1 {
			crsr.SetString(tsums.Month1, tetrta.AddStyles(style1, Warn()))
		} else {
			crsr.SetString(tsums.Month1, style1)
		}
		warnSub = warnSub || hasWarning(tl, tsums.ms, tsums.me, tiktak.IsATask(t))
		if tsums.MonthSub == empty {
			crsr.SetString(empty, tetrta.Center)
		} else if warnSub {
			crsr.SetString(tsums.MonthSub, tetrta.AddStyles(styleSub, Warn()))
		} else {
			crsr.SetString(tsums.MonthSub, styleSub)
		}
		if total {
			warn1 = warn1 || hasWarning(tl, ts, te, tiktak.SameTask(t))
			d, _, _ := tl.Duration(ts, te, now, tiktak.SameTask(t))
			if d == 0 {
				crsr.SetString(empty, tetrta.Center)
			} else if warn1 {
				crsr.SetString(sm.Fmts.Duration(d), tetrta.AddStyles(style1, Warn()))
			} else {
				crsr.SetString(sm.Fmts.Duration(d), style1)
			}
			warnSub = warnSub || hasWarning(tl, ts, te, tiktak.IsATask(t))
			d, _, _ = tl.Duration(ts, te, now, tiktak.IsATask(t))
			if d == 0 {
				crsr.SetString(empty, tetrta.Center)
			} else if warnSub {
				crsr.SetString(sm.Fmts.Duration(d), tetrta.AddStyles(styleSub, Warn()))
			} else {
				crsr.SetString(sm.Fmts.Duration(d), styleSub)
			}
		}

		crsr.NextRow()
		return nil
	})
	for i := 2; i < tbl.Columns(); i++ {
		tbl.Align(tetrta.Right, i)
	}
	sm.Layout.Write(w, &tbl)
}

type TaskSums struct {
	Day1, DaySub     string
	Week1, WeekSub   string
	Month1, MonthSub string
	Open             bool

	now    time.Time
	ds, de time.Time
	ws, we time.Time
	ms, me time.Time
}

func NewTaskSums(now time.Time, sow time.Weekday) *TaskSums {
	res := &TaskSums{
		now: now,
		ds:  tiktak.StartDay(now, 0, time.Local),
		de:  tiktak.StartDay(now, 1, time.Local),
		ms:  tiktak.StartMonth(now, 0, time.Local),
		me:  tiktak.StartMonth(now, 1, time.Local),
	}
	res.ws = tiktak.LastDay(sow, res.ds, time.Local)
	res.we = tiktak.NextDay(sow, res.ds, time.Local)
	return res
}

func (ts *TaskSums) Of(tl tiktak.TimeLine, t *tiktak.Task, fmts Formats) {
	forInterval := func(s, e time.Time, open bool) (i1, is string, io bool) {
		i1, is = empty, empty
		d, ds, de := tl.Duration(s, e, ts.now, tiktak.SameTask(t))
		open = open || (!ds.IsZero() && de.IsZero())
		if d > 0 {
			i1 = fmts.Duration(d)
		}
		if t != nil && len(t.Subtasks()) > 0 {
			d, _, _ = tl.Duration(s, e, ts.now, tiktak.IsATask(t))
			if d > 0 {
				is = fmts.Duration(d)
			}
		}
		return i1, is, open
	}
	ts.Day1, ts.DaySub, ts.Open = forInterval(ts.ds, ts.de, false)
	ts.Week1, ts.WeekSub, ts.Open = forInterval(ts.ws, ts.we, ts.Open)
	ts.Month1, ts.MonthSub, ts.Open = forInterval(ts.ms, ts.me, ts.Open)
}
