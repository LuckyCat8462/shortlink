package urltool

import "testing"

func TestNormalizeLongURL(t *testing.T) {
	u, err := NormalizeLongURL("  https://example.com/a?x=1  ")
	if err != nil || u != "https://example.com/a?x=1" {
		t.Fatalf("got %q err=%v", u, err)
	}
	if _, err := NormalizeLongURL("ftp://x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestURLHashStable(t *testing.T) {
	h1 := URLHash("", "https://example.com")
	h2 := URLHash("", "https://example.com")
	if h1 != h2 || len(h1) != 32 {
		t.Fatalf("%s %s", h1, h2)
	}
	h3 := URLHash("camp", "https://example.com")
	if h1 == h3 {
		t.Fatal("scene should change hash")
	}
}

func TestGetBasePath(t *testing.T) {
	p, err := GetBasePath("https://s.example.com/aB3xY2")
	if err != nil || p != "aB3xY2" {
		t.Fatalf("%q %v", p, err)
	}
}
