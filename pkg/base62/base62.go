package base62

import (
	"fmt"
	"strings"
)

const DefaultCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Encoder encodes/decodes integers with a configurable Base62 alphabet.
type Encoder struct {
	chars  string
	index  [256]int
	base   uint64
	padLen int
}

// NewEncoder builds an encoder. charset must be exactly 62 unique runes; padLen is fixed output length.
func NewEncoder(charset string, padLen int) (*Encoder, error) {
	if charset == "" {
		charset = DefaultCharset
	}
	if len(charset) != 62 {
		return nil, fmt.Errorf("base62 charset must be 62 chars, got %d", len(charset))
	}
	seen := make(map[byte]struct{}, 62)
	e := &Encoder{chars: charset, base: 62, padLen: padLen}
	for i := range e.index {
		e.index[i] = -1
	}
	for i := 0; i < 62; i++ {
		c := charset[i]
		if _, ok := seen[c]; ok {
			return nil, fmt.Errorf("duplicate char in base62 charset: %c", c)
		}
		seen[c] = struct{}{}
		e.index[c] = i
	}
	if padLen < 1 {
		return nil, fmt.Errorf("padLen must be >= 1")
	}
	return e, nil
}

// EncodeFixed encodes n into a fixed-length Base62 string (left-padded with chars[0]).
func (e *Encoder) EncodeFixed(n uint64) string {
	if n == 0 {
		return strings.Repeat(string(e.chars[0]), e.padLen)
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = e.chars[n%e.base]
		n /= e.base
	}
	s := string(buf[i:])
	if len(s) >= e.padLen {
		// Truncate from the left if somehow longer (should not happen for n < 62^padLen)
		return s[len(s)-e.padLen:]
	}
	return strings.Repeat(string(e.chars[0]), e.padLen-len(s)) + s
}

// Decode parses a Base62 string using this encoder's charset.
func (e *Encoder) Decode(s string) (uint64, error) {
	var n uint64
	for i := 0; i < len(s); i++ {
		idx := e.index[s[i]]
		if idx < 0 {
			return 0, fmt.Errorf("invalid base62 char: %c", s[i])
		}
		n = n*e.base + uint64(idx)
	}
	return n, nil
}

// Int2String keeps backward-compatible helper using default charset (variable length).
func Int2String(seq uint64) string {
	e, _ := NewEncoder(DefaultCharset, 1)
	if seq == 0 {
		return string(DefaultCharset[0])
	}
	var buf []byte
	for seq > 0 {
		buf = append(buf, DefaultCharset[seq%62])
		seq /= 62
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	_ = e
	return string(buf)
}
