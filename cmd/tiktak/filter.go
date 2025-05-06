package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
)

func runFilters(ls []string) {
	var buf bytes.Buffer
	must(tiktak.Write(&buf, timeline))
	for _, name := range ls {
		fcmd := cfg.TikTak.Filters[name]
		if len(fcmd) == 0 {
			log.Fatalf("unknown filter '%s'", name)
		}
		args := make([]string, len(fcmd)-1)
		for i, a := range fcmd[1:] {
			switch a {
			case "{now}":
				args[i] = now.Format(time.RFC3339)
			default:
				args[i] = a
			}
		}
		var close bool
		cmd := exec.Command(fcmd[0], args...)
		cmd.Stdin = &buf
		cmd.Stderr, close = filterErr(cfg.TikTak.FilterErr)
		if cw, ok := cmd.Stderr.(io.Closer); ok && close {
			defer cw.Close()
		}
		data := mustRet(cmd.Output())
		buf.Reset()
		buf.Write(data)
	}
	timeline = mustRet(tiktak.Read(&buf, &rootTask))
}

func filterErr(fe string) (ferr io.Writer, close bool) {
	switch fe {
	case "":
		return os.Stderr, false
	}
	if fe[0] == '+' {
		w, err := os.OpenFile(fe[1:], os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Println("open filter err:", err)
			return os.Stderr, false
		}
		return w, true
	}
	w, err := os.Create(fe)
	if err != nil {
		log.Println("creare filter err:", err)
		return os.Stderr, false
	}
	return w, true
}
