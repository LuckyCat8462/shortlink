package stats

import (
	"fmt"
	"time"
)

const LocationName = "Asia/Shanghai"

// Location returns Asia/Shanghai, falling back to fixed UTC+8 if tzdata is missing.
func Location() *time.Location {
	loc, err := time.LoadLocation(LocationName)
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

// ParseDate parses YYYY-MM-DD as a Shanghai calendar date (midnight local).
func ParseDate(s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", s, Location())
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, want YYYY-MM-DD", s)
	}
	return t, nil
}

// DayRangeUTC returns [start, end) in UTC for the Shanghai calendar day of dayLocal.
func DayRangeUTC(dayLocal time.Time) (startUTC, endUTC time.Time) {
	loc := Location()
	y, m, d := dayLocal.In(loc).Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)
	return start.UTC(), end.UTC()
}

// YesterdayShanghai returns yesterday's calendar date at midnight in Shanghai (relative to now).
func YesterdayShanghai(now time.Time) time.Time {
	loc := Location()
	n := now.In(loc)
	y, m, d := n.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return today.AddDate(0, 0, -1)
}

// LastNDaysRangeUTC returns UTC [start, end) covering the last n Shanghai calendar days
// ending yesterday? No — "近 7 日" usually means last 7 days including today.
// Plan: 近 7 日 PV 趋势 — 7 points. Include today and previous 6 days (7 days total).
func LastNDaysRangeUTC(now time.Time, n int) (startUTC, endUTC time.Time, days []time.Time) {
	if n <= 0 {
		n = 7
	}
	loc := Location()
	cur := now.In(loc)
	y, m, d := cur.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)
	startLocal := today.AddDate(0, 0, -(n - 1))
	endLocal := today.Add(24 * time.Hour)
	days = make([]time.Time, 0, n)
	for i := 0; i < n; i++ {
		days = append(days, startLocal.AddDate(0, 0, i))
	}
	return startLocal.UTC(), endLocal.UTC(), days
}

// FormatDate formats a time as YYYY-MM-DD in Shanghai.
func FormatDate(t time.Time) string {
	return t.In(Location()).Format("2006-01-02")
}
