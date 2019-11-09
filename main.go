package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/language"
)

const (
	envDDir     = "TIKTOK_DATA"
	rPrefix     = "  "
	dateFormat  = "Mon, 02 Jan 2006"
	clockFormat = "15:04"
)

var (
	timeFmt string
	reports = make(map[string]reportFactory)
)

func reportNames() (res []string) {
	for nm := range reports {
		res = append(res, nm)
	}
	return res
}

type reportFactory func(now time.Time, lang language.Tag, args []string) func(*Task)

func pathString(p []string) string {
	switch len(p) {
	case 0:
		log.Fatal("illegal empty path")
	case 1:
		return "/" + p[0]
	}
	return strings.Join(p, "/")
}

func loadTasks(t time.Time) *Task {
	res := &Task{}
	rd, err := os.Open(dataFile(t))
	if err != nil {
		log.Printf("cannot open '%s': %s", dataFile(t), err)
		return res
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	err = dec.Decode(res)
	if err != nil {
		log.Fatalf("error reading '%s': %s", dataFile(t), err)
	}
	res.WalkAll(nil, func(tp []*Task, p []string) {
		depth := len(tp) - 1
		t := tp[depth]
		if t.Subs == nil {
			t.Subs = make(map[string]*Task)
		}
		if depth > 0 {
			t.parent = tp[depth-1]
		}
	})
	return res
}

func saveTasks(filename string, root *Task) {
	root.WalkAll(nil, func(tp []*Task, _ []string) {
		t := tp[len(tp)-1]
		sort.Slice(t.Spans, func(i, j int) bool {
			return t.Spans[i].Start.Before(t.Spans[j].Start)
		})
	})
	tmp := filename + "~"
	wr, err := os.Create(tmp)
	if err != nil {
		log.Fatalf("cannot create '%s': %s", tmp, err)
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")
	err = enc.Encode(root)
	if err != nil {
		log.Fatalf("cannot write '%s': %s", tmp, err)
	}
	err = wr.Close()
	if err != nil {
		log.Fatalf("error closing '%s': %s", tmp, err)
	}
	os.Rename(tmp, filename)
	if err != nil {
		log.Fatalf("error renaming '%s' to '%s': %s", tmp, filename, err)
	}
}

var (
	dataFileNm string
	microGap   time.Duration
	flagLang   string
)

func dataFile(t time.Time) string {
	dir := os.Getenv(envDDir)
	if dir == "" {
		dir = "."
	}
	if dataFileNm == "" {
		dataFileNm = t.Format("tiktok-2006-01.json")
	}
	if strings.IndexByte(dataFileNm, '/') >= 0 {
		return dataFileNm
	}
	return filepath.Join(dir, dataFileNm)
}

type hm time.Duration

func (d hm) String() string {
	hs := time.Duration(d).Hours()
	h := int(hs)
	var sb strings.Builder
	switch timeFmt {
	case "hf":
		fmt.Fprintf(&sb, "%05.2f", hs)
	case "hm":
		m := int(math.Round(60 * (hs - float64(h))))
		if m >= 60 {
			h++
			m -= 60
		}
		fmt.Fprintf(&sb, "%02d", int(h))
		sb.WriteByte(':')
		fmt.Fprintf(&sb, "%02d", m)
	default:
		log.Fatalf("not a time format: '%s'", timeFmt)
	}
	return sb.String()
}

func findPath(root *Task, pat []string) (res *Task, err error) {
	root.Walk(nil, func(tp []*Task, nmp []string) bool {
		if len(nmp) > 0 {
			if PathMatch(nmp[1:], pat) {
				if res != nil {
					res = nil
					err = fmt.Errorf(
						"task pattern '%s' is ambiguous",
						pathString(pat))
					return true
				}
				res = tp[len(tp)-1]
			}
		}
		return false
	})
	return res, err
}

func startKnown(t time.Time, root *Task, pat []string) {
	selected, err := findPath(root, pat)
	if err != nil {
		log.Fatal(err)
	}
	if selected == nil {
		log.Fatalf("no task matches '%s'", pathString(pat))
	}
	selected.Start(t)
}

func reportMonth(t time.Time) string {
	return t.Format("01/2006")
}

func parseTime(tstr string) time.Time {
	if tstr == "" {
		return time.Now().Round(time.Second)
	}
	t, err := time.ParseInLocation("2006-01-02T15:04", tstr, time.Local)
	if err == nil {
		return t
	}
	t, err = time.ParseInLocation("15:04", tstr, time.Local)
	if err == nil {
		n := time.Now()
		return time.Date(
			n.Year(), n.Month(), n.Day(),
			t.Hour(), t.Minute(), n.Second(),
			0, n.Location())
	}
	t, err = time.ParseInLocation("01/2006", tstr, time.Local)
	if err != nil {
		log.Fatal(err)
	}
	n := time.Now()
	return time.Date(
		t.Year(), t.Month(),
		n.Day(), n.Hour(), n.Minute(), n.Second(),
		0, n.Location())
}

func main() {
	flag.StringVar(&dataFileNm, "f", "",
		`explicitly choose data file.
When not explicitly selected tiktok will look in the directory given
in the `+envDDir+` environment variable.`)
	flag.DurationVar(&microGap, "ugap", time.Minute,
		`length of µ-gap`)
	flag.StringVar(&flagLang, "lang", "",
		`select language`)
	flag.StringVar(&timeFmt, "d", "hm",
		`select format for durations: hm, hf`)
	stop := flag.Bool("stop", false,
		`stop all running clocks`)
	printFile := flag.Bool("print-file", false,
		`print data file name`)
	flagNow := flag.String("t", "",
		`Set current time for operation. Missing elements are taken from now().
Formats:
 - yyyy-mm-ddTHH:MM
 - mm/yyyy
 - HH:MM`)
	//title := flag.String("title", "", "set task's title")
	report := flag.String("r", "",
		`select report: `+strings.Join(reportNames(), ", "))
	flag.Parse()
	t := parseTime(*flagNow)
	if *printFile {
		fmt.Println(dataFile(t))
		return
	}
	var lang language.Tag
	if flagLang != "" {
		lang = language.Make(flagLang)
	} else {
		lang = language.Make(os.Getenv("LANG"))
	}
	root := loadTasks(t)
	switch {
	case *stop:
		root.WalkAll(nil, CloseOpenSpans{t}.Do)
		saveTasks(dataFile(t), root)
	case *report != "":
		rep := reports[*report](t, lang, flag.Args())
		rep(root)
	case len(flag.Args()) == 0:
		defaultOutput(os.Stdout, root, t, lang)
	default:
		if len(flag.Args()) != 1 {
			log.Fatal("cannot start more than one task")
		}
		pstr := strings.TrimSpace(flag.Arg(0))
		switch {
		case pstr == "" || pstr == "/":
			root.WalkAll(nil, CloseOpenSpans{t}.Do)
			root.Start(t)
		case pstr[0] == '/':
			root.WalkAll(nil, CloseOpenSpans{t}.Do)
			path := strings.Split(pstr, "/")
			root.Get(path[1:]...).Start(t)
		default:
			root.WalkAll(nil, CloseOpenSpans{t}.Do)
			path := strings.Split(pstr, "/")
			startKnown(t, root, path)
		}
	}
	saveTasks(dataFile(t), root)
}
