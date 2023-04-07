package tiktak

import (
	"fmt"
	"testing"
	"time"
)

type sw struct {
	w time.Time
	t *Task
}

func expectTL(t *testing.T, tl TimeLine, sws ...sw) error {
	var last *Switch
	for i, ts := range tl {
		if last != nil && last.next != ts {
			t.Errorf("corrupt switch link %d -> %d", i-1, i)
		}
		last = ts
	}
	if last != nil && last.next != nil {
		t.Errorf("last switch has link to %s", last.next.when)
	}

	var errCount int
	l := len(tl)
	if l != len(sws) {
		errCount++
		t.Errorf("timeline has %d switches, want %d", l, len(sws))
		if len(sws) < l {
			l = len(sws)
		}
	}
	for i := 0; i < l; i++ {
		s := sws[i]
		ts := tl[i]
		if s.w != ts.When() {
			errCount++
			t.Errorf("switch %d: time %s != %s", i, s.w, ts.When())
		}
		if s.t != ts.Task() {
			errCount++
			t.Errorf("switch %d: task %s != %s", i, s.t, ts.Task())
		}
	}
	if errCount > 0 {
		return fmt.Errorf("timeline has %d errors", errCount)
	}
	return nil
}

func TestTimeLine_Switch(t *testing.T) {
	now := time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC)
	dt := func(d int) time.Time { return now.Add(time.Duration(d) * time.Minute) }
	t.Run("1st switch", func(t *testing.T) {
		var tr Task
		var tl TimeLine
		tt := tr.Get("test")
		tl.Switch(now, tt)
		expectTL(t, tl, sw{now, tt})
	})

	setup := func() (rt, t0 *Task, tl TimeLine) {
		rt = new(Task)
		t0 = rt.Get("task0")
		tl.Switch(now, t0)
		return
	}
	t.Run("switch before", func(t *testing.T) {
		rt, t0, tl := setup()
		tt := rt.Get("test")
		tl.Switch(dt(-15), tt)
		expectTL(t, tl, sw{dt(-15), tt}, sw{now, t0})
	})
	t.Run("switch before same", func(t *testing.T) {
		_, t0, tl := setup()
		tl.Switch(dt(-15), t0)
		expectTL(t, tl, sw{dt(-15), t0})
	})
	t.Run("switch before off", func(t *testing.T) {
		_, t0, tl := setup()
		tl.Switch(dt(-15), nil)
		expectTL(t, tl, sw{now, t0})
	})
	t.Run("switch after", func(t *testing.T) {
		rt, t0, tl := setup()
		tt := rt.Get("test")
		tl.Switch(dt(15), tt)
		expectTL(t, tl, sw{now, t0}, sw{dt(15), tt})
	})
	t.Run("switch after same", func(t *testing.T) {
		_, t0, tl := setup()
		tl.Switch(dt(15), t0)
		expectTL(t, tl, sw{now, t0})
	})
	t.Run("switch at", func(t *testing.T) {
		rt, _, tl := setup()
		tt := rt.Get("test")
		tl.Switch(now, tt)
		expectTL(t, tl, sw{now, tt})
	})
	t.Run("switch at same", func(t *testing.T) {
		_, t0, tl := setup()
		tl.Switch(now, t0)
		expectTL(t, tl, sw{now, t0})
	})

	setup2 := func() (rt, t0, t1 *Task, tl TimeLine) {
		rt, t0, tl = setup()
		t1 = rt.Get("task1")
		tl.Switch(dt(30), t1)
		return
	}
	t.Run("between", func(t *testing.T) {
		rt, t0, t1, tl := setup2()
		tt := rt.Get("test")
		tl.Switch(dt(15), tt)
		expectTL(t, tl, sw{now, t0}, sw{dt(15), tt}, sw{dt(30), t1})
	})
	t.Run("between to t0", func(t *testing.T) {
		_, t0, t1, tl := setup2()
		tl.Switch(dt(15), t0)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), t1})
	})
	t.Run("between to t1", func(t *testing.T) {
		_, t0, t1, tl := setup2()
		tl.Switch(dt(15), t1)
		expectTL(t, tl, sw{now, t0}, sw{dt(15), t1})
	})

	setup3 := func(t0eqt2 bool) (rt, t0, t1, t2 *Task, tl TimeLine) {
		rt, t0, t1, tl = setup2()
		if t0eqt2 {
			t2 = t0
		} else {
			t2 = rt.Get("task2")
		}
		tl.Switch(dt(60), t2)
		return
	}
	t.Run("change", func(t *testing.T) {
		rt, t0, _, t2, tl := setup3(false)
		tt := rt.Get("test")
		tl.Switch(dt(30), tt)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), tt}, sw{dt(60), t2})
	})
	t.Run("change to t1", func(t *testing.T) {
		_, t0, t1, t2, tl := setup3(false)
		tl.Switch(dt(30), t1)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), t1}, sw{dt(60), t2})
	})
	t.Run("change to t0", func(t *testing.T) {
		_, t0, _, t2, tl := setup3(false)
		tl.Switch(dt(30), t0)
		expectTL(t, tl, sw{now, t0}, sw{dt(60), t2})
	})
	t.Run("change to t2", func(t *testing.T) {
		_, t0, _, t2, tl := setup3(false)
		tl.Switch(dt(30), t2)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), t2})
	})

	t.Run("change between", func(t *testing.T) {
		rt, t0, _, _, tl := setup3(true)
		tt := rt.Get("test")
		tl.Switch(dt(30), tt)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), tt}, sw{dt(60), t0})
	})
	t.Run("change between to t1", func(t *testing.T) {
		_, t0, t1, _, tl := setup3(true)
		tl.Switch(dt(30), t1)
		expectTL(t, tl, sw{now, t0}, sw{dt(30), t1}, sw{dt(60), t0})
	})
	t.Run("change between to t0", func(t *testing.T) {
		_, t0, _, _, tl := setup3(true)
		tl.Switch(dt(30), t0)
		expectTL(t, tl, sw{now, t0})
	})
}

