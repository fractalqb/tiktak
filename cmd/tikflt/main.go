package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	stdFilter "git.fractalqb.de/fractalqb/tiktak/internal/filters"
)

type filter interface {
	Filter(tl *tiktak.TimeLine) error
}

var filters = map[string]func() filter{
	"round": func() filter { return new(stdFilter.Round) },
	"ugap":  func() filter { return new(stdFilter.MicroGap) },
}

func main() {
	list := flag.Bool("l", false, "List filters")
	flag.Parse()
	if *list {
		listFilters()
		return
	}
	var tr tiktak.Task
	tl, err := tiktak.Read(os.Stdin, &tr)
	if err != nil {
		log.Fatal(err)
	}
	cl := flag.CommandLine
	for len(cl.Args()) > 0 {
		fname := cl.Arg(0)
		newFilter := filters[fname]
		if newFilter == nil {
			log.Fatalf("unknown filter '%s'", fname)
		}
		filter := newFilter()
		tmp := flag.NewFlagSet(fname, flag.ExitOnError)
		filterFlags(tmp, filter)
		if err := tmp.Parse(cl.Args()[1:]); err != nil {
			log.Fatal(err)
		}
		if err := filter.Filter(&tl); err != nil {
			log.Fatalf("%s: %s", fname, err)
		}
		cl = tmp
	}
	if err := tiktak.Write(os.Stdout, tl); err != nil {
		log.Fatal(err)
	}
}

func listFilters() {
	var names []string
	for n := range filters {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Println(strings.Join(names, ", "))
}

func filterFlags(fs *flag.FlagSet, f filter) {
	cfgv := reflect.Indirect(reflect.ValueOf(f))
	cfgt := cfgv.Type()
	for i := 0; i < cfgt.NumField(); i++ {
		f := cfgt.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("tikf")
		var doc string
		if tag == "-" {
			continue
		} else if tag == "" {
			tag = strings.ToLower(f.Name)
		} else {
			tmp := strings.Split(tag, ",")
			if tmp[0] == "" {
				tag = strings.ToLower(f.Name)
			} else {
				tag = tmp[0]
			}
			if len(tmp) > 1 {
				doc = tmp[1]
			}
		}
		switch {
		case f.Type == reflect.TypeOf(time.Minute):
			fs.DurationVar(
				cfgv.Field(i).Addr().Interface().(*time.Duration),
				tag,
				cfgv.Field(i).Interface().(time.Duration),
				doc,
			)
		}
	}
}
