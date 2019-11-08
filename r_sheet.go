package main

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"golang.org/x/text/language"
)

func init() {
	reports["sheet"] = timeSheetFactory
}

func timeSheetFactory(now time.Time, lang language.Tag, args []string) func(*Task) {
	r := timeSheetReport{
		now:   now,
		tasks: args,
	}
	return r.report
}

type timeSheetReport struct {
	now   time.Time
	tasks []string
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
	return days
}

func (rep *timeSheetReport) report(root *Task) {
	var wr = os.Stdout
	days := allDays(root)
	tbl := []tableCol{
		tableCol{"Day", utf8.RuneCountInString(dateFormat)},
		tableCol{"Start", 5},
		tableCol{"Stop", 5},
		tableCol{"Break", 5},
		tableCol{"Work", 5},
	}
	for _, task := range rep.tasks {
		w := utf8.RuneCountInString(task)
		if w < 6 {
			w = 6
		}
		tbl = append(tbl, tableCol{task, w})
	}
	var brk, wrk time.Duration
	dayNo := 0
	tableHead(wr, rPrefix, tbl...)
	tableHRule(wr, rPrefix, tbl...)
	for _, day := range days {
		tableStartRow(wr, rPrefix)
		tableCell(wr, tbl[0].Width(), day.Start.Format(dateFormat))
		var work *Span
		var tdur time.Duration
		root.WalkAll(nil, func(tp []*Task, nmp []string) {
			task := tp[len(tp)-1]
			for _, span := range task.Spans {
				if span.Stop == nil {
					span.Stop = &rep.now
				}
				today := IntersectSpans(&day, &span)
				if d, _ := today.Duration(*today.Stop); d == 0 {
					continue
				} else {
					tdur += d
				}
				if work == nil {
					work = &today
				} else {
					*work = CoverSpans(work, &today)
				}
			}
		})
		if work == nil {
			tableCell(wr, tableColsWidth(tbl[1:]...), "")
		} else {
			tableCell(wr, tbl[1].Width(), work.Start.Format(clockFormat))
			tableCell(wr, tbl[2].Width(), work.Stop.Format(clockFormat))
			wdur, _ := work.Duration(rep.now)
			tableCell(wr, tbl[3].Width(), hm(wdur-tdur).String())
			tableCell(wr, tbl[4].Width(), hm(tdur).String())
			brk += wdur - tdur
			wrk += tdur
			dayNo++
		}
		fmt.Fprintln(wr)
	}
	tableHRule(wr, rPrefix, tbl...)
	tableStartRow(wr, rPrefix)
	tableCell(wr, -tbl[0].Width(), "Sum:")
	tableCell(wr, tableColsWidth(tbl[1:3]...), "")
	tableCell(wr, tbl[3].Width(), hm(brk).String())
	tableCell(wr, tbl[4].Width(), hm(wrk).String())
	fmt.Fprintln(wr)
	tableStartRow(wr, rPrefix)
	tableCell(wr, -tbl[0].Width(), "Avg:")
	tableCell(wr, tableColsWidth(tbl[1:3]...), "")
	tableCell(wr, tbl[3].Width(), hm(brk/time.Duration(dayNo)).String())
	tableCell(wr, tbl[4].Width(), hm(wrk/time.Duration(dayNo)).String())
	fmt.Fprintln(wr)
}
