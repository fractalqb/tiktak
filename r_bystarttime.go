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
	tw := borderedWriter{os.Stdout, rPrefix}
	fmt.Fprintf(tw.wr, "TIMESPANS PER DAY %s:\n", reportMonth(rep.now))
	tw.Head(tbl...)
	day := 0
	var lastSpan *Span
	for _, start := range starts {
		thisDay := 100*start.Year() + start.YearDay()
		if thisDay != day {
			tw.HRule(tbl...)
			tw.StartRow()
			tw.Cell(colsWidth(tw, tbl...), start.Format(dateFormat))
			fmt.Fprintln(tw.wr)
			tw.HRule(tbl...)
			day = thisDay
		}
		tasks := t2ts[start]
		for _, task := range tasks {
			for _, span := range task.Spans {
				if span.Start == start {
					tw.StartRow()
					if lastSpan == nil {
						tw.Cell(tbl[0].Width(), "")
						lastSpan = new(Span)
						*lastSpan = span
					} else {
						is := IntersectSpans(lastSpan, &span)
						d, fin := is.Duration(rep.now)
						if !fin || d > 0 {
							tw.Cell(tbl[0].Width(), "⇸" /*↪"*/)
						} else if span.Start.After(*lastSpan.Stop) {
							tw.Cell(tbl[0].Width(), "⇢")
						} else {
							tw.Cell(tbl[0].Width(), "")
						}
						*lastSpan = span
					}
					tw.Cell(tbl[1].Width(), start.Format(clockFormat))
					if span.Stop == nil {
						tw.Cell(tbl[2].Width(), "")
					} else {
						tw.Cell(tbl[2].Width(), span.Stop.Format(clockFormat))
					}
					dur, _ := span.Duration(rep.now)
					tw.Cell(tbl[3].Width(), hm(dur).String())
					tw.Cell(tbl[4].Width(), pathString(task.Path()))
					fmt.Fprintln(tw.wr)
				}
			}
		}
	}
}