func TestTimeLine_Del(t *testing.T) {
	var rt Task
	var tl TimeLine
	t0, t1, t2 := rt.Get("task0"), rt.Get("task1"), rt.Get("task2")
	now := time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC)
	tl.Switch(now.Add(-30*time.Minute), t0)
	tl.Switch(now, t1)
	tl.Switch(now.Add(30*time.Minute), t2)
	tl.Del(now.Add(10*time.Minute), ShiftAll, ShiftAll)
	expectTL(t, tl,
		sw{now.Add(-20 * time.Minute), t0},
		sw{now.Add(10 * time.Minute), t2},
	)
}

func TestTimeLine_Pick(t *testing.T) {
	var tasks Task
	var tl TimeLine
	ts := time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC)
	ti := tl.Switch(ts, tasks.Get("foo"))
	if n := tl[ti].Task().Name(); n != "foo" {
		t.Fatalf("wrong task '%s'", n)
	}
	t.Run("before", func(t *testing.T) {
		pi, pt := tl.Pick(ts.Add(-time.Second))
		if pi != -1 {
			t.Errorf("unexpected index %d, want -1", pi)
		}
		if pt != nil {
			t.Errorf("illegal task: %+v", pt)
		}
	})
	t.Run("same", func(t *testing.T) {
		pi, pt := tl.Pick(ts)
		if pi != ti {
			t.Errorf("unexpected index %d, want %d", pi, ti)
		}
		if pt == nil {
			t.Error("empty switch to task")
		}
	})
	t.Run("after", func(t *testing.T) {
		pi, pt := tl.Pick(ts.Add(time.Second))
		if pi != ti {
			t.Errorf("unexpected index %d, want %d", pi, ti)
		}
		if pt == nil {
			t.Error("empty switch to task")
		}
	})
}

func ExampleTimeLine() {
	root := new(Task)
	var tl TimeLine
	show := func(i int) {
		sw := tl[i]
		fmt.Printf("%d %s %s", i, sw.When(), sw.Task())
		if n := sw.Next(); n != nil {
			fmt.Printf(" > %s", n.Task())
		}
		fmt.Println()
	}
	t := time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC)
	show(tl.Switch(t, root.Get("1")))
	show(tl.Switch(t.Add(time.Hour), root.Get("2")))
	show(tl.Switch(t.Add(-time.Hour), root.Get("3")))
	show(tl.Switch(t.Add(-30*time.Minute), root.Get("4")))
	show(tl.Switch(t.Add(30*time.Minute), root.Get("5")))
	// Output:
	// 0 2023-04-01 12:00:00 +0000 UTC /1
	// 1 2023-04-01 13:00:00 +0000 UTC /2
	// 0 2023-04-01 11:00:00 +0000 UTC /3 > /1
	// 1 2023-04-01 11:30:00 +0000 UTC /4 > /1
	// 3 2023-04-01 12:30:00 +0000 UTC /5 > /2
}

