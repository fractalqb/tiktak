package main

import (
	"log"
	"time"

	"git.fractalqb.de/fractalqb/tiktak/tmth"
	"golang.org/x/text/collate"
)

func PathMatch(path, pattern []string) bool {
	pth, ptt := len(path)-1, len(pattern)-1
	if ptt < 0 || ptt > pth {
		return false
	}
	for ptt >= 0 {
		if path[pth] != pattern[ptt] {
			return false
		}
		ptt--
		pth--
	}
	return true
}

type Span struct {
	Start time.Time              `json:"start"`
	Stop  *time.Time             `json:"stop,omitempty"`
	Tags  map[string]interface{} `json:"tags,omitempty"`
}

func DaySpan(t time.Time) Span {
	res := Span{
		Start: tmth.StartDay(t, 0, nil),
		Stop:  new(time.Time),
	}
	*res.Stop = tmth.AddDay(res.Start, 1, nil)
	return res
}

// TODO Week always starts monday
func WeekSpan(t time.Time) Span {
	res := Span{
		Start: tmth.StartDay(tmth.LastDay(time.Monday, t, nil), 0, nil),
		Stop:  new(time.Time),
	}
	*res.Stop = tmth.AddDay(res.Start, 7, nil)
	return res
}

func CoverSpans(s1, s2 *Span) Span {
	if s2.Start.Before(s1.Start) {
		s1, s2 = s2, s1
	}
	res := Span{Start: s1.Start}
	if s1.Stop != nil && s2.Stop != nil {
		res.Stop = new(time.Time)
		if s1.Stop.Before(*s2.Stop) {
			*res.Stop = *s2.Stop
		} else {
			*res.Stop = *s1.Stop
		}
	}
	return res
}

func IntersectSpans(s1, s2 *Span) Span {
	if s2.Start.Before(s1.Start) {
		s1, s2 = s2, s1
	}
	res := Span{Start: s2.Start}
	if s1.Stop == nil {
		res.Stop = s2.Stop
	} else if s1.Stop.Before(res.Start) {
		res.Stop = &res.Start
	} else if s2.Stop == nil {
		res.Stop = s1.Stop
	} else if s1.Stop.Before(*s2.Stop) {
		res.Stop = s1.Stop
	} else {
		res.Stop = s2.Stop
	}
	return res
}

func (s *Span) Includes(t time.Time) bool {
	if t.Before(s.Start) {
		return false
	}
	if s.Stop == nil {
		return true
	}
	return t.Before(*s.Stop)
}

func (s *Span) Duration(now time.Time) (dt time.Duration, finite bool) {
	if s.Stop == nil {
		return now.Sub(s.Start), false
	}
	return s.Stop.Sub(s.Start), true
}

type Task struct {
	parent *Task
	Tags   map[string]interface{} `json:"tags,omitempty"`
	Spans  []Span                 `json:"spans,omitempty"`
	Subs   map[string]*Task       `json:"tasks,omitempty"`
}

func (t *Task) Path() []string {
	if t.parent == nil {
		return []string{""}
	}
	for tag, sub := range t.parent.Subs {
		if sub == t {
			return append(t.parent.Path(), tag)
		}
	}
	log.Fatal("invalid child")
	return nil
}

func (t *Task) Find(path ...string) *Task {
	for _, nm := range path {
		if t = t.Subs[nm]; t == nil {
			return nil
		}
	}
	return t
}

func (t *Task) Get(path ...string) *Task {
	for _, nm := range path {
		if len(nm) == 0 {
			return nil
		}
		sub, ok := t.Subs[nm]
		if !ok {
			sub = &Task{parent: t, Subs: make(map[string]*Task)}
			t.Subs[nm] = sub
		}
		t = sub
	}
	return t
}

func (t *Task) Start(at time.Time) bool {
	ext := 0
	for _, span := range t.Spans {
		if span.Start.After(at) {
			continue
		}
		if span.Stop == nil {
			ext++
			continue
		}
		if !span.Stop.Before(at) {
			span.Stop = nil
			ext++
		}
	}
	if ext == 0 {
		t.Spans = append(t.Spans, Span{Start: at})
		return true
	}
	return false
}

