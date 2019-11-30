package main

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"git.fractalqb.de/fractalqb/tiktak/txtab"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func init() {
	reports["tasks"] = taskTimesFactory
}

func taskTimesFactory(lang language.Tag, args []string) Reporter {
	res := &taskTimesReport{
		tw: txtab.Writer{
			W: os.Stdout,
			F: newTabFormatter(),
		},
		lang: lang,
	}
	return res
}

type taskTimesReport struct {
	tw   txtab.Writer
	lang language.Tag
}

func (rep taskTimesReport) Generate(root *Task, now time.Time) {
	coll := collate.New(rep.lang)
	var tpWidth, titleWidth int
	today := DaySpan(now)
	thisWeek := WeekSpan(now)
	runNo := 0
	running := make(map[*Task]bool)
	root.WalkAll(coll, func(tp []*Task, nmp []string) {
		depth := len(tp) - 1
		if w := utf8.RuneCountInString(tp[depth].Title); w > titleWidth {
			titleWidth = w
		}
		pw := depth
		for _, nm := range nmp {
			pw += utf8.RuneCountInString(nm)
		}
		if pw > tpWidth {
			tpWidth = pw
		}
		ct := tp[depth]
		run := false
		for _, span := range ct.Spans {
			run = run || span.Stop == nil
		}
		if run {
			runNo++
			running[ct] = true
		} else {
			running[ct] = false
		}
	})
	tw := &rep.tw
	fmt.Fprintf(tw.W, "TASK TIMES %s:\n", now.Format(dateFormat))
	tw.AddColumn("⏲")
	tw.AddColumn("Task", tpWidth)
	if titleWidth > 0 {
		tw.AddColumn("Title", titleWidth+2)
	}
	tw.AddColumn("Today", 5, txtab.Right)
	tw.AddColumn("Week", 5, txtab.Right)
	tw.AddColumn("Month", 6, txtab.Right)
	tw.Header()
	tw.Hrule()
	var sumAll, sumWeek, sumDay time.Duration
	root.WalkAll(coll, func(tp []*Task, nmp []string) {
		ct := tp[len(tp)-1]
		if len(ct.Spans) == 0 {
			return
		}
		tw.RowStart()
		if running[ct] {
			tw.Cell("↻")
		} else {
			tw.Cell("")
		}
		tw.Cell(pathString(nmp))
		if titleWidth > 0 {
			if ct.Title == "" {
				tw.Cell("")
			} else {
				tw.Cell(`"` + ct.Title + `"`)
			}
		}
		var durAll, durWeek, durDay time.Duration
		for _, span := range ct.Spans {
			d, _ := span.Duration(now)
			durAll += d
			if span.Stop == nil {
				span.Stop = &now
			}
			weekSpan := IntersectSpans(&thisWeek, &span)
			d, _ = weekSpan.Duration(now)
			durWeek += d
			daySpan := IntersectSpans(&today, &span)
			d, _ = daySpan.Duration(now)
			durDay += d
		}
		if durDay > 0 {
			tw.Cell(hm(durDay))
		} else {
			tw.Cell("")
		}
		if durWeek > 0 {
			tw.Cell(hm(durWeek))
		} else {
			tw.Cell("")
		}
		tw.Cell(hm(durAll))
		tw.RowEnd()
		sumAll += durAll
		sumWeek += durWeek
		sumDay += durDay
	})
	tw.Hrule()
	tw.RowStart()
	if titleWidth > 0 {
		tw.Cells(3, "Sum:", txtab.Right)
	} else {
		tw.Cells(2, "Sum:", txtab.Right)
	}
	tw.Cell(hm(sumDay))
	tw.Cell(hm(sumWeek))
	tw.Cell(hm(sumAll))
	tw.RowEnd()
	if runNo > 1 {
		fmt.Printf("\n%s%d RUNNING TASKS\n", rPrefix, runNo)
	}
}
