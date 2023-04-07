package reports

import (
	"fmt"
	"io"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
)

type Sums struct {
	Report
	WeekStart time.Weekday
}

func (sm *Sums) Write(w io.Writer, tl tiktak.TimeLine, now time.Time) {
	troot := tl.FirstTask().Root()
	if troot == nil {
		return
	}
	_, week := now.ISOWeek()
	tsums := NewTaskSums(now, sm.WeekStart)

	var tbl tiktbl.Data
	crsr := tbl.At(0, 0).
		SetString(fmt.Sprintf("SUMS: %s; Week %d:", sm.Fmts.Date(now), week), tiktbl.SpanAll, Bold()).NextRow().
		SetString("", tiktbl.SpanAll, tiktbl.Pad('-')).NextRow().
		With(tiktbl.Left).SetStrings("", "Task", "Today.", "Today/", "Week.", "Week/", "Month.", "Month/").
		NextRow().
		SetString("", tiktbl.SpanAll, tiktbl.Pad('-')).NextRow()

	troot.Visit(false, func(t *tiktak.Task) error {
		var markers string
		tsums.Of(tl, t, sm.Fmts)
		style := tiktbl.NoStyle()
		if tsums.Open {
			style = Bold()
			markers = ">"
		}

		crsr.With(style).SetStrings(
			markers,
			t.String(),
			tsums.Day1, tsums.DaySub,
			tsums.Week1, tsums.WeekSub,
			tsums.Month1, tsums.MonthSub,
		).NextRow()
		return nil
	})
	for i := 2; i < tbl.Columns(); i++ {
		tbl.Align(tiktbl.Right, i)
	}
	sm.Layout.Write(w, &tbl)
}

type TaskSums struct {
	Day1, DaySub     string
	Week1, WeekSub   string
	Month1, MonthSub string
	Open             bool

	now, ds, de, ws, we, ms, me time.Time
}

func NewTaskSums(now time.Time, sow time.Weekday) *TaskSums {
	return &TaskSums{
		now: now,
		ds:  tiktak.StartDay(now, 0, time.Local),
		de:  tiktak.StartDay(now, 1, time.Local),
		ws:  tiktak.LastDay(sow, now, time.Local),
		we:  tiktak.NextDay(sow, now, time.Local),
		ms:  tiktak.StartMonth(now, 0, time.Local),
		me:  tiktak.StartMonth(now, 1, time.Local),
	}
}

func (ts *TaskSums) Of(tl tiktak.TimeLine, t *tiktak.Task, fmts Formats) {
	forInterval := func(s, e time.Time, open bool) (i1, is string, io bool) {
		i1, is = "-", "-"
		d, ds, de := tl.Duration(s, e, ts.now, func(s *tiktak.Switch) bool {
			return s.Task() == t
		})
		open = open || (!ds.IsZero() && de.IsZero())
		if d > 0 {
			i1 = fmts.Duration(d)
		}
		if t != nil && len(t.Subtasks()) > 0 {
			d, _, _ = tl.Duration(s, e, ts.now, func(s *tiktak.Switch) bool {
				return s.Task().Is(t)
			})
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