func ExampleTimeLine_Duration() {
	root := new(Task)
	var tl TimeLine
	t := time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC)
	filter := func(sw *Switch) bool { return sw.Task() != nil }
	tl.Switch(t, root)
	var d time.Duration
	var s, e time.Time
	print := func(name string) {
		fmt.Print(name, ": ", d)
		if s.IsZero() {
			fmt.Print(" -")
		} else {
			fmt.Print(" ", s.Format(time.TimeOnly))
		}
		if e.IsZero() {
			fmt.Print(" -")
		} else {
			fmt.Print(" ", e.Format(time.TimeOnly))
		}
		fmt.Println()
	}

	now := t.Add(20 * time.Minute)
	d, s, e = tl.Duration(t.Add(-time.Hour), t.Add(-30*time.Minute), now, filter)
	print("open/before")
	d, s, e = tl.Duration(t.Add(-30*time.Minute), t, now, filter)
	print("open/touch")
	d, s, e = tl.Duration(t.Add(-15*time.Minute), t.Add(15*time.Minute), now, filter)
	print("open/x-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), now, filter)
	print("open/at-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), t.Add(-10*time.Minute), filter)
	print("open/future")
	d, s, e = tl.Duration(t.Add(15*time.Minute), t.Add(45*time.Minute), now, filter)
	print("open/after")

	tl.Switch(t.Add(30*time.Minute), nil)
	d, s, e = tl.Duration(t.Add(-time.Hour), t.Add(-30*time.Minute), now, filter)
	print("within/before")
	d, s, e = tl.Duration(t.Add(-30*time.Minute), t, now, filter)
	print("within/touch")
	d, s, e = tl.Duration(t.Add(-15*time.Minute), t.Add(15*time.Minute), now, filter)
	print("within/x-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), now, filter)
	print("within/at-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), t.Add(-10*time.Minute), filter)
	print("within/future")
	d, s, e = tl.Duration(t.Add(15*time.Minute), t.Add(45*time.Minute), now, filter)
	print("within/after")

	now = t.Add(40 * time.Minute)
	d, s, e = tl.Duration(t.Add(-time.Hour), t.Add(-30*time.Minute), now, filter)
	print("after/before")
	d, s, e = tl.Duration(t.Add(-30*time.Minute), t, now, filter)
	print("after/touch")
	d, s, e = tl.Duration(t.Add(-15*time.Minute), t.Add(15*time.Minute), now, filter)
	print("after/x-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), now, filter)
	print("after/at-start")
	d, s, e = tl.Duration(t, t.Add(30*time.Minute), t.Add(-10*time.Minute), filter)
	print("after/future")
	d, s, e = tl.Duration(t.Add(15*time.Minute), t.Add(45*time.Minute), now, filter)
	print("after/after")

	// Output:
	// open/before: 0s - -
	// open/touch: 0s - -
	// open/x-start: 15m0s 12:00:00 -
	// open/at-start: 20m0s 12:00:00 -
	// open/future: -10m0s 12:00:00 -
	// open/after: 5m0s 12:15:00 -
	// within/before: 0s - -
	// within/touch: 0s - -
	// within/x-start: 15m0s 12:00:00 12:15:00
	// within/at-start: 30m0s 12:00:00 12:30:00
	// within/future: 30m0s 12:00:00 12:30:00
	// within/after: 15m0s 12:15:00 12:30:00
	// after/before: 0s - -
	// after/touch: 0s - -
	// after/x-start: 15m0s 12:00:00 12:15:00
	// after/at-start: 30m0s 12:00:00 12:30:00
	// after/future: 30m0s 12:00:00 12:30:00
	// after/after: 15m0s 12:15:00 12:30:00
}
