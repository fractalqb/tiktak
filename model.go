package tiktak

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"time"
)

type Task struct {
	parent *Task
	name   string
	subs   []*Task
	starts []*Switch
	title  string
}

func (t *Task) Name() string { return t.name }

func (t *Task) Title() string     { return t.title }
func (t *Task) SetTitle(s string) { t.title = s }

func (t *Task) Subtasks() []*Task { return t.subs }

func (t *Task) Is(in *Task) bool {
	if in == nil {
		return t == nil
	}
	for t != nil {
		if t == in {
			return true
		}
		t = t.parent
	}
	return false
}

func cmprName(n, m string) int {
	c := strings.Compare(strings.ToLower(n), strings.ToLower(m))
	if c == 0 {
		c = strings.Compare(n, m)
	}
	return c
}

func (t *Task) Get(path ...string) *Task {
	for _, n := range path {
		l := len(t.subs)
		i := sort.Search(l, func(i int) bool { return cmprName(t.subs[i].name, n) >= 0 })
		if i == l {
			nt := &Task{
				parent: t,
				name:   n,
			}
			t.subs = append(t.subs, nt)
			t = nt
		} else if t.subs[i].name == n {
			t = t.subs[i]
		} else {
			nt := &Task{
				parent: t,
				name:   n,
			}
			if l := len(t.subs); cap(t.subs) > l {
				t.subs = t.subs[:l+1]
				copy(t.subs[i+1:], t.subs[i:])
			} else {
				tmp := make([]*Task, l+1)
				copy(tmp, t.subs[:i])
				copy(tmp[i+1:], t.subs[i:])
				t.subs = tmp
			}
			t.subs[i] = nt
			t = nt
		}
	}
	return t
}

func (t *Task) GetString(p string) *Task {
	if p == "/" {
		return t.Root()
	}
	if path.IsAbs(p) {
		t = t.Root()
		p = p[1:]
	}
	return t.Get(strings.Split(p, "/")...)
}

func (t *Task) Find(tip bool, path ...string) *Task {
	for _, n := range path {
		if n == "" || strings.IndexByte(n, '/') >= 0 {
			return nil
		}
		l := len(t.subs)
		i := sort.Search(l, func(i int) bool { return cmprName(t.subs[i].name, n) >= 0 })
		if i == l {
			return nil
		} else if st := t.subs[i]; st.Name() != n {
			return nil
		} else {
			t = st
		}
	}
	if tip && len(t.subs) > 0 {
		return nil
	}
	return t
}

func (t *Task) FindString(p string) *Task {
	if p == "/" {
		return t.Root()
	}
	if path.IsAbs(p) {
		t = t.Root()
		p = p[1:]
	}
	tip := true
	if strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
		tip = false
	}
	return t.Find(tip, strings.Split(p, "/")...)
}

func pathMatch(tpath, pattern []string, tip bool) (int, error) {
	if len(tpath) < len(pattern) {
		return -1, nil
	}
	if tip {
		start := len(tpath) - len(pattern)
		for pi, pe := range pattern {
			if ok, err := path.Match(pe, tpath[start+pi]); err != nil {
				return -1, err
			} else if !ok {
				return -1, nil
			}
		}
		return start, nil
	}
	for i := 0; i+len(pattern) <= len(tpath); i++ {
		for pi, pe := range pattern {
			if ok, err := path.Match(pe, tpath[i+pi]); err != nil {
				return -1, err
			} else if ok {
				return i, nil
			}
		}
	}
	return -1, nil
}

func (t *Task) Match(tip bool, pattern ...string) (matches []*Task, err error) {
	var match func(*Task) error
	match = func(t *Task) error {
		if i, err := pathMatch(t.Path(), pattern, tip); err != nil {
			return err
		} else if i >= 0 {
			matches = append(matches, t)
		}
		for _, s := range t.subs {
			if err = match(s); err != nil {
				return err
			}
		}
		return nil
	}
	if err = match(t); err != nil {
		return nil, err
	}
	return matches, nil
}

func (t *Task) MatchString(p string) ([]*Task, error) {
	if p == "/" {
		return []*Task{t.Root()}, nil
	}
	if path.IsAbs(p) {
		t = t.Root()
		p = p[1:]
	}
	tip := true
	if strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
		tip = false
	}
	return t.Match(tip, strings.Split(p, "/")...)
}