func (t *Task) DelSpan(i int) {
	copy(t.Spans[i:], t.Spans[i+1:])
	t.Spans = t.Spans[:len(t.Spans)-1]
}

func (t *Task) WalkAll(
	coll *collate.Collator,
	do func(tPath []*Task, nmPath []string),
) {
	t.Walk(coll, func(tPath []*Task, nmPath []string) bool {
		do(tPath, nmPath)
		return false
	})
}

func (t *Task) Walk(
	coll *collate.Collator,
	do func(tPath []*Task, nmPath []string) (done bool),
) {
	var walk func(*Task)
	done := false
	tp := []*Task{t}
	np := []string{""}
	walk = func(t *Task) {
		done = do(tp, np)
		if done || len(t.Subs) == 0 {
			return
		}
		var nms []string
		for nm := range t.Subs {
			nms = append(nms, nm)
		}
		if coll != nil {
			coll.SortStrings(nms)
		}
		sub := t.Subs[nms[0]]
		i := len(tp)
		tp = append(tp, sub)
		np = append(np, nms[0])
		walk(sub)
		if done {
			return
		}
		for _, nm := range nms[1:] {
			sub = t.Subs[nm]
			tp[i] = sub
			np[i] = nm
			walk(sub)
			if done {
				return
			}
		}
		tp = tp[:i]
		np = np[:i]
	}
	walk(t)
}

type CloseAllOpen time.Time

func (cao CloseAllOpen) Do(tp []*Task, nmp []string) {
	t := tp[len(tp)-1]
	for i := range t.Spans {
		span := &t.Spans[i]
		if span.Stop == nil {
			log.Printf("in task %v close span %d starting %s", nmp, i, span.Start)
			span.Stop = new(time.Time)
			*span.Stop = time.Time(cao)
		}
	}
}

type UGap struct {
	Task *Task
	Span int
}

type FindUGaps struct {
	To        time.Time
	TooShort  []UGap
	EndBefore []UGap
}

func (cfn *FindUGaps) Do(tp []*Task, nmp []string) {
	t := tp[len(tp)-1]
	for i := range t.Spans {
		span := &t.Spans[i]
		if span.Stop == nil {
			if dt := cfn.To.Sub(span.Start); dt >= 0 && dt < microGap {
				cfn.TooShort = append(cfn.TooShort, UGap{t, i})
			}
		} else if dt := cfn.To.Sub(*span.Stop); dt >= 0 && dt < microGap {
			cfn.EndBefore = append(cfn.EndBefore, UGap{t, i})
		}
	}
}

func isExcept(t *Task, except []*Task) bool {
	for _, e := range except {
		if t == e {
			return true
		}
	}
	return false
}

func CloseForNext(root *Task, t time.Time, except ...*Task) time.Time {
	gaps := FindUGaps{To: t}
	root.WalkAll(nil, gaps.Do)
	for _, eb := range gaps.EndBefore {
		if isExcept(eb.Task, except) {
			continue
		}
		*eb.Task.Spans[eb.Span].Stop = t
		log.Printf("extending span %d of %s to close ugap", eb.Span, pathString(eb.Task.Path()))
	}
	for _, ts := range gaps.TooShort {
		log.Printf("squash uspan %d of task %s", ts.Span, pathString(ts.Task.Path()))
		ts.Task.DelSpan(ts.Span)
	}
	root.WalkAll(nil, func(tp []*Task, nmp []string) {
		task := tp[len(tp)-1]
		if isExcept(task, except) {
			return
		}
		for i := range task.Spans {
			span := &task.Spans[i]
			if span.Stop == nil {
				log.Printf("close span %d of %s starting %s",
					i,
					span.Start,
					pathString(task.Path()))
				span.Stop = new(time.Time)
				*span.Stop = t
			}
		}
	})
	return t
}
