package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/cmd"
	"git.fractalqb.de/fractalqb/tiktak/internal/reports"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
)

var (
	cfg = struct {
		cmd.Config
		Verbose  bool
		NowRound string
		Formats  string
		Report   struct {
			Default string
			Layout  string
		}
		StartOfWeek time.Weekday
	}{
		StartOfWeek: time.Monday, // Corresponds to ISO weeks
	}

	file, info   string
	stop, report bool
	fmts                       = reports.MinutesFmts
	tblwr        tiktbl.Writer = &tiktbl.Terminal{CellPad: "  "}

	now   time.Time
	troot tiktak.Task
	tmln  tiktak.TimeLine
)

func flags() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s v%s:\n", os.Args[0], cmd.Version)
		flag.PrintDefaults()
	}
	fFlag := flag.String("f", "",
		`Set tiktak data file. If file is '-' tiktak reads time data from stdin.
If time data has to be written it is written to stderr.`)
	nowStr := flag.String("t", "", cmd.TimeFlagDoc)
	flag.BoolVar(&stop, "zzz", stop,
		"Stop timing",
	)
	reptFlag := flag.String("r", "",
		`Select report: plain, spans, sums, sheet
Config path: .Report.Default`,
	)
	flag.StringVar(&info, "q", info,
		fmt.Sprintf(`Query infos:
 - dir/d: Print tiktak's data directory. Can be set with environment
          variable %s.
 - file/f: Print current tiktat data file name.
 - match/m [parttern…]: Show known task names from current data file
                        that match given patterns.`,
			cmd.EnvTiktakData),
	)
	flag.StringVar(&cfg.Formats, "formats", cfg.Formats,
		`Select formatters for time display:
 - minute/m: Durations are rounded to minutes and written as hh:mm
             format.
 - second/s: Durations are written with seconds as [_d]__:__'__._"
             Where _d is number of days, __: is hours, __' is minutes
             and __._" is the number of seconds including fractions.
 - c       : Durations are written as hours with fractions.
Config path: .Formats`,
	)
	flag.StringVar(&cfg.Report.Layout, "layout", cfg.Report.Layout,
		`Select report layout: term, csv
Config path: .Report.Layout`,
	)
	flag.BoolVar(&cfg.Verbose, "v", cfg.Verbose, "Request verbose output.")
	flag.Parse()
	if *nowStr != "" {
		now = mustRet(cmd.ParseTime(*nowStr))
	} else {
		now = time.Now()
	}
	if cfg.NowRound != "" {
		round := mustRet(time.ParseDuration(cfg.NowRound))
		now = now.Round(round)
	} else {
		now = now.Round(time.Second)
	}
	if *fFlag == "" {
		file = cfg.DataFile(now)
	} else {
		file = *fFlag
	}
	switch cfg.Formats {
	case "":
	case "m", "minute":
		fmts = reports.MinutesFmts
	case "s", "second":
		fmts = reports.SecondsFmts
	case "c":
		fmts = reports.FracCFmts
	default:
		log.Fatalf("invalid formats: '%s'", cfg.Formats)
	}
	switch cfg.Report.Layout {
	case "":
	case "term":
		tblwr = &tiktbl.Terminal{CellPad: "  "}
	case "csv":
		tblwr = &tiktbl.CSV{FS: ";", SkipEmptyLines: true}
	default:
		log.Fatalf("invalid report layout: '%s'", cfg.Report.Layout)
	}
	if *reptFlag != "" {
		cfg.Report.Default = *reptFlag
		report = true
	}
}

