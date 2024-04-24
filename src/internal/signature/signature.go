package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

var signature = os.Getenv("GTUN_SIGNATURE")
var ErrInvalidSignature = errors.New("invalid signature")
var ErrEmptyConfig = errors.New("empty config")

const signPrefix = "sign="

func Sign(buf []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(append(buf, []byte(signature)...))
	sum := h.Sum(nil)
	calcSignature := hex.EncodeToString(sum)

	return []byte("sign=" + calcSignature + "\n" + string(buf)), nil
}

func UnSign(buf []byte) ([]byte, error) {
	if signature == "" {
		return buf, nil
	}
	content := string(buf)

	lines := strings.Split(content, "\n")
	if !strings.HasPrefix(lines[0], signPrefix) {
		return nil, fmt.Errorf("signature error")
	}

	sign := strings.Split(lines[0], signPrefix)[1]

	bodyOffset := len(lines[0] + "\n")
	if bodyOffset > len(buf) {
		return nil, ErrEmptyConfig
	}

	body := string(buf[bodyOffset:])

	h := sha256.New()
	h.Write([]byte(body + signature))
	sum := h.Sum(nil)
	calcSignature := hex.EncodeToString(sum)

	if calcSignature != sign {
		return nil, ErrInvalidSignature
	}

	return []byte(body), nil
}

func SetSignature(sig string) {
	signature = sig
}
