package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/internal/reports"
)

const help = `Edit commands refer to task switch events by switch ID.
You can find switch IDs in the first column of the output of
'tiktak -r spans -v'.

tiktak edit commands:
- help           : Show tiktak edit help
- delete (del, d): delete task switch
- move (mv)      : move task's span to new position
`

func edit(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, help)
		log.Fatal("missing edit command")
	}
	switch args[0] {
	case "help":
		fmt.Fprint(os.Stderr, help)
		os.Exit(0)
	case "delete", "del", "d":
		edDelete(args)
	case "move", "mv":
		edMove(args)
	default:
		log.Fatalf("invalid edit command '%s'", args[0])
	}
}

func edDelete(args []string) {
	flags := flag.NewFlagSet("delete", flag.ExitOnError)
	flags.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s delete: <switch-id>...\n", os.Args[0])
		flags.PrintDefaults()

	}
	flags.Parse(args[1:])
	for _, sid := range flags.Args() {
		idx := mustRet(reports.ParseSpanID(sid))
		if err := timeline.DelSwitch(idx); err != nil {
			log.Fatalf("switch ID '%s': %s", sid, err)
		}
	}
}

func edMove(args []string) {
	flags := flag.NewFlagSet("move", flag.ExitOnError)
	flags.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s move: <source-id> <destination-id>\n", os.Args[0])
		flags.PrintDefaults()

	}
	flags.Parse(args[1:])
	if flags.NArg() != 2 {
		flags.Usage()
		log.Fatal("invalid arguments")
	}
	sIdx := mustRet(reports.ParseSpanID(args[1]))
	if sIdx < 0 || sIdx >= len(timeline) {
		log.Fatalf("source switch %d (%s) out of 0..%d range", sIdx, args[1], len(timeline))
	}
	dIdx := mustRet(reports.ParseSpanID(args[2]))
	if dIdx < 0 || dIdx >= len(timeline) {
		log.Fatalf("destination switch %d (%s) out of 0..%d range", dIdx, args[2], len(timeline))
	}
	if sIdx == dIdx {
		return
	}
	dur := timeline[sIdx].Duration()
	if dur < 0 {
		log.Fatalf("cannot move open switch %d (%s) at %s",
			sIdx,
			args[1],
			timeline[sIdx].When(),
		)
	}
	sSw, dSw := timeline[sIdx], timeline[dIdx]
	if dIdx < sIdx {
		tt := dSw.When().Add(dur)
		// log.Printf("moving task to %s < %s\n", tt, sSw.When())
		timeline.Del(sSw.When(), tiktak.AllSwitch, nil)
		timeline.Insert(tt, sSw.Task(), -dur, tiktak.AllSwitch, 0, nil)
	} else if dSw.Duration() < 0 {
		log.Fatal("cannot move forward to open destination task")
	} else {
		tt := dSw.Next().When().Add(-dur)
		// log.Printf("moving task at %s > %s\n", timeline[sIdx].When(), tt)
		timeline.Del(timeline[sIdx].When(), nil, tiktak.AllSwitch)
		timeline.Insert(tt, sSw.Task(), 0, nil, dur, tiktak.AllSwitch)
	}
}
