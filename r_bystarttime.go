package main

import (
	"fmt"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	"git.fractalqb.de/fractalqb/tiktak/txtab"

	"golang.org/x/text/language"

	"golang.org/x/text/collate"
)

func init() {
	reports["spans"] = byStartTimeFactory
}

func byStartTimeFactory(lang language.Tag, args []string) Reporter {
	r := &byStartTimeReport{
		tw: txtab.Writer{
			W: os.Stdout,
			F: newTabFormatter(),
		},
		coll: collate.New(lang),
	}
	return r
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
	tw   txtab.Writer
	coll *collate.Collator
}

func (rep *byStartTimeReport) Generate(root *Task, now time.Time) {
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
	tw := &rep.tw
	tw.AddColumn("↹", 1)
	tw.AddColumn("Start", 5)
	tw.AddColumn("Stop", 5)
	tw.AddColumn("Dur", 5)
	tw.AddColumn("Task", tpWidth)
	fmt.Fprintf(tw.W, "TIMESPANS PER DAY %s:\n", reportMonth(now))
	tw.Header()
	day := 0
	var lastSpan *Span
	for _, start := range starts {
		thisDay := 100*start.Year() + start.YearDay()
		if thisDay != day {
			tw.Hrule()
			tw.RowStart()
			_, kw := start.ISOWeek()
			tw.Cells(-1, fmt.Sprintf("%s; KW%d", start.Format(dateFormat), kw))
			tw.RowEnd()
			tw.Hrule()
			day = thisDay
		}
		tasks := t2ts[start]
		for _, task := range tasks {
			for _, span := range task.Spans {
				if span.Start == start {
					tw.RowStart()
					if lastSpan == nil {
						tw.Cell("")
						lastSpan = new(Span)
						*lastSpan = span
					} else {
						is := IntersectSpans(lastSpan, &span)
						d, fin := is.Duration(now)
						if !fin || d > 0 {
							tw.Cell("⇸" /*↪"*/)
						} else if span.Start.After(*lastSpan.Stop) {
							tw.Cell("⇢")
						} else {
							tw.Cell("")
						}
						*lastSpan = span
					}
					tw.Cell(start.Format(clockFormat))
					if span.Stop == nil {
						tw.Cell("")
					} else {
						tw.Cell(span.Stop.Format(clockFormat))
					}
					dur, _ := span.Duration(now)
					tw.Cell(hm(dur))
					tw.Cell(pathString(task.Path()))
					tw.RowEnd()
				}
			}
		}
	}
}
