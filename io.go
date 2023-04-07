package tiktak

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/mod/semver"
)

const (
	FileVersion = "1.0.0"
	IOTimeFmt   = time.RFC3339
)

func Write(w io.Writer, tl TimeLine) error {
	fmt.Fprintf(w, "v%s\ttiktak time tracker\n", FileVersion)
	if root := tl.FirstTask().Root(); root != nil {
		var wrTasks func(*Task)
		wrTasks = func(t *Task) {
			if len(t.subs) == 0 || t.Title() != "" {
				if t.Title() != "" {
					fmt.Fprintln(w, t.String(), t.Title())
				} else {
					fmt.Fprintln(w, t.String())
				}
			}
			for _, s := range t.subs {
				wrTasks(s)
			}
		}
		wrTasks(root)
	}
	if len(tl) == 0 {
		return nil
	}
	var day Date
	for _, s := range tl {
		sday := DateOf(s.When())
		if sday.Compare(&day) != 0 {
			fmt.Fprintf(w, "# %s\n", s.When().Format("Mon, 02 Jan 2006"))
			day = sday
		}
		if t := s.Task(); t == nil {
			fmt.Fprintf(w, "%s\n", s.When().Format(IOTimeFmt))
		} else {
			fmt.Fprintf(w, "%s %s\n", s.When().Format(IOTimeFmt), t)
		}
		for _, note := range s.notes {
			if note.Sym == 0 {
				fmt.Fprintf(w, "\t. %s\n", note.Text)
			} else {
				fmt.Fprintf(w, "\t!%c %s\n", note.Sym, note.Text)
			}
		}
	}
	return nil
}

var majorFileVersion = semver.Major("v" + FileVersion)

func Read(r io.Reader, root *Task) (tl TimeLine, err error) {
	if root == nil {
		root = new(Task)
	}
	lno := 0
	scn := bufio.NewScanner(r)
	lastSwitch := -1
	for scn.Scan() {
		lno++
		line := scn.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		switch line[0] {
		case '/':
			sep := strings.IndexAny(line, " \t")
			if sep < 0 {
				root.GetString(line)
			} else {
				t := root.GetString(line[:sep])
				t.SetTitle(strings.TrimSpace(line[sep:]))
			}
		case 'v':
			sep := strings.IndexAny(line, " \t")
			if sep > 0 {
				line = line[:sep]
			}
			if !semver.IsValid(line) {
				return nil, fmt.Errorf("%d:syntax error in file version '%s'",
					lno,
					line,
				)
			}
			major := semver.Major(line)
			if major != majorFileVersion {
				return nil, fmt.Errorf("%d: incompatible file version %s, current v%s",
					lno,
					line,
					FileVersion,
				)
			}
			if semver.Compare(line, "v"+FileVersion) > 0 {
				log.Printf("%d: file version %s greater than current v%s",
					lno,
					line,
					FileVersion,
				)
			}
		default:
			if strings.IndexAny(line, " \t") == 0 {
				if lastSwitch < 0 {
					return nil, fmt.Errorf("%d:note bfore first switch", lno)
				}
				n, err := parseNote(line)
				if err != nil {
					return nil, fmt.Errorf("%d:%w", lno, err)
				}
				tl[lastSwitch].notes = append(tl[lastSwitch].notes, n)
			} else {
				fs := strings.Split(line, " ")
				t, err := time.Parse(IOTimeFmt, fs[0])
				if err != nil {
					return nil, fmt.Errorf("%d:%w", lno, err)
				}
				if len(fs) == 1 {
					lastSwitch = tl.Switch(t, nil)
					continue
				}
				switch {
				case len(fs[1]) == 0:
					return nil, fmt.Errorf("%d:empty task path", lno)
				case fs[1][0] != '/':
					return nil, fmt.Errorf("%d:not an absolute path '%s'", lno, fs[1])
				}
				task := root.GetString(fs[1])
				lastSwitch = tl.Switch(t, task)
			}
		}
	}
	return tl, nil
}

func parseNote(line string) (Note, error) {
	line = strings.TrimSpace(line)
	switch line[0] {
	case '.':
		if len(line) < 3 {
			return Note{}, errors.New("empty note")
		}
		return Note{Text: line[2:]}, nil
	case '!':
		if len(line) < 4 {
			return Note{}, errors.New("empty warning")
		}
		sym, sz := utf8.DecodeRuneInString(line[1:])
		return Note{Sym: sym, Text: line[2+sz:]}, nil
	}
	return Note{}, fmt.Errorf("invalid note type '%c'", line[0])
}
