package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.fractalqb.de/fractalqb/tiktak/cmd"
	"git.fractalqb.de/fractalqb/tiktak/internal/reports"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
	"gopkg.in/yaml.v3"
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
	fNow := flag.String("t", "", cmd.TimeFlagDoc)
	fStop := flag.Bool("zzz", false,
		"Stop timing",
	)
	fRept := flag.String("r", "",
		`Select report: plain, spans, sums, sheet
Config path: .Report.Default`,
	)
	fEdit := flag.Bool("e", false,
		"Edit timeline",
	)
	flag.StringVar(&query, "q", query,
		fmt.Sprintf(`Query infos:
 - dir/d: Print tiktak's data directory. Can be set with environment
          variable %s.
 - file/f: Print current tiktat data file name.
 - match/m [parttern…]: Show known task names from current data file
                        that match given patterns.
 - format: Print example of tiktak file format.`,
			cmd.EnvTiktakData),
	)
	flag.StringVar(&cfg.TikTak.Formats, "formats", cfg.TikTak.Formats,
		`Select formatters for time display:
 - minute/m: Durations are rounded to minutes and written as hh:mm
             format.
 - second/s: Durations are written with seconds as [_d]__:__'__._"
             Where _d is number of days, __: is hours, __' is minutes
             and __._" is the number of seconds including fractions.
 - c       : Durations are written as hours with fractions.
Config path: .Formats`,
	)
	flag.StringVar(&cfg.TikTak.Report.Layout, "layout", cfg.TikTak.Report.Layout,
		`Select report layout: term, csv
Config path: .Report.Layout`,
	)
	flag.StringVar(&cfg.TikTak.Filter, "x", cfg.TikTak.Filter,
		"Select filter to apply before saving or writing reports.",
	)
	fCfgDump := flag.Bool("dump-config", false,
		"Dump config to stdout and exit.",
	)
	flag.BoolVar(&cfg.Verbose, "v", cfg.Verbose, "Request verbose output.")
	flag.Parse()
	if *fCfgDump {
		yaml.NewEncoder(os.Stdout).Encode(&cfg)
		os.Exit(0)
	}

	switch {
	case *fStop:
		mode = StopMode
	case query != "":
		mode = QueryMode
	case *fEdit:
		mode = EditMode
	case *fRept != "":
		mode = ReportMode
	case flag.NArg() == 1:
		mode = SwitchMode
	case flag.NArg() > 1:
		log.Fatal("cannot switch to more than one task")
	}

	now := computeNow(*fNow)

	if *fFlag == "" {
		file = cfg.DataFile(now)
	} else {
		file = *fFlag
	}

	if *fRept != "" {
		cfg.TikTak.Report.Default = *fRept
	}

	switch cfg.TikTak.Formats {
	case "":
	case "m", "minute":
		formats = reports.MinutesFmts
	case "s", "second":
		formats = reports.SecondsFmts
	case "c":
		formats = reports.FracCFmts
	default:
		log.Fatalf("invalid formats: '%s'", cfg.TikTak.Formats)
	}

	switch cfg.TikTak.Report.Layout {
	case "":
	case "term":
		tableWr = &tiktbl.Terminal{CellPad: "  "}
	case "csv":
		tableWr = &tiktbl.CSV{FS: ";", SkipEmptyLines: true}
	default:
		log.Fatalf("invalid report layout: '%s'", cfg.TikTak.Report.Layout)
	}
}

func computeNow(f string) time.Time {
	if f != "" {
		now = mustRet(cmd.ParseTime(f))
	} else {
		now = time.Now()
	}
	if cfg.TikTak.NowRound != "" {
		round := mustRet(time.ParseDuration(cfg.TikTak.NowRound))
		now = now.Round(round)
	} else {
		now = now.Round(time.Second)
	}
	return now
}
