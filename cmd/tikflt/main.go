package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	stdFilter "git.fractalqb.de/fractalqb/tiktak/internal/filters"
)

var filters = map[string]func() filter{
	"round":   func() filter { return new(stdFilter.Round) },
	"ugap":    func() filter { return new(stdFilter.MicroGap) },
	"dearbzg": func() filter { return new(stdFilter.DEArbZG) },
	"netloc":  func() filter { return new(stdFilter.NetLoc) },
}

func main() {
	list := flag.Bool("l", false, "List filters")
	nowStr := flag.String("t", "", "Set current time")
	flag.Parse()
	if *list {
		listFilters()
		return
	}
	var now time.Time
	if *nowStr == "" {
		now = time.Now()
	} else {
		var err error
		if now, err = time.Parse(time.RFC3339, *nowStr); err != nil {
			log.Fatalf("invalid current time '%s', expect RFC 3339 format", *nowStr)
		}
	}
	var tr tiktak.Task
	tl, err := tiktak.Read(os.Stdin, &tr)
	if err != nil {
		log.Fatal(err)
	}
	for args := flag.Args(); len(args) > 0; {
		fname := args[0]
		newFilter := filters[fname]
		if newFilter == nil {
			log.Fatalf("unknown filter '%s'", fname)
		}
		filter := newFilter()
		if args, err = filter.Flags(args[1:]); err != nil {
			log.Fatalf("%s: %s", fname, err)
		}
		if err := filter.Filter(&tl, now); err != nil {
			log.Fatalf("%s: %s", fname, err)
		}
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

type filter interface {
	Flags(args []string) ([]string, error)
	Filter(tl *tiktak.TimeLine, now time.Time) error
}
