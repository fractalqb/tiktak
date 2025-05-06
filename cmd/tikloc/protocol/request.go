package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/nacl/secretbox"
)

type Key [32]byte

func (k *Key) FromHex(hexKey string) error {
	bin, err := hex.DecodeString(hexKey)
	if err != nil {
		return fmt.Errorf("decode hex key: %w", err)
	}
	if l := len(bin); l != 32 {
		return fmt.Errorf("hex key with %d byte, need 32", l)
	}
	copy((*k)[:], bin)
	return nil
}

type Request struct {
	Key *Key
}

func (rq Request) AppendBinary(b []byte) ([]byte, error) {
	var nc nonce
	if _, err := rand.Read(nc[:]); err != nil {
		return b, err
	}
	b = append(b, nc[:]...)
	pl, err := time.Now().MarshalBinary()
	if err != nil {
		return b, err
	}
	b = secretbox.Seal(b, pl, &nc, (*[32]byte)(rq.Key))
	return b, nil
}

func (rq Request) UnmarshalBinary(data []byte) error {
	if len(data) < nonceSize {
		return errors.New("invalid message")
	}
	var nc nonce
	copy(nc[:], data)
	_, ok := secretbox.Open(nil, data[nonceSize:], &nc, (*[32]byte)(rq.Key))
	if !ok {
		return errors.New("failed to decrypt request")
	}
	return nil
}

const nonceSize = 24

type nonce = [nonceSize]byte
