package stats

import (
	"testing"
	"time"
)

func TestDayRangeUTC_Shanghai(t *testing.T) {
	day, err := ParseDate("2026-10-10")
	if err != nil {
		t.Fatal(err)
	}
	start, end := DayRangeUTC(day)
	// 2026-10-10 00:00 CST = 2026-10-09 16:00 UTC
	if got := start.UTC().Format(time.RFC3339); got != "2026-10-09T16:00:00Z" {
		t.Fatalf("start=%s", got)
	}
	if got := end.UTC().Format(time.RFC3339); got != "2026-10-10T16:00:00Z" {
		t.Fatalf("end=%s", got)
	}
}

func TestYesterdayShanghai(t *testing.T) {
	// 2026-10-11 01:30 UTC = 2026-10-11 09:30 Shanghai → yesterday 2026-10-10
	now := time.Date(2026, 10, 11, 1, 30, 0, 0, time.UTC)
	y := YesterdayShanghai(now)
	if got := FormatDate(y); got != "2026-10-10" {
		t.Fatalf("yesterday=%s", got)
	}
}

func TestLastNDaysRangeUTC(t *testing.T) {
	now := time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC) // Shanghai 11:00 on 10-11
	start, end, days := LastNDaysRangeUTC(now, 7)
	if len(days) != 7 {
		t.Fatalf("len=%d", len(days))
	}
	if FormatDate(days[0]) != "2026-10-05" || FormatDate(days[6]) != "2026-10-11" {
		t.Fatalf("days=%s..%s", FormatDate(days[0]), FormatDate(days[6]))
	}
	if start.UTC().Format(time.RFC3339) != "2026-10-04T16:00:00Z" {
		t.Fatalf("start=%s", start.UTC().Format(time.RFC3339))
	}
	if end.UTC().Format(time.RFC3339) != "2026-10-11T16:00:00Z" {
		t.Fatalf("end=%s", end.UTC().Format(time.RFC3339))
	}
}

func TestParseDateInvalid(t *testing.T) {
	if _, err := ParseDate("10/11/2026"); err == nil {
		t.Fatal("expected error")
	}
}