func (t *Task) Root() *Task {
	if t == nil {
		return nil
	}
	for t.parent != nil {
		t = t.parent
	}
	return t
}

func (t *Task) Path() []string {
	switch {
	case t == nil:
		return nil
	case t.parent == nil:
		return []string{}
	}
	return append(t.parent.Path(), t.name)
}

func (t *Task) String() string {
	if t == nil {
		return "-"
	}
	p := t.Path()
	switch {
	case p == nil:
		return ""
	case len(p) == 0:
		return "/"
	}
	return "/" + strings.Join(p, "/")
}

func (t *Task) addStart(s *Switch) {
	if t == nil {
		return
	}
	l := len(t.starts)
	i := sort.Search(l, func(i int) bool {
		return !t.starts[i].When().Before(s.When())
	})
	if i == l {
		t.starts = append(t.starts, s)
	} else if cap(t.starts) > l {
		t.starts = t.starts[:l+1]
		copy(t.starts[i+1:], t.starts[i:])
	} else {
		tmp := make([]*Switch, l+1)
		copy(tmp, t.starts[:i])
		copy(tmp[i+1:], t.starts[i:])
		t.starts = tmp
	}
	t.starts[i] = s
}

func (t *Task) rmStart(s *Switch) {
	for i, r := range t.starts {
		if r == s {
			copy(t.starts[i:], t.starts[i+1:])
			t.starts = t.starts[:len(t.starts)-1]
			return
		}
	}
}

func (t *Task) Visit(pre bool, do func(*Task) error) error {
	if t == nil {
		return nil
	}
	if pre {
		if err := do(t); err != nil {
			return err
		}
	}
	for _, st := range t.subs {
		if err := st.Visit(pre, do); err != nil {
			return err
		}
	}
	if !pre {
		if err := do(t); err != nil {
			return err
		}
	}
	return nil
}

type Note struct {
	Sym  rune
	Text string
}

type Switch struct {
	to    *Task
	when  time.Time
	next  *Switch
	notes []Note
}

func (s *Switch) Task() *Task     { return s.to }
func (s *Switch) When() time.Time { return s.when }
func (s *Switch) Next() *Switch   { return s.next }
func (s *Switch) Notes() []Note   { return s.notes }

func Warning(n Note) bool { return n.Sym != 0 }

