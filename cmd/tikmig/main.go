package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/cmd"
)

var nameDay bool

type inspan struct {
	task *tiktak.Task
	s, e time.Time
}

func main() {
	flag.BoolVar(&nameDay, "day", nameDay, "Use day in file names")
	flag.Parse()
	if len(flag.Args()) == 0 {
		tl := migrate(os.Stdin)
		if err := tiktak.Write(os.Stdout, tl); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, arg := range flag.Args() {
			migrateFile(arg)
		}
	}
}

func migrateFile(name string) {
	r, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	tl := migrate(r)
	name = filepath.Join(filepath.Dir(name), cmd.OutputBasename(tl, nameDay))
	w, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	if err := tiktak.Write(w, tl); err != nil {
		log.Fatal(err)
	}
}

func migrate(r io.Reader) tiktak.TimeLine {
	var (
		root  tiktak.Task
		spans []inspan
		tl    tiktak.TimeLine
	)
	in := ggja.Obj{
		Bare:    make(ggja.BareObj),
		OnError: func(err error) { log.Fatal(err) },
	}
	err := json.NewDecoder(r).Decode(&in.Bare)
	if err != nil {
		log.Fatal(err)
	}
	spans = convert(&root, spans, nil, in)
	sort.Slice(spans, func(i, j int) bool { return spans[i].s.Before(spans[j].s) })
	for _, s := range spans {
		tl.Switch(s.s, s.task)
		if !s.e.IsZero() {
			tl.Switch(s.e, nil)
		}
	}
	return tl
}

func convert(root *tiktak.Task, spans []inspan, path []string, in ggja.Obj) []inspan {
	task := root.Get(path...)
	sps := in.Arr("spans")
	for i := 0; i < sps.Len(); i++ {
		span := sps.Obj(i)
		tStr := span.MStr("start")
		t, err := time.Parse(time.RFC3339, tStr)
		if err != nil {
			log.Fatal(err)
		}
		spn := inspan{task: task, s: t}
		if tStr = span.Str("stop", ""); tStr != "" {
			t, err = time.Parse(time.RFC3339, tStr)
			if err != nil {
				log.Fatal(err)
			}
			spn.e = t
		}
		spans = append(spans, spn)
	}
	tasks := in.Obj("tasks")
	if tasks != nil {
		for task, data := range tasks.Bare {
			path = append(path, task)
			spans = convert(
				root,
				spans,
				path,
				ggja.Obj{Bare: data.(ggja.BareObj), OnError: in.OnError},
			)
			path = path[:len(path)-1]
		}
	}
	return spans
}
