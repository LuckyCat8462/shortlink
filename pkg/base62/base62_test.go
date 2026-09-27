package base62

import "testing"

func TestEncodeFixedLength(t *testing.T) {
	e, err := NewEncoder(DefaultCharset, 6)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []uint64{0, 1, 61, 62, 1000, 62*62*62 + 5} {
		s := e.EncodeFixed(n)
		if len(s) != 6 {
			t.Fatalf("EncodeFixed(%d)=%q len=%d", n, s, len(s))
		}
		got, err := e.Decode(s)
		if err != nil {
			t.Fatal(err)
		}
		if got != n {
			t.Fatalf("roundtrip %d -> %q -> %d", n, s, got)
		}
	}
}

func TestShuffledCharset(t *testing.T) {
	shuffled := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	e, err := NewEncoder(shuffled, 6)
	if err != nil {
		t.Fatal(err)
	}
	s := e.EncodeFixed(10)
	if len(s) != 6 {
		t.Fatal(s)
	}
	n, err := e.Decode(s)
	if err != nil || n != 10 {
		t.Fatalf("got %d err=%v", n, err)
	}
}

func TestInt2StringCompat(t *testing.T) {
	if Int2String(0) != "0" {
		t.Fatal(Int2String(0))
	}
	if Int2String(62) != "10" {
		t.Fatal(Int2String(62))
	}
}
