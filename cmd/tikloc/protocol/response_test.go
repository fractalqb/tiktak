package protocol

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"testing/quick"

	"git.fractalqb.de/fractalqb/rqre"
	"git.fractalqb.de/fractalqb/testerr"
)

func TestResponse(t *testing.T) {
	const loc = "Oz"
	const addr = "hocallost:8080"

	msgr := rqre.NewMessenger(rqre.Msg32(1024), rqre.NewBuffers(128))
	key := Key{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
	}
	var net bytes.Buffer
	send := Response{Key: &key, Loc: loc, Addr: addr}
	testerr.Shall(msgr.Send(&net, send)).BeNil(t)
	data := net.Bytes()

	t.Run("unchanged", func(t *testing.T) {
		recv := Response{Key: &key}
		testerr.Shall(msgr.Recv(&net, &recv)).BeNil(t)
		if recv.Loc != loc {
			t.Errorf("received location '%s'", recv.Loc)
		}
		if recv.Addr != addr {
			t.Errorf("received address '%s'", recv.Addr)
		}
	})

	t.Run("changed", func(t *testing.T) {
		data := bytes.Clone(data)
		data[len(data)/2]++
		net := bytes.NewBuffer(data)
		recv := Response{Key: &key}
		testerr.Shall(msgr.Recv(net, &recv)).NotNil(t)
	})
}

func Test_str255(t *testing.T) {
	t.Run("quick check", func(t *testing.T) {
		count := 0
		f := func(s string) bool {
			if len(s) > math.MaxUint8 {
				return true
			}
			count++
			buf, err := appendStr255(nil, s)
			if err != nil {
				t.Log(err)
				return false
			}
			r, n, err := str255(buf)
			if r != s {
				t.Logf("[%s] =/= [%s]", r, s)
				return false
			}
			if n != len(r)+1 {
				t.Logf("n=%d, expect %d", n, len(r)+1)
				return false
			}
			return true
		}
		testerr.Shall(quick.Check(f, nil)).BeNil(t)
		if count == 0 {
			t.Fatal("no significant quick checks")
		}
	})
	t.Run("too long", func(t *testing.T) {
		testerr.Shall1(appendStr255(nil, strings.Repeat("x", math.MaxUint8+1))).
			NotNil(t)
	})
	t.Run("overflow", func(t *testing.T) {
		buf := testerr.Shall1(appendStr255(nil, strings.Repeat("x", 20))).BeNil(t)
		_, _, err := str255(buf[:len(buf)-1])
		testerr.Shall(err).NotNil(t)
	})
}
