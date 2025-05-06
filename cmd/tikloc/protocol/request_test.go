package protocol

import (
	"bytes"
	"testing"

	"git.fractalqb.de/fractalqb/rqre"
	"git.fractalqb.de/fractalqb/testerr"
)

func TestRequest(t *testing.T) {
	msgr := rqre.NewMessenger(rqre.Msg32(1024), rqre.NewBuffers(128))
	key := Key{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
	}
	var net bytes.Buffer
	send := Request{Key: &key}
	testerr.Shall(msgr.Send(&net, send)).BeNil(t)
	data := net.Bytes()

	t.Run("unchanged", func(t *testing.T) {
		recv := Request{Key: &key}
		testerr.Shall(msgr.Recv(&net, &recv)).BeNil(t)
	})

	t.Run("changed", func(t *testing.T) {
		data := bytes.Clone(data)
		data[len(data)/2]++
		net := bytes.NewBuffer(data)
		recv := Request{Key: &key}
		testerr.Shall(msgr.Recv(net, &recv)).NotNil(t)
	})
}
