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

func defaultOutput(wr io.Writer, root *Task, now time.Time, lang language.Tag) {
	coll := collate.New(lang)
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
	fmt.Println("TASK TIMES:")
	tbl := []tableCol{tableCol{title: "⏲"}, tableCol{"Task", tpWidth}}
	if titleWidth > 0 {
		tbl = append(tbl, tableCol{"Title", titleWidth + 2})
	}
	tbl = append(tbl, tableCol{"All", 6}, tableCol{"Today", 5})
	tableHead(wr, rPrefix, tbl...)
	tableHRule(wr, rPrefix, tbl...)
	var sumAll, sumDay time.Duration
	root.WalkAll(coll, func(tp []*Task, nmp []string) {
		ct := tp[len(tp)-1]
		if len(ct.Spans) == 0 {
			return
		}
		tableStartRow(wr, rPrefix)
		if running[ct] {
			tableCell(wr, tbl[0].Width(), "↻")
		} else {
			tableCell(wr, tbl[0].Width(), " ")
		}
		tableCell(wr, tbl[1].Width(), pathString(nmp))
		colOff := 0
		if titleWidth > 0 {
			if ct.Title == "" {
				tableCell(wr, tbl[2].Width(), "")
			} else {
				tableCell(wr, tbl[2].Width(), `"`+ct.Title+`"`)
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
		tableCell(wr, -tbl[2+colOff].Width(), hm(durAll).String())
		if durDay > 0 {
			tableCell(wr, -tbl[3+colOff].Width(), hm(durDay).String())
		} else {
			tableCell(wr, tbl[3+colOff].Width(), "")
		}
		fmt.Fprintln(wr)
		sumAll += durAll
		sumDay += durDay
	})
	tableHRule(os.Stdout, rPrefix, tbl...)
	tableStartRow(wr, rPrefix)
	if titleWidth > 0 {
		tableCell(wr, -tableColsWidth(tbl[:3]...), "Sum:")
		tableCell(wr, -tbl[3].Width(), hm(sumAll).String())
		tableCell(wr, -tbl[4].Width(), hm(sumDay).String())
	} else {
		tableCell(wr, -tableColsWidth(tbl[:2]...), "Sum:")
		tableCell(wr, -tbl[2].Width(), hm(sumAll).String())
		tableCell(wr, -tbl[3].Width(), hm(sumDay).String())
	}
	fmt.Fprintln(wr)
	if runNo > 1 {
		fmt.Printf("\n%s%d RUNNING TASKS\n", rPrefix, runNo)
	}
}
