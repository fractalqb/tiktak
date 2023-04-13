package reports

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/tiktbl"
)

type Spans struct {
	Report
	Verbose bool
}

func (spans *Spans) Write(w io.Writer, tl tiktak.TimeLine, now time.Time) {
	fmts := spans.Fmts
	if fmts == nil {
		fmts = MinutesFmts
	}
	today := tiktak.DateOf(now)
	var (
		tbl tiktbl.Data
		day tiktak.Date
	)
	crsr := tbl.At(0, 0)
	for i, s := range tl {
		if s.Task() == nil && s.Next() == nil {
			continue
		}
		sday := tiktak.DateOf(s.When())
		if sday.Compare(&day) != 0 {
			style := Underline()
			if sday.Compare(&today) == 0 {
				style = tiktbl.Styles{Bold(), Underline()}
			}
			_, week := s.When().ISOWeek()
			d := fmt.Sprintf("%s; Week %d", fmts.Date(s.When()), week)
			crsr.SetString(d, tiktbl.SpanAll, style).NextRow()
			day = sday
		}
		end := "..."
		style := Bold()
		var dur string
		if ns := s.Next(); ns == nil {
			dur = fmts.Duration(now.Sub(s.When()))
		} else {
			end = fmts.Clock(ns.When())
			dur = fmts.Duration(ns.When().Sub(s.When()))
			style = tiktbl.NoStyle()
		}
		if s.Task() == nil {
			style = tiktbl.AddStyles(style, Muted())
		}
		var flags []rune
		for _, note := range s.Notes() {
			if note.Sym == 0 {
				continue
			}
			if !strings.ContainsRune(string(flags), note.Sym) {
				flags = append(flags, note.Sym)
			}
		}
		if len(flags) > 0 {
			sort.Slice(flags, func(i, j int) bool { return flags[i] < flags[j] })
			style = tiktbl.AddStyles(style, Warn())
		}
		if spans.Verbose {
			crsr.SetString(SpanID(i), tiktbl.Right)
		}
		crsr = crsr.With(style).SetStrings(
			string(flags),
			fmts.Clock(s.When()),
			end,
			dur,
		)
		if s.Task() != nil {
			crsr = crsr.SetString(s.Task().String(), style)
		}
		crsr = crsr.NextRow()
		if spans.Verbose {
			for _, note := range s.Notes() {
				crsr.SetString("")
				if note.Sym == 0 {
					crsr.SetString(note.Text, tiktbl.SpanAll, Underline())
				} else {
					crsr.SetString(fmt.Sprintf("%c %s", note.Sym, note.Text), tiktbl.SpanAll, Underline())
				}
				crsr.NextRow()
			}
		}
	}
	tbl.Align(tiktbl.Left, 0)
	tbl.Align(tiktbl.Right, 3)
	spans.Layout.Write(w, &tbl)
}