func (s *Switch) SelectNotes(idxs []int, f func(Note) bool) []int {
	for i, note := range s.notes {
		if f(note) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (s *Switch) AddNote(text string) (i int) {
	i = len(s.notes)
	s.notes = append(s.notes, Note{Text: text})
	return i
}

func (s *Switch) AddWarning(sym rune, text string) (i int) {
	i = len(s.notes)
	s.notes = append(s.notes, Note{Sym: sym, Text: text})
	return i
}

func (s *Switch) DelNote(i int) {
	if len(s.notes) == 0 {
		return
	}
	if i < 0 {
		i += len(s.notes)
	}
	copy(s.notes[i:], s.notes[i+1:])
	s.notes = s.notes[:len(s.notes)-1]
}

func (s *Switch) FilterNotes(f func(Note) bool) {
	var tmp []Note
	for _, note := range s.notes {
		if f(note) {
			tmp = append(tmp, note)
		}
	}
	s.notes = tmp
}

func Infinite(d time.Duration) bool { return d < 0 }

func (s *Switch) Duration() time.Duration {
	if s.next == nil {
		return -1
	}
	return s.Next().When().Sub(s.When())
}

func (s *Switch) reset() {
	if s != nil {
		if s.to != nil {
			s.to.rmStart(s)
			s.to = nil
		}
		s.when = time.Time{}
		s.next = nil
	}
}

type TimeLine []*Switch

func (tl TimeLine) FirstTask() *Task {
	for _, s := range tl {
		if t := s.Task(); t != nil {
			return t
		}
	}
	return nil
}

func (tl TimeLine) RootTask() *Task {
	t := tl.FirstTask()
	if t == nil {
		return nil
	}
	return t.Root()
}

// Pick returns the latest task switch that happens at or before t. If t is
// before all task switches Pick retunrns -1, nil. Otherwies it retuns the index
// of the switch in the TimeLine and the switch itself.
func (ts TimeLine) Pick(t time.Time) (int, *Switch) {
	l := len(ts)
	if l == 0 {
		return -1, nil
	} else if !t.Before(ts[l-1].When()) { // standard case
		l--
		return l, ts[l]
	}
	i := sort.Search(l, func(i int) bool { return t.Before(ts[i].When()) })
	if i == 0 {
		return -1, nil
	}
	i--
	return i, ts[i]
}

func (tl *TimeLine) Switch(at time.Time, to *Task) int {
	i, sw := tl.Pick(at)
	var next *Switch
	if sw != nil {
		if sw.When().Equal(at) {
			return tl.changeTask(i, sw, to)
		}
		if sw.Task() == to { // to already running at
			return i
		}
		next = sw.Next()
	} else if to == nil {
		return -1
	} else if len(*tl) > 0 {
		next = (*tl)[0]
	}
	i++
	if next != nil && next.Task() == to { // next starts earlier
		next.when = at
		return i
	}
	if l := len(*tl); cap(*tl) > l {
		*tl = (*tl)[:l+1]
		copy((*tl)[i+1:], (*tl)[i:])
	} else {
		tmp := make(TimeLine, l+1)
		copy(tmp, (*tl)[:i])
		copy(tmp[i+1:], (*tl)[i:])
		*tl = tmp
	}
	s := &Switch{to: to, when: at, next: next}
	(*tl)[i] = s
	if i > 0 {
		(*tl)[i-1].next = s
	}
	if to != nil {
		to.addStart(s)
	}
	return i
}

func (tl *TimeLine) changeTask(i int, of *Switch, to *Task) int {
	if of.Task() == to {
		return i
	}
	var p *Switch
	if i > 0 {
		p = (*tl)[i-1]
	}
	if n := of.Next(); n != nil && n.Task() == to {
		if t := of.Task(); t != nil {
			t.rmStart(of)
		}
		copy((*tl)[i:], (*tl)[i+1:])
		*tl = (*tl)[:len(*tl)-1]
		if p != nil {
			p.next = n
		}
		n.when = of.when
		of = n
	}
	if p == nil || p.Task() != to {
		if of.to != nil {
			of.to.rmStart(of)
		}
		of.to = to
		if to != nil {
			to.addStart(of)
		}
		return i
	}
	if t := of.Task(); t != nil {
		t.rmStart(of)
	}
	copy((*tl)[i:], (*tl)[i+1:])
	*tl = (*tl)[:len(*tl)-1]
	p.next = of.next
	return i - 1
}

func (tl *TimeLine) ClipBefore(t time.Time) {
	i, rt := tl.Pick(t)
	for _, s := range (*tl)[:i] {
		s.reset()
	}
	if rt.Task() == nil {
		rt.reset()
		*tl = (*tl)[i+1:]
		return
	}
	rt.when = t
}

func (tl *TimeLine) ClipAfter(t time.Time) {
	i, rt := tl.Pick(t)
	for _, s := range (*tl)[i+1:] {
		s.reset()
	}
	if rt.Task() == nil {
		return
	}
	if rt.When().Equal(t) {
		rt.reset()
		rt.when = t
	} else {
		tl.Switch(t, nil)
	}
}

func (tl *TimeLine) Clip(start, end time.Time) {
	tl.ClipBefore(start)
	tl.ClipAfter(end)
}

func (tl *TimeLine) Reschedule(i int, to time.Time) error {
	sw := (*tl)[i]
	if to.Before(sw.When()) {
		if i == 0 || (*tl)[i-1].When().Before(to) {
			sw.when = to
		} else {
			return fmt.Errorf("new time %s not after previous switch at %s",
				to,
				(*tl)[i-1].When(),
			)
		}
	} else if to.After(sw.When()) {
		if sw.next == nil || to.Before(sw.Next().When()) {
			sw.when = to
		} else {
			return fmt.Errorf("new time %s not before next switch at %s",
				to,
				sw.next.When(),
			)
		}
	}
	return nil
}

type SelectSwitch = func(int, *Switch) bool

func AllSwitch(int, *Switch) bool { return true }

func NonNilTask(_ int, sw *Switch) bool {
	return sw.Task() != nil
}

func (tl *TimeLine) Insert(
	at time.Time, to *Task,
	dtPast time.Duration, swPast SelectSwitch,
	dtFutr time.Duration, swFutr SelectSwitch,
) int {
	if dtPast > 0 {
		dtPast = 0
	}
	if dtFutr < 0 {
		dtFutr = 0
	}
	i, sw := tl.Pick(at)
	if dtPast < 0 && i >= 0 {
		mi := i
		if at.Equal(sw.When()) {
			mi--
		}
		var t time.Time
		for mi >= 0 && swPast(mi, (*tl)[mi]) {
			t = (*tl)[mi].when.Add(dtPast)
			(*tl)[mi].when = t
			mi--
		}
		if !t.IsZero() {
			for mi >= 0 && !(*tl)[mi].When().Before(t) {
				tl.DelSwitch(mi)
				mi--
				i--
			}
		}
		if (*tl)[i] != sw {
			panic("failed to compute switch index")
		}
	}
	if dtFutr > 0 {
		mi := i
		if mi < 0 {
			mi = 0
		} else if at.After(sw.When()) {
			mi++
		}
		var t time.Time
		for mi < len(*tl) && swFutr(mi, (*tl)[mi]) {
			t := (*tl)[mi].when.Add(dtFutr)
			(*tl)[mi].when = t
			mi++
		}
		if !t.IsZero() {
			for mi < len(*tl) && !t.Before((*tl)[mi].When()) {
				tl.DelSwitch(mi)
			}
		}
	}
	i = tl.Switch(at.Add(dtPast), to)
	if i+1 == len(*tl) {
		tl.Switch(at.Add(dtFutr), nil)
	}
	return i
}

func (tl *TimeLine) DelSwitch(i int) error {
	if i < 0 || i >= len(*tl) {
		return fmt.Errorf("invalid switch index: %d", i)
	}
	if i == 0 {
		(*tl)[i].next = nil
		(*tl)[i] = nil
		*tl = (*tl)[1:]
		return nil
	}
	pSw, dSw := (*tl)[i-1], (*tl)[i]
	if dSw.next != nil && pSw.Task() == dSw.next.Task() {
		pSw.next = dSw.next.next
		copy((*tl)[i:], (*tl)[i+2:])
		*tl = (*tl)[:len(*tl)-2]
	} else {
		pSw.next = dSw.next
		copy((*tl)[i:], (*tl)[i+1:])
		*tl = (*tl)[:len(*tl)-1]
	}
	return nil
}

// If at == start & pre != nil && post == nil => shift the past
func (tl *TimeLine) Del(at time.Time, pre, post SelectSwitch) {
	di, sw := tl.Pick(at)
	if di < 0 {
		return
	}
	tl.DelSwitch(di)
	dd := at.Sub(sw.When())
	if pre != nil {
		if dd == 0 && post == nil && sw.next != nil {
			dd = sw.Duration()
		}
		if dd > 0 {
			for i := di - 1; i >= 0; i-- {
				ms := (*tl)[i]
				if !pre(di-i, ms) {
					break
				}
				ms.when = ms.when.Add(dd)
			}
		}
	}
	if ns := sw.Next(); post != nil && ns != nil {
		dd = at.Sub(ns.When())
		i := 1
		if dd < 0 {
			for ns != nil && post(i, ns) {
				ns.when = ns.when.Add(dd)
				ns = ns.next
				i++
			}
		}
	}
}

func SameTask(t *Task) func(*Switch) bool {
	return func(sw *Switch) bool { return sw.Task() == t }
}

func IsATask(t *Task) func(*Switch) bool {
	return func(sw *Switch) bool { return sw.Task().Is(t) }
}

func AnyTask(sw *Switch) bool { return sw.Task() != nil }

// Duration computes the sum of durations of all time line switches filtered by
// f. If now switch was considered, s will be zero. Otherwise it is the time of
// the first switch.
func (tl TimeLine) Duration(from, to, now time.Time, f func(*Switch) bool) (d time.Duration, s, e time.Time) {
	if len(tl) == 0 {
		return 0, time.Time{}, time.Time{}
	}
	i, _ := tl.Pick(from)
	if i < 0 {
		i = 0
	}
	var open bool
	for _, sw := range tl[i:] {
		if !sw.When().Before(to) {
			break
		}
		if f(sw) {
			var end time.Time
			if sw.Next() != nil {
				end = sw.Next().When()
			} else if !now.IsZero() {
				end = now
				open = true
			} else {
				d = -1
				break
			}
			start := sw.When()
			if start.Before(from) {
				start = from
			}
			if s.IsZero() {
				s = start
			}
			if to.Before(end) {
				end = to
			}
			if e.IsZero() || e.Before(end) {
				e = end
			}
			d += end.Sub(start)
		}
	}
	if open {
		return d, s, time.Time{}
	}
	return d, s, e
}
