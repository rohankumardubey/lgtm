package model

import (
	"crypto/rand"
	"io"
)

// standard characters allowed in token string.
var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// default token length
var length = 32

// Rand generates a 32-bit random string.
func Rand() string {
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		_, err := io.ReadFull(rand.Reader, r)
		if err != nil {
			return ""
		}
		for _, c := range r {
			if c >= maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
