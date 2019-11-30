package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func usage() {
	wr := flag.CommandLine.Output()
	fmt.Fprintf(wr, "Usage of %s (v%s):\n", os.Args[0], version)
	flag.PrintDefaults()
}

const (
	flagDocCsv = `set separator and use CSV output`

	flagDocFile = `explicitly choose data file.
When not explicitly selected tiktak will look in the directory given
in the ` + envDDir + ` environment variable.`

	flagDocLang = `select language`

	flagDocUgap = `length of µ-gap`

	flagDocDurFmt = `select format for durations: m, f, c
 - m: show hours and minutes
 - f: show hours and 2-digit fraction with a locale specific separator
 - c: show hours and 2-digit fraction with '.' separator (C locale)
`

	flagDocDateFmt = `set format for date output`

	flagDocZzz = `stop all running clocks`

	flagDocPFile = `print data file name`

	flagDocNow = `Set current time for operation. Missing elements are taken from now().
Formats:
 - HH:MM            : local wall clock
 - {dwmy}-n[Thh:mm] : go back n days, weeks, months or years
                      and optionally set wall clock time
 - mm/yyyy          : month and year
 - yyyy-mm-ddTHH:MM : local wall clock time on a specific date`
)

func flagDocReport() string {
	return `select report: ` + strings.Join(reportNames(), ", ")
}
