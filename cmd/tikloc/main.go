package main

import (
	"flag"
	"log"
	"net"
	"time"

	"git.fractalqb.de/fractalqb/tiktak/cmd/tikloc/protocol"
	"git.fractalqb.de/fractalqb/yacfg"
	"git.fractalqb.de/fractalqb/yacfg/yasec"
	"golang.org/x/time/rate"
)

const yaecSalt = "165420032114010003102131116136054556020611050211140565071762"

var (
	cfg = struct {
		Addr   string
		Loc    string
		KeyHex yasec.Secret
	}{
		Addr: ":7171",
	}

	key protocol.Key
)

func flags() {
	flag.StringVar(&cfg.Loc, "loc", cfg.Loc, "Fix location name")
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "Server listen address")
	noTerm := flag.Bool("no-term", false, "Do not read yasec passphrase from terminal")
	flag.Parse()

	yasecInit(*noTerm)
	lkey := mustRet(cfg.KeyHex.Open())
	defer lkey.Destroy()
	must(key.FromHex(lkey.String()))
}

func main() {
	mustRet(yacfg.FromEnvThenFiles{
		EnvPrefix:     "TICLOC_",
		FilesFlagName: "cfg",
		Flags:         flag.CommandLine,
		Log:           func(m string) { log.Print(m) },
	}.Configure(&cfg))
	flags()

	limiter := rate.NewLimiter(rate.Every(time.Second), 5)
	lstn := mustRet(net.ListenPacket("udp", cfg.Addr))
	defer lstn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, err := lstn.ReadFrom(buf)
		must(err)
		if limiter.Allow() {
			netloc(lstn, addr, buf[:n])
		} else {
			log.Println("ratelimit rejects", addr)
		}
	}
}

func netloc(conn net.PacketConn, client net.Addr, data []byte) {
	var err error
	rq := protocol.Request{Key: &key}
	if err = rq.UnmarshalBinary(data); err != nil {
		log.Printf("message from %s: %s", client, err)
		return
	}
	re := protocol.Response{Key: &key, Loc: cfg.Loc, Addr: client.String()}
	if data, err = re.AppendBinary(data[:0]); err != nil {
		log.Printf("encoding response to %s: %s", re.Addr, err)
		return
	}
	if err = conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		log.Println("SetWriteDeadline:", err)
	}
	if _, err = conn.WriteTo(data, client); err != nil {
		log.Printf("writing response to %s: %s", re.Addr, err)
	}
}
