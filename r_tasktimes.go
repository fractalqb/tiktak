package main

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func init() {
	reports["tasks"] = taskTimesFactory
}

func taskTimesFactory(lang language.Tag, args []string) Reporter {
	res := &taskTimesReport{
		wr:   os.Stdout,
		tw:   borderedWriter{os.Stdout, rPrefix},
		lang: lang,
	}
	return res
}

type taskTimesReport struct {
	wr   io.Writer
	tw   tableWriter
	lang language.Tag
}

func (rep taskTimesReport) Generate(root *Task, now time.Time) {
	coll := collate.New(rep.lang)
	var tpWidth, titleWidth int
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
	fmt.Fprintf(rep.wr, "TASK TIMES %s:\n", reportMonth(now))
	tbl := []tableCol{tableCol{title: "⏲"}, tableCol{"Task", tpWidth}}
	if titleWidth > 0 {
		tbl = append(tbl, tableCol{"Title", titleWidth + 2})
	}
	tbl = append(tbl, tableCol{"Today", 5}, tableCol{"All", 6})
	rep.tw.Head(tbl...)
	rep.tw.HRule(tbl...)
	var sumAll, sumDay time.Duration
	root.WalkAll(coll, func(tp []*Task, nmp []string) {
		ct := tp[len(tp)-1]
		if len(ct.Spans) == 0 {
			return
		}
		rep.tw.StartRow()
		if running[ct] {
			rep.tw.Cell(tbl[0].Width(), "↻")
		} else {
			rep.tw.Cell(tbl[0].Width(), " ")
		}
		rep.tw.Cell(tbl[1].Width(), pathString(nmp))
		colOff := 0
		if titleWidth > 0 {
			if ct.Title == "" {
				rep.tw.Cell(tbl[2].Width(), "")
			} else {
				rep.tw.Cell(tbl[2].Width(), `"`+ct.Title+`"`)
			}
			colOff = 1
		}
		var (
			durAll, durDay time.Duration
			today          = DaySpan(now)
		)
		for _, span := range ct.Spans {
			d, _ := span.Duration(now)
			durAll += d
			if span.Stop == nil {
				span.Stop = &now
			}
			daySpan := IntersectSpans(&today, &span)
			d, _ = daySpan.Duration(now)
			durDay += d
		}
		if durDay > 0 {
			rep.tw.Cell(-tbl[2+colOff].Width(), hm(durDay).String())
		} else {
			rep.tw.Cell(tbl[2+colOff].Width(), "")
		}
		rep.tw.Cell(-tbl[3+colOff].Width(), hm(durAll).String())
		fmt.Fprintln(rep.wr)
		sumAll += durAll
		sumDay += durDay
	})
	rep.tw.HRule(tbl...)
	rep.tw.StartRow()
	if titleWidth > 0 {
		rep.tw.Cell(-colsWidth(rep.tw, tbl[:3]...), "Sum:")
		rep.tw.Cell(-tbl[3].Width(), hm(sumDay).String())
		rep.tw.Cell(-tbl[4].Width(), hm(sumAll).String())
	} else {
		rep.tw.Cell(-colsWidth(rep.tw, tbl[:2]...), "Sum:")
		rep.tw.Cell(-tbl[2].Width(), hm(sumDay).String())
		rep.tw.Cell(-tbl[3].Width(), hm(sumAll).String())
	}
	fmt.Fprintln(rep.wr)
	if runNo > 1 {
		fmt.Printf("\n%s%d RUNNING TASKS\n", rPrefix, runNo)
	}
}
