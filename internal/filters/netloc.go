package filters

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"git.fractalqb.de/fractalqb/tiktak"
	"git.fractalqb.de/fractalqb/tiktak/cmd/tikloc/protocol"
	"github.com/ccding/go-stun/stun"
)

type NetLoc struct {
	checks []netLocCheck
	opts   netLocOpts
}

type netLocOpts struct {
	debug bool
}

func (f *NetLoc) Flags(args []string) (_ []string, err error) {
	flags := flag.NewFlagSet("netloc", flag.ContinueOnError)
	flags.Usage = func() {
		w := flags.Output()
		fmt.Fprintln(w, "netloc checks: http, stun, tikloc")
		flags.PrintDefaults()
	}
	flags.BoolVar(&f.opts.debug, "debug", f.opts.debug, "Log debug messages")
	if err := flags.Parse(args); err != nil {
		return args, err
	}
	var chk netLocCheck
CHECK_LOOP:
	for len(args) > 0 {
		switch args[0] {
		case "http":
			if chk, args, err = newNetLocHTTP(args[1:]); err != nil {
				return args, fmt.Errorf("netloc http: %w", err)
			}
		case "stun":
			if chk, args, err = newNetLocSTUN(args[1:]); err != nil {
				return args, fmt.Errorf("netloc http: %w", err)
			}
		case "tikloc":
			if chk, args, err = newNetLocTik(&f.opts, args[1:]); err != nil {
				return args, fmt.Errorf("netloc tikloc: %w", err)
			}
		default:
			break CHECK_LOOP
		}
		f.checks = append(f.checks, chk)
	}
	if len(f.checks) == 0 {
		return args, errors.New("no netloc filter configuration")
	}
	return args, nil
}

func (f *NetLoc) Filter(tl *tiktak.TimeLine, now time.Time) error {
	l := len(*tl)
	if l == 0 {
		return nil
	}
	_, sw := tl.Pick(now)
	if sw == nil {
		return nil
	}
	locIdx := slices.IndexFunc(sw.Notes(), func(n tiktak.Note) bool {
		return strings.HasPrefix(n.Text, netLocIndicator)
	})
	if locIdx >= 0 {
		return nil
	}
	for _, check := range f.checks {
		if loc := check.Location(); loc != "" {
			sw.AddNote(fmt.Sprintf("%s %s", netLocIndicator, loc))
			break
		}
	}
	return nil
}

const netLocIndicator = "NetLoc:"

type netLocCheck interface{ Location() string }

type netLocHTTPGet struct {
	name    string
	url     string
	find    string
	timeout time.Duration
}

func newNetLocHTTP(args []string) (netLocHTTPGet, []string, error) {
	chk := netLocHTTPGet{timeout: time.Second}
	flags := flag.NewFlagSet("netloc-http", flag.ContinueOnError)
	flags.StringVar(&chk.name, "n", chk.name, "Location name")
	flags.StringVar(&chk.url, "url", chk.url, "HTTP URL")
	flags.StringVar(&chk.url, "find", chk.url, "Find substring")
	flags.DurationVar(&chk.timeout, "timeout", chk.timeout, "Timeout")
	if err := flags.Parse(args); err != nil {
		return chk, args, err
	}
	return chk, flag.Args(), nil
}

func (chk netLocHTTPGet) Location() string {
	htc := http.Client{Timeout: chk.timeout}
	resp, err := htc.Get(chk.url)
	if err != nil {
		log.Printf("netloc http-get '%s': %s", chk.url, err)
		return ""
	}
	defer resp.Body.Close()
	var body bytes.Buffer
	if _, err = io.Copy(&body, io.LimitReader(resp.Body, 1024*1024)); err != nil {
		log.Printf("netloc http-get '%s' read body: %s", chk.url, err)
		return ""
	}
	if bytes.Contains(body.Bytes(), []byte(chk.find)) {
		return chk.name
	}
	return ""
}

type netLocSTUN struct{}

func newNetLocSTUN(args []string) (netLocSTUN, []string, error) {
	return netLocSTUN{}, args, nil
}

func (chk netLocSTUN) Location() string {
	nty, host, err := stun.NewClient().Discover()
	switch {
	case err != nil:
		log.Printf("netloc STUN: %s", err)
		return ""
	case nty <= stun.NATBlocked:
		log.Printf("netloc STUN: %s", nty)
		return ""
	}
	return host.String()
}

type netLocTik struct {
	opts    *netLocOpts
	addr    string
	key     protocol.Key
	nms     map[string]string
	timeout time.Duration
}

func newNetLocTik(opts *netLocOpts, args []string) (netLocTik, []string, error) {
	chk := netLocTik{opts: opts, timeout: 500 * time.Millisecond}
	flags := flag.NewFlagSet("netloc-tikloc", flag.ContinueOnError)
	flags.StringVar(&chk.addr, "addr", chk.addr, "Tikloc address")
	hexkey := flags.String("key", "", "Hex of 32 byte encryption key")
	if err := flags.Parse(args); err != nil {
		return chk, args, err
	}
	if err := chk.key.FromHex(*hexkey); err != nil {
		return chk, args, err
	}
	return chk, flags.Args(), nil
}

func (chk netLocTik) Location() string {
	const logPrefix = "netloc ticloc:"
	srvAddr, err := net.ResolveUDPAddr("udp", chk.addr)
	if err != nil {
		log.Println(logPrefix, "resolve server", err)
		return ""
	}
	conn, err := net.DialUDP("udp", nil, srvAddr)
	if err != nil {
		log.Println(logPrefix, "dial server server", err)
		return ""
	}
	defer conn.Close()
	rq := protocol.Request{Key: &chk.key}
	data := make([]byte, 0, 1024)
	data, err = rq.AppendBinary(data)
	if err != nil {
		log.Println(logPrefix, err)
		return ""
	}
	if _, err = conn.Write(data); err != nil {
		log.Println(logPrefix, err)
		return ""
	}
	if err = conn.SetReadDeadline(time.Now().Add(chk.timeout)); err != nil {
		log.Println("SetReadDeadline:", err)
	}
	n, err := conn.Read(data[:1024])
	if err != nil {
		if chk.opts.debug {
			log.Println(logPrefix, err)
		}
		return ""
	}
	re := protocol.Response{Key: &chk.key}
	if err = re.UnmarshalBinary(data[:n]); err != nil {
		log.Println(logPrefix, err)
		return ""
	}
	for match, nm := range chk.nms {
		rx, err := regexp.Compile(match)
		if err != nil {
			log.Printf(logPrefix+" pattern '%s' %s", match, err)
			continue
		}
		if rx.MatchString(re.Addr) {
			return nm
		}
	}
	return re.Loc
}
