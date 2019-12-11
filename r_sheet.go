package main

import (
	"fmt"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	"git.fractalqb.de/fractalqb/tiktak/txtab"

	"golang.org/x/text/language"
)

func init() {
	reports["sheet"] = timeSheetFactory
}

func timeSheetFactory(lang language.Tag, args []string) Reporter {
	r := &timeSheetReport{
		tw: txtab.Writer{
			W: os.Stdout,
			F: newTabFormatter(),
		},
		tasks: args,
	}
	return r
}

type sheetDay struct {
	Span
	hasRunning bool
}

const flagStrMax = 1

func (sd *sheetDay) FlagStr() string {
	if sd.hasRunning {
		return "↻"
	}
	return " "
}

func allDays(t *Task) (days []sheetDay, hasFlags bool) {
	t.WalkAll(nil, func(tp []*Task, nmp []string) {
		task := tp[len(tp)-1]
	NEXT_SPAN:
		for _, span := range task.Spans {
			day := DaySpan(span.Start)
			for i := range days {
				d := &days[i]
				if day.Start == d.Start {
					d.hasRunning = d.hasRunning || span.Stop == nil
					hasFlags = hasFlags || d.hasRunning
					continue NEXT_SPAN
				}
			}
			d := sheetDay{
				Span:       day,
				hasRunning: span.Stop == nil,
			}
			days = append(days, d)
			hasFlags = hasFlags || d.hasRunning
		}
	})
	sort.Slice(days, func(i, j int) bool {
		return days[i].Start.Before(days[j].Start)
	})
	return days, hasFlags
}

type timeSheetReport struct {
	tw    txtab.Writer
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
	days, hasFlags := allDays(root)
	tw := &rep.tw
	if hasFlags {
		tw.AddColumn("⚐", flagStrMax, txtab.Center)
	}
	tw.AddColumn("Day", utf8.RuneCountInString(dateFormat))
	tw.AddColumn("Start", 5, txtab.Right)
	tw.AddColumn("Stop", 5, txtab.Right)
	tw.AddColumn("Break", 6, txtab.Right)
	tw.AddColumn("Work", 6, txtab.Right)
	for _, task := range rep.tasks {
		w := utf8.RuneCountInString(task)
		if w < 6 {
			w = 6
		}
		tw.AddColumn(task, w, txtab.Right)
	}
	var brk, wrk time.Duration
	var starts, stops int
	tsk := make(map[string]time.Duration)
	dayNo := 0
	fmt.Fprintf(tw.W, "TIME-SHEET %s:\n", reportMonth(now))
	tw.Header()
	tw.Hrule()
	week := -1
	for _, day := range days {
		if dayNo == 0 {
			_, week = day.Start.ISOWeek()
		} else if _, w := day.Start.ISOWeek(); w != week {
			tw.Hrule()
			week = w
		}
		tw.RowStart()
		if hasFlags {
			tw.Cell(day.FlagStr())
		}
		tw.Cell(day.Start.Format(dateFormat))
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
				today := IntersectSpans(&day.Span, &span)
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
			tw.Cells(tw.F.ColumnNo()-1, "")
		} else {
			tw.Cell(work.Start.Format(clockFormat))
			tw.Cell(work.Stop.Format(clockFormat))
			wdur, _ := work.Duration(now)
			tw.Cell(hm(wdur - tdur))
			tw.Cell(hm(tdur))
			for _, task := range rep.tasks {
				if td := perTask[task]; td > 0 {
					tw.Cell(hm(perTask[task]))
				} else {
					tw.Cell("")
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
		tw.RowEnd()
	}
	tw.Hrule()
	if dayNo > 0 {
		starts /= dayNo
		stops /= dayNo
		div := time.Duration(dayNo)
		tw.RowStart()
		if hasFlags {
			tw.Cell(nil)
		}
		tw.Cell("Avg:", txtab.Right)
		tw.Cell(fmt.Sprintf("%02d:%02d", starts/60, starts%60))
		tw.Cell(fmt.Sprintf("%02d:%02d", stops/60, stops%60))
		tw.Cell(hm(brk / div))
		tw.Cell(hm(wrk / div))
		for _, task := range rep.tasks {
			tw.Cell(hm(tsk[task] / div))
		}
		tw.RowEnd()
	}
	tw.RowStart()
	if hasFlags {
		tw.Cell(nil)
	}
	tw.Cell("Count:", txtab.Right)
	tw.Cell(dayNo, txtab.Left)
	tw.Cell("Sum:", txtab.Right)
	tw.Cell(hm(brk))
	tw.Cell(hm(wrk))
	for _, task := range rep.tasks {
		tw.Cell(hm(tsk[task]))
	}
	tw.RowEnd()
}
