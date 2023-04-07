package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"
)

var relTimeRegexp = regexp.MustCompile(`^([ymwdHM])-(\d+)(?:T(\d\d:\d\d))?$`)

func ParseRelativeTime(rt string) (time.Time, bool, error) {
	match := relTimeRegexp.FindStringSubmatch(rt)
	if match == nil {
		return time.Time{}, false, nil
	}
	diff, err := strconv.Atoi(match[2])
	if err != nil {
		return time.Time{}, false, err
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
	ho, mi, se := res.Clock()
	switch match[1] {
	case "M":
		mi -= diff
	case "H":
		ho -= diff
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
		ho, mi, se, res.Nanosecond(),
		res.Location())
	return res, true, nil
}

func ParseTime(tstr string) (time.Time, error) {
	if tstr == "" {
		return time.Now().Round(time.Second), nil
	}
	if res, ok, err := ParseRelativeTime(tstr); err != nil {
		return time.Time{}, err
	} else if ok {
		return res, nil
	}
	t, err := time.ParseInLocation("2006-01-02T15:04", tstr, time.Local)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("15:04", tstr, time.Local)
	if err == nil {
		n := time.Now()
		return time.Date(
			n.Year(), n.Month(), n.Day(),
			t.Hour(), t.Minute(), n.Second(),
			0, n.Location(),
		), nil
	}
	t, err = time.ParseInLocation("01/2006", tstr, time.Local)
	if err != nil {
		log.Fatal(err)
	}
	n := time.Now()
	return time.Date(
		t.Year(), t.Month(),
		n.Day(), n.Hour(), n.Minute(), n.Second(),
		0, n.Location(),
	), nil
}

var durRegexp = regexp.MustCompile(`^(\d+)([smh]?)$`)

func ParseDuration(s string) (time.Duration, error) {
	match := durRegexp.FindStringSubmatch(s)
	if match == nil {
		return 0, fmt.Errorf("invalid duration: '%s'", s)
	}
	res := time.Minute
	switch match[2] {
	case "s":
		res = time.Second
	case "h":
		res = time.Hour
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * res, nil
}

const TimeFlagDoc = `Set current time for operation. Missing elements are taken from now.
Formats:
 - HH:MM              : Local wall clock.
 - {MHdwmy}-n[Thh:mm] : Go back n days, weeks, months or years
                        and optionally set wall clock time.
 - mm/yyyy            : Month and year.
 - yyyy-mm-ddTHH:MM   : Local wall clock time on a specific date.`
