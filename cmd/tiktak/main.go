package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/cmd"
	"git.fractalqb.de/fractalqb/tiktak/internal/reports"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
)

type Config struct {
	NowRound string
	Formats  string
	Report   struct {
		Default string
		Layout  string
	}
	StartOfWeek time.Weekday
	Filters     map[string][]string
	Filter      string
}

type cmdMode int

const (
	ReportMode cmdMode = iota
	StopMode
	EditMode
	QueryMode
	SwitchMode
)

var (
	cfg = struct {
		cmd.Config
		Verbose bool
		TikTak  Config
	}{
		TikTak: Config{
			StartOfWeek: time.Monday, // Corresponds to ISO weeks
		},
	}

	mode        = ReportMode
	file, query string
	formats                   = reports.MinutesFmts
	tableWr     tiktbl.Writer = &tiktbl.Terminal{CellPad: "  "}

	now      time.Time
	rootTask tiktak.Task
	timeline tiktak.TimeLine

	//go:embed format.txt
	formatMsg string
)

func main() {
	must(cmd.ReadConfig(&cfg))
	flags()

	switch mode {
	case ReportMode:
		read()
		showReport()
	case StopMode:
		read()
		sum := reports.NewTaskSums(now, cfg.TikTak.StartOfWeek)
		sum.Of(timeline, nil, formats)
		timeline.Switch(now, nil)
		write(file)
		log.Printf("Zzz\t%s\n", sumString(sum))
	case SwitchMode:
		read()
		p := flag.Arg(0)
		must(cmd.CheckPathString(p))
		var t *tiktak.Task
		if path.IsAbs(p) {
			var err error
			if t, err = rootTask.GetString(p); err != nil {
				log.Fatal(err)
			}
		} else if m := match(&rootTask, p); len(m) == 0 {
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
		sum := reports.NewTaskSums(now, cfg.TikTak.StartOfWeek)
		sum.Of(timeline, t, formats)
		timeline.Switch(now, t)
		write(file)
		log.Printf("%s\t%s\n", t, sumString(sum))
	case EditMode:
		read()
		edit(flag.Args())
		showReport() // TODO switch to write
	case QueryMode:
		showInfos()
	}
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
	runFilter(cfg.TikTak.Filter)
	if file == "-" {
		must(tiktak.Write(os.Stdout, timeline))
		return
	}
	tmp := file + "~"
	w := mustRet(os.Create(tmp))
	if err := tiktak.Write(w, timeline); err != nil {
		w.Close()
		must(err)
	}
	must(w.Close())
	must(os.Rename(tmp, file))
}

func read() {
	if file == "-" {
		timeline = mustRet(tiktak.Read(os.Stdin, &rootTask))
		return
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if !copyTemplate() {
			return
		}
	}
	r := mustRet(os.Open(file))
	defer r.Close()
	timeline = mustRet(tiktak.Read(r, &rootTask))
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
		if ft := t.FindString(s); ft != nil {
			return []*tiktak.Task{ft}
		}
		return nil
	}
	return mustRet(t.MatchString(s))
}

func showReport() {
	runFilter(cfg.TikTak.Filter)
	switch cfg.TikTak.Report.Default {
	case "", "plain":
		tiktak.Write(os.Stdout, timeline)
	case "spans":
		r := reports.Spans{Report: reptCfg(), Verbose: cfg.Verbose}
		r.Write(os.Stdout, timeline, now)
	case "sums":
		r := reports.Sums{Report: reptCfg(), WeekStart: cfg.TikTak.StartOfWeek}
		r.Write(os.Stdout, timeline, now)
	case "sheet":
		r := reports.Sheet{Report: reptCfg(), WeekStart: cfg.TikTak.StartOfWeek}
		for _, arg := range flag.Args() {
			ts := match(&rootTask, arg)
			r.Tasks = append(r.Tasks, ts...)
		}
		r.Write(os.Stdout, timeline, now)
	default:
		log.Fatalf("unknown report '%s'", cfg.TikTak.Report)
	}
}

func showInfos() {
	switch query {
	case "d", "dir":
		fmt.Println(cmd.TikTakDir())
	case "f", "file":
		fmt.Println(file)
	case "m", "match":
		read()
		var tbl tiktbl.Data
		if cfg.Verbose {
			crsr := tbl.At(0, 0).With(reports.Bold()).
				SetStrings("Match", "Task", "Title").
				NextRow()
			for _, arg := range flag.Args() {
				matches := match(&rootTask, arg)
				if len(matches) == 0 {
					crsr.SetStrings(arg, "-").NextRow()
				} else {
					for _, m := range matches {
						crsr.SetStrings(arg, m.String(), m.Title()).NextRow()
					}
				}
			}
			tableWr.Write(os.Stdout, &tbl)
		} else {
			seen := make(map[*tiktak.Task]bool)
			for _, arg := range flag.Args() {
				matches := match(&rootTask, arg)
				for _, t := range matches {
					if !seen[t] {
						seen[t] = true
						fmt.Println(t.String())
					}
				}
			}
		}
	case "format":
		fmt.Print(formatMsg)
	default:
		log.Fatalf("invalid info request: '%s'", query)
	}
}

func reptCfg() reports.Report { return reports.Report{Layout: tableWr, Fmts: formats} }

func runFilter(name string) {
	if name == "" {
		return
	}
	cmdArgs := cfg.TikTak.Filters[name]
	if len(cmdArgs) == 0 {
		log.Fatalf("unknown filter '%s'", name)
	}
	var buf bytes.Buffer
	must(tiktak.Write(&buf, timeline))
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = &buf
	cmd.Stderr = os.Stderr
	data := mustRet(cmd.Output())
	timeline = mustRet(tiktak.Read(bytes.NewReader(data), &rootTask))
}
