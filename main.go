package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"git.fractalqb.de/fractalqb/tiktak/txtab"
)

//go:generate versioner -bno build_no ./VERSION ./version.go

const (
	envDDir     = "TIKTAK_DATA"
	envTmpl     = "TIKTAK_TEMPLATE"
	rPrefix     = "  "
	clockFormat = "15:04"
)

var (
	timeFmt    string
	reports    = make(map[string]reportFactory)
	dateFormat string
	lang       language.Tag
	msgPr      *message.Printer
	csvOut     string
	version    = fmt.Sprintf("%d.%d.%d-%s+%d", Major, Minor, Patch, Quality, BuildNo)
)

func newTabFormatter() txtab.Formatter {
	if csvOut == "" {
		return &txtab.Table{Prefix: rPrefix}
	}
	return &txtab.Csv{FSep: csvOut}
}

func reportNames() (res []string) {
	for nm := range reports {
		res = append(res, nm)
	}
	return res
}

type reportFactory func(lang language.Tag, args []string) Reporter

type Reporter interface {
	Generate(root *Task, now time.Time)
}

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
		rd, err = os.Open(templateFile())
		if err != nil {
			log.Printf("cannot open '%s': %s", dataFile(t), err)
			return res
		} else {
			log.Printf("loading template file '%s'", templateFile())
		}
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
		if t.Tags == nil {
			t.Tags = make(map[string]interface{})
		}
		if depth > 0 {
			t.parent = tp[depth-1]
		}
		sort.Slice(t.Spans, func(i, j int) bool {
			return t.Spans[i].Start.Before(t.Spans[j].Start)
		})
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
		dataFileNm = t.Format("tiktak-2006-01.json")
	}
	if strings.IndexRune(dataFileNm, filepath.Separator) >= 0 {
		return dataFileNm
	}
	return filepath.Join(dir, dataFileNm)
}

func templateFile() string {
	tmplFileNm := os.Getenv(envTmpl)
	if tmplFileNm == "" {
		tmplFileNm = "template.json"
	}
	if strings.IndexRune(tmplFileNm, filepath.Separator) >= 0 {
		return tmplFileNm
	}
	if dir := os.Getenv(envDDir); dir == "" {
		return tmplFileNm
	} else {
		return filepath.Join(dir, tmplFileNm)
	}
}

type hm time.Duration

func (d hm) String() string {
	hs := time.Duration(d).Hours()
	h := int(hs)
	var sb strings.Builder
	switch timeFmt {
	case "f":
		msgPr.Fprintf(&sb, "%05.2f", hs)
	case "c":
		fmt.Fprintf(&sb, "%05.2f", hs)
	case "m":
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

func reportMonth(t time.Time) string {
	return t.Format("01/2006")
}

var relTimeRegexp = regexp.MustCompile(`^([ymwd])-(\d+)(?:T(\d\d:\d\d))?$`)

func parseRelTime(rt string) (time.Time, bool) {
	match := relTimeRegexp.FindStringSubmatch(rt)
	if match == nil {
		return time.Time{}, false
	}
	diff, err := strconv.Atoi(match[2])
	if err != nil {
		log.Fatal(err)
	}
	res := time.Now()
	if match[3] != "" {
		t, err := time.ParseInLocation("15:04", match[3], res.Location())
		if err != nil {
			log.Fatal(err)
		}
		res = time.Date(
			res.Year(), res.Month(), res.Day(),
			t.Hour(), t.Minute(), res.Second(), res.Nanosecond(),
			res.Location())
	}
	ye, mo, da := res.Date()
	switch match[1] {
	case "d":
		da -= diff
	case "w":
		da -= 7 * diff
	case "m":
		mo -= time.Month(diff)
	case "y":
		ye -= diff
	}
	res = time.Date(
		ye, mo, da,
		res.Hour(), res.Minute(), res.Second(), res.Nanosecond(),
		res.Location())
	return res, true
}

func parseTime(tstr string) time.Time {
	if tstr == "" {
		return time.Now().Round(time.Second)
	}
	if res, ok := parseRelTime(tstr); ok {
		return res
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
	flag.Usage = usage
	flag.StringVar(&dataFileNm, "f", "", flagDocFile)
	flag.DurationVar(&microGap, "ugap", time.Minute, flagDocUgap)
	flag.StringVar(&flagLang, "lang", "", flagDocLang)
	flag.StringVar(&timeFmt, "d", "m", flagDocDurFmt)
	stop := flag.Bool("zzz", false, flagDocZzz)
	printFile := flag.Bool("print-file", false, flagDocPFile)
	flagNow := flag.String("t", "", flagDocNow)
	flag.StringVar(&dateFormat, "date", "Mon, 02 Jan 2006", flagDocDateFmt)
	flag.StringVar(&csvOut, "csv", "", flagDocCsv)
	//title := flag.String("title", "", "set task's title")
	report := flag.String("r", "", flagDocReport())
	flag.Parse()
	t := parseTime(*flagNow)
	if *printFile {
		fmt.Println(dataFile(t))
		return
	}
	if flagLang != "" {
		lang = language.Make(flagLang)
	} else {
		lang = language.Make(os.Getenv("LANG"))
	}
	msgPr = message.NewPrinter(lang)
	root := loadTasks(t)
	doSave := false
	switch {
	case *stop:
		root.WalkAll(nil, CloseAllOpen(t).Do)
		doSave = true
	case *report != "":
		rep := reports[*report](lang, flag.Args())
		rep.Generate(root, t)
	case len(flag.Args()) == 0:
		taskTimesReport{
			tw: txtab.Writer{
				W: os.Stdout,
				F: newTabFormatter(),
			},
			lang: lang,
		}.Generate(root, t)
	default:
		if len(flag.Args()) != 1 {
			log.Fatal("cannot start more than one task")
		}
		pstr := strings.TrimSpace(flag.Arg(0))
		switch {
		case pstr == "" || pstr == "/":
			t = CloseForNext(root, t, root)
			root.Start(t)
			doSave = true
			fmt.Printf("switched to task '%s'\n", pathString(root.Path()))
		case pstr[0] == '/':
			path := strings.Split(pstr, "/")
			task := root.Get(path[1:]...)
			if task == nil {
				log.Fatalf("no task '%s'", pstr)
			}
			t = CloseForNext(root, t, task)
			task.Start(t)
			doSave = true
			fmt.Printf("switched to task '%s'\n", pathString(task.Path()))
		default:
			pat := strings.Split(pstr, "/")
			task, err := findPath(root, pat)
			if err != nil {
				log.Fatal(err)
			}
			if task == nil {
				log.Fatalf("no task matches '%s'", pathString(pat))
			}
			t = CloseForNext(root, t, task)
			task.Start(t)
			doSave = true
			fmt.Printf("switched to task '%s'\n", pathString(task.Path()))
		}
	}
	if doSave {
		saveTasks(dataFile(t), root)
	}
}
