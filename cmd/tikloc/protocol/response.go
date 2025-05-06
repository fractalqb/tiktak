package protocol

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"unicode/utf8"

	"golang.org/x/crypto/nacl/secretbox"
)

type Response struct {
	Key  *Key
	Loc  string
	Addr string
}

func (re Response) AppendBinary(b []byte) ([]byte, error) {
	if err := re.validate(); err != nil {
		return b, err
	}
	msg, err := appendStr255(nil, re.Loc)
	if err != nil {
		return b, fmt.Errorf("append response location: %w", err)
	}
	if msg, err = appendStr255(msg, re.Addr); err != nil {
		return b, fmt.Errorf("append address location: %w", err)
	}
	var nc nonce
	if _, err := rand.Read(nc[:]); err != nil {
		return b, err
	}
	b = append(b, nc[:]...)
	b = secretbox.Seal(b, msg, &nc, (*[32]byte)(re.Key))
	return b, nil
}

func (re *Response) UnmarshalBinary(data []byte) (err error) {
	var nc nonce
	copy(nc[:], data)
	data, ok := secretbox.Open(nil, data[nonceSize:], &nc, (*[32]byte)(re.Key))
	if !ok {
		return errors.New("failed to decrypt response")
	}
	if len(data) < 1 {
		return errors.New("empty response")
	}
	var n int
	if re.Loc, n, err = str255(data); err != nil {
		return fmt.Errorf("unmarshal response location: %w", err)
	}
	data = data[n:]
	if re.Addr, n, err = str255(data); err != nil {
		return fmt.Errorf("unmarshal response address: %w", err)
	}
	return nil
}

func (re Response) validate() error {
	switch {
	case len(re.Loc) > math.MaxUint8:
		return errors.New("location too long")
	case len(re.Addr) > math.MaxUint8:
		return errors.New("address too long")
	}
	return nil
}

func appendStr255(b []byte, s string) ([]byte, error) {
	l := len(s)
	if l > math.MaxUint8 {
		return b, errors.New("string too long")
	}
	b = append(b, byte(l))
	b = append(b, s...)
	return b, nil
}

func str255(b []byte) (string, int, error) {
	if len(b) == 0 {
		return "", 0, errors.New("empty string data")
	}
	l := int(b[0])
	if l >= len(b) {
		return "", 0, errors.New("string length overflow")
	}
	s := string(b[1 : 1+l])
	if !utf8.ValidString(s) {
		return "", 0, errors.New("invalid UTF8 string encoding")
	}
	return s, l + 1, nil
}
