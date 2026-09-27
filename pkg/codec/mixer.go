package codec

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Mixer performs a bijective map: mixed = (seq*A + B) mod M.
type Mixer struct {
	M uint64
	A uint64
	B uint64
}

// NewMixerFromSalt derives A (coprime to M) and B from salt.
func NewMixerFromSalt(salt string, m uint64) (*Mixer, error) {
	if m < 2 {
		return nil, fmt.Errorf("modulus too small")
	}
	sum := sha256.Sum256([]byte(salt))
	a := binary.BigEndian.Uint64(sum[0:8]) % m
	if a%2 == 0 {
		a++
	}
	for gcd(a, m) != 1 {
		a += 2
		if a >= m {
			a = 1
		}
	}
	b := binary.BigEndian.Uint64(sum[8:16]) % m
	return &Mixer{M: m, A: a, B: b}, nil
}

// Mix maps seq into [0, M).
func (m *Mixer) Mix(seq uint64) uint64 {
	return (mulMod(seq, m.A, m.M) + m.B) % m.M
}

// Unmix inverts Mix (for tests / debugging).
func (m *Mixer) Unmix(mixed uint64) uint64 {
	inv := modInverse(m.A, m.M)
	x := (mixed + m.M - (m.B % m.M)) % m.M
	return mulMod(x, inv, m.M)
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func mulMod(a, b, mod uint64) uint64 {
	var result uint64
	a %= mod
	for b > 0 {
		if b&1 == 1 {
			result = (result + a) % mod
		}
		a = (a * 2) % mod
		b >>= 1
	}
	return result
}

// modInverse via extended Euclid; a and mod must be coprime.
func modInverse(a, mod uint64) uint64 {
	var t, newT int64 = 0, 1
	var r, newR int64 = int64(mod), int64(a % mod)
	for newR != 0 {
		q := r / newR
		t, newT = newT, t-q*newT
		r, newR = newR, r-q*newR
	}
	if r > 1 {
		return 0
	}
	if t < 0 {
		t += int64(mod)
	}
	return uint64(t)
}

// ModulusForLength returns 62^length.
func ModulusForLength(length int) uint64 {
	var m uint64 = 1
	for i := 0; i < length; i++ {
		m *= 62
	}
	return m
}
