package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	"golang.org/x/text/language"
)

func init() {
	reports["sheet"] = timeSheetFactory
}

func timeSheetFactory(lang language.Tag, args []string) Reporter {
	r := &timeSheetReport{
		wr:    os.Stdout,
		tw:    borderedWriter{os.Stdout, rPrefix},
		tasks: args,
	}
	return r
}

func allDays(t *Task) (days []Span) {
	t.WalkAll(nil, func(tp []*Task, nmp []string) {
		task := tp[len(tp)-1]
	NEXT_SPAN:
		for _, span := range task.Spans {
			day := DaySpan(span.Start)
			for _, d := range days {
				if day.Start == d.Start {
					continue NEXT_SPAN
				}
			}
			days = append(days, day)
		}
	})
	sort.Slice(days, func(i, j int) bool {
		return days[i].Start.Before(days[j].Start)
	})
	return days
}

type timeSheetReport struct {
	wr    io.Writer
	tw    tableWriter
	tasks []string
}

func (rep *timeSheetReport) accountOn(t *Task) string {
	tPath := t.Path()
	for len(tPath) > 0 {
		tpStr := pathString(tPath)
		for _, t := range rep.tasks {
			if tpStr == t {
				return t
			}
		}
		tPath = tPath[:len(tPath)-1]
	}
	return ""
}

func (rep *timeSheetReport) Generate(root *Task, now time.Time) {
	days := allDays(root)
	tbl := []tableCol{
		tableCol{"Day", utf8.RuneCountInString(dateFormat)},
		tableCol{"Start", 5},
		tableCol{"Stop", 5},
		tableCol{"Break", 6},
		tableCol{"Work", 6},
	}
	for _, task := range rep.tasks {
		w := utf8.RuneCountInString(task)
		if w < 6 {
			w = 6
		}
		tbl = append(tbl, tableCol{task, w})
	}
	var brk, wrk time.Duration
	var starts, stops int
	tsk := make(map[string]time.Duration)
	dayNo := 0
	fmt.Fprintf(rep.wr, "TIME-SHEET %s:\n", reportMonth(now))
	rep.tw.Head(tbl...)
	rep.tw.HRule(tbl...)
	for _, day := range days {
		rep.tw.StartRow()
		rep.tw.Cell(tbl[0].Width(), day.Start.Format(dateFormat))
		var work *Span
		var tdur time.Duration
		perTask := make(map[string]time.Duration)
		root.WalkAll(nil, func(tp []*Task, nmp []string) {
			task := tp[len(tp)-1]
			accOn := rep.accountOn(task)
			for _, span := range task.Spans {
				if span.Stop == nil {
					span.Stop = &now
				}
				today := IntersectSpans(&day, &span)
				if d, _ := today.Duration(*today.Stop); d == 0 {
					continue
				} else {
					tdur += d
					perTask[accOn] += d
					tsk[accOn] += d
				}
				if work == nil {
					work = &today
				} else {
					*work = CoverSpans(work, &today)
				}
			}
		})
		if work == nil {
			rep.tw.Cell(colsWidth(rep.tw, tbl[1:]...), "")
		} else {
			rep.tw.Cell(tbl[1].Width(), work.Start.Format(clockFormat))
			rep.tw.Cell(tbl[2].Width(), work.Stop.Format(clockFormat))
			wdur, _ := work.Duration(now)
			rep.tw.Cell(-tbl[3].Width(), hm(wdur-tdur).String())
			rep.tw.Cell(-tbl[4].Width(), hm(tdur).String())
			for i, task := range rep.tasks {
				if td := perTask[task]; td > 0 {
					rep.tw.Cell(-tbl[5+i].Width(), hm(perTask[task]).String())
				} else {
					rep.tw.Cell(-tbl[5+i].Width(), "")
				}
			}
			brk += wdur - tdur
			wrk += tdur
			h, m, _ := work.Start.Clock()
			starts += 60*h + m
			h, m, _ = work.Stop.Clock()
			stops += 60*h + m
			dayNo++
		}
		fmt.Fprintln(rep.wr)
	}
	rep.tw.HRule(tbl...)
	rep.tw.StartRow()
	rep.tw.Cell(-tbl[0].Width(), "Sum:")
	rep.tw.Cell(colsWidth(rep.tw, tbl[1:3]...), "")
	rep.tw.Cell(-tbl[3].Width(), hm(brk).String())
	rep.tw.Cell(-tbl[4].Width(), hm(wrk).String())
	for i, task := range rep.tasks {
		rep.tw.Cell(-tbl[5+i].Width(), hm(tsk[task]).String())
	}
	fmt.Fprintln(rep.wr)
	if dayNo > 0 {
		starts /= dayNo
		stops /= dayNo
		div := time.Duration(dayNo)
		rep.tw.StartRow()
		rep.tw.Cell(-tbl[0].Width(), "Avg:")
		rep.tw.Cell(tbl[1].Width(), fmt.Sprintf("%02d:%02d", starts/60, starts%60))
		rep.tw.Cell(tbl[1].Width(), fmt.Sprintf("%02d:%02d", stops/60, stops%60))
		//tw.Cell(tableColsWidth(tbl[1:3]...), "")
		rep.tw.Cell(-tbl[3].Width(), hm(brk/div).String())
		rep.tw.Cell(-tbl[4].Width(), hm(wrk/div).String())
		for i, task := range rep.tasks {
			rep.tw.Cell(-tbl[5+i].Width(), hm(tsk[task]/div).String())
		}
		fmt.Fprintln(rep.wr)
	}
}
