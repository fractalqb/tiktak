package main

import (
	"fmt"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	"golang.org/x/text/language"

	"golang.org/x/text/collate"
)

func init() {
	reports["spans"] = byStartTimeFactory
}

func byStartTimeFactory(now time.Time, lang language.Tag, args []string) func(*Task) {
	r := byStartTimeReport{
		now:  now,
		coll: collate.New(lang),
	}
	return r.report
}

func appendUniqeTask(ts []*Task, t *Task) []*Task {
	for _, e := range ts {
		if e == t {
			return ts
		}
	}
	return append(ts, t)
}

type byStartTimeReport struct {
	now  time.Time
	coll *collate.Collator
}

func (rep *byStartTimeReport) report(root *Task) {
	var starts []time.Time
	t2ts := make(map[time.Time][]*Task)
	var tpWidth int
	root.WalkAll(rep.coll, func(tp []*Task, nmp []string) {
		task := tp[len(tp)-1]
		if tpl := utf8.RuneCountInString(pathString(nmp)); tpl > tpWidth {
			tpWidth = tpl
		}
		for _, span := range task.Spans {
			if tTasks, ok := t2ts[span.Start]; ok {
				tTasks = appendUniqeTask(tTasks, task)
				t2ts[span.Start] = tTasks
			} else {
				t2ts[span.Start] = []*Task{task}
				starts = append(starts, span.Start)
			}
		}
	})
	sort.Slice(starts, func(i, j int) bool {
		return starts[i].Before(starts[j])
	})
	tbl := []tableCol{
		tableCol{"↹", 1},
		tableCol{"Start", 5},
		tableCol{"Stop", 5},
		tableCol{"Dur", 5},
		tableCol{"Task", tpWidth},
	}
	wr := os.Stdout
	fmt.Fprintf(wr, "TIMESPANS PER DAY %s:\n", reportMonth(rep.now))
	tableHead(wr, rPrefix, tbl...)
	day := 0
	var lastSpan *Span
	for _, start := range starts {
		thisDay := 100*start.Year() + start.YearDay()
		if thisDay != day {
			tableHRule(wr, rPrefix, tbl...)
			tableStartRow(wr, rPrefix)
			tableCell(wr, tableColsWidth(tbl...), start.Format(dateFormat))
			fmt.Fprintln(wr)
			tableHRule(wr, rPrefix, tbl...)
			day = thisDay
		}
		tasks := t2ts[start]
		for _, task := range tasks {
			for _, span := range task.Spans {
				if span.Start == start {
					tableStartRow(wr, rPrefix)
					if lastSpan == nil {
						tableCell(wr, tbl[0].Width(), "")
						lastSpan = new(Span)
						*lastSpan = span
					} else {
						is := IntersectSpans(lastSpan, &span)
						d, fin := is.Duration(rep.now)
						if !fin || d > 0 {
							tableCell(wr, tbl[0].Width(), "⇸" /*↪"*/)
						} else if span.Start.After(*lastSpan.Stop) {
							tableCell(wr, tbl[0].Width(), "⇢")
						} else {
							tableCell(wr, tbl[0].Width(), "")
						}
						*lastSpan = span
					}
					tableCell(wr, tbl[1].Width(), start.Format(clockFormat))
					if span.Stop == nil {
						tableCell(wr, tbl[2].Width(), "")
					} else {
						tableCell(wr, tbl[2].Width(), span.Stop.Format(clockFormat))
					}
					dur, _ := span.Duration(rep.now)
					tableCell(wr, tbl[3].Width(), hm(dur).String())
					tableCell(wr, tbl[4].Width(), pathString(task.Path()))
					fmt.Fprintln(wr)
				}
			}
		}
	}
}