func main() {
	must(cmd.ReadConfig(&cfg))
	flags()
	if info != "" {
		showInfos()
		return
	}
	read()
	if report || len(flag.Args()) == 0 && !stop {
		showReport()
		return
	}
	if l := len(flag.Args()); l > 1 {
		log.Fatal("cannot switch to more than one task")
	} else if l == 0 {
		if stop {
			sum := reports.NewTaskSums(now, cfg.StartOfWeek)
			sum.Of(tmln, nil, fmts)
			tmln.Switch(now, nil)
			write(file)
			log.Printf("Zzz\t%s\n", sumString(sum))
		}
		return
	}
	p := flag.Arg(0)
	must(cmd.CheckPathString(p))
	var t *tiktak.Task
	if path.IsAbs(p) {
		t = troot.GetString(p)
	} else if m := match(&troot, p); len(m) == 0 {
		log.Fatalf("no matching task for '%s'", p)
	} else if len(m) > 1 {
		var sb strings.Builder
		for _, match := range m {
			sb.WriteByte(' ')
			sb.WriteString(match.String())
		}
		log.Fatalf("ambiguous pattern '%s' matches:%s", p, sb.String())
	} else {
		t = m[0]
	}
	sum := reports.NewTaskSums(now, cfg.StartOfWeek)
	sum.Of(tmln, t, fmts)
	tmln.Switch(now, t)
	write(file)
	log.Printf("%s\t%s\n", t, sumString(sum))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustRet[T any](v T, err error) T {
	must(err)
	return v
}

func write(file string) {
	if file == "-" {
		must(tiktak.Write(os.Stdout, tmln))
		return
	}
	tmp := file + "~"
	w := mustRet(os.Create(tmp))
	if err := tiktak.Write(w, tmln); err != nil {
		w.Close()
		must(err)
	}
	must(w.Close())
	must(os.Rename(tmp, file))
}

func read() {
	if file == "-" {
		tmln = mustRet(tiktak.Read(os.Stdin, &troot))
		return
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if !copyTemplate() {
			return
		}
	}
	r := mustRet(os.Open(file))
	defer r.Close()
	tmln = mustRet(tiktak.Read(r, &troot))
}

func copyTemplate() bool {
	tmplFile := cmd.TikTakFile("template.tiktak")
	if _, err := os.Stat(tmplFile); os.IsNotExist(err) {
		return false
	}
	r := mustRet(os.Open(tmplFile))
	defer r.Close()
	w := mustRet(os.Create(file))
	defer w.Close()
	mustRet(io.Copy(w, r))
	return true
}

func sumString(sum *reports.TaskSums) string {
	return fmt.Sprintf("D:%s/%s  W:%s/%s  M:%s/%s",
		sum.Day1, sum.DaySub,
		sum.Week1, sum.WeekSub,
		sum.Month1, sum.MonthSub,
	)
}

func match(t *tiktak.Task, s string) []*tiktak.Task {
	if path.IsAbs(s) {
		return []*tiktak.Task{t.FindString(s)}
	}
	return mustRet(t.MatchString(s))
}

func showReport() {
	switch cfg.Report.Default {
	case "", "plain":
		tiktak.Write(os.Stdout, tmln)
	case "spans":
		r := reports.Spans{Report: reptCfg(), Verbose: cfg.Verbose}
		r.Write(os.Stdout, tmln, now)
	case "sums":
		r := reports.Sums{Report: reptCfg(), WeekStart: cfg.StartOfWeek}
		r.Write(os.Stdout, tmln, now)
	case "sheet":
		r := reports.Sheet{Report: reptCfg(), WeekStart: cfg.StartOfWeek}
		for _, arg := range flag.Args() {
			ts := match(&troot, arg)
			r.Tasks = append(r.Tasks, ts...)
		}
		r.Write(os.Stdout, tmln, now)
	default:
		log.Fatalf("unknown report '%s'", cfg.Report)
	}
}

func showInfos() {
	switch info {
	case "d", "dir":
		fmt.Println(cmd.TikTakDir())
	case "f", "file":
		fmt.Println(file)
	case "m", "match":
		read()
		var tbl tiktbl.Data
		crsr := tbl.At(0, 0).With(reports.Bold()).
			SetStrings("Match", "Task", "Title").
			NextRow()
		for _, arg := range flag.Args() {
			matches := match(&troot, arg)
			if len(matches) == 0 {
				crsr.SetStrings(arg, "-").NextRow()
			} else {
				for _, m := range matches {
					crsr.SetStrings(arg, m.String(), m.Title()).NextRow()
				}
			}
		}
		tblwr.Write(os.Stdout, &tbl)
	default:
		log.Fatalf("invalid info request: '%s'", info)
	}
}

func reptCfg() reports.Report { return reports.Report{Layout: tblwr, Fmts: fmts} }
