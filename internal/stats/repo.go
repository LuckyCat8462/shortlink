package stats

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"shortlink/internal/channel"
	"shortlink/model"
)

// RankRow is one channel's PV for a Shanghai calendar day.
type RankRow struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	Pv          int64  `json:"pv"`
}

// TrendPoint is site-wide or per-code daily PV.
type TrendPoint struct {
	Date string `json:"date"`
	Pv   int64  `json:"pv"`
}

// UVLastClick is exact UV + last click time (UTC RFC3339) for a short code.
type UVLastClick struct {
	ShortCode   string     `json:"shortCode"`
	Uv          int64      `json:"uv"`
	LastClickAt *time.Time `json:"lastClickAt"`
}

// Repo exposes read/write helpers for assistant tools (Phase A contract).
type Repo struct {
	clicks model.ClickEventModel
}

func NewRepo(clicks model.ClickEventModel) *Repo {
	return &Repo{clicks: clicks}
}

// VisitorHash is sha256(ip + "|" + ua) hex.
func VisitorHash(ip, ua string) string {
	sum := sha256.Sum256([]byte(ip + "|" + ua))
	return hex.EncodeToString(sum[:])
}

// RecordClick inserts one analytics row. clickedAt should be wall clock; stored as UTC.
func (r *Repo) RecordClick(ctx context.Context, shortCode, sceneID, ip, ua string, clickedAt time.Time) error {
	if shortCode == "" {
		return fmt.Errorf("empty shortCode")
	}
	_, err := r.clicks.Insert(ctx, &model.ClickEvent{
		ShortCode:   shortCode,
		SceneId:     sceneID,
		VisitorHash: VisitorHash(ip, ua),
		ClickedAt:   clickedAt.UTC(),
	})
	return err
}

// GetCodePV returns total PV for a short code (all time).
func (r *Repo) GetCodePV(ctx context.Context, shortCode string) (int64, error) {
	if shortCode == "" {
		return 0, fmt.Errorf("empty shortCode")
	}
	return r.clicks.CountByShortCode(ctx, shortCode)
}

// GetCodeUVAndLastClick returns exact UV and last click time.
func (r *Repo) GetCodeUVAndLastClick(ctx context.Context, shortCode string) (*UVLastClick, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("empty shortCode")
	}
	uv, err := r.clicks.CountDistinctVisitor(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	last, err := r.clicks.MaxClickedAt(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	return &UVLastClick{ShortCode: shortCode, Uv: uv, LastClickAt: last}, nil
}

// GetChannelRanking ranks channels by PV on a Shanghai calendar day (date = YYYY-MM-DD).
// Known demo channels with 0 PV are included so the model can compare fairly.
func (r *Repo) GetChannelRanking(ctx context.Context, dateShanghai string) ([]RankRow, error) {
	day, err := ParseDate(dateShanghai)
	if err != nil {
		return nil, err
	}
	start, end := DayRangeUTC(day)
	rows, err := r.clicks.ChannelPVBetween(ctx, start, end)
	if err != nil {
		return nil, err
	}
	pvByScene := make(map[string]int64, len(rows))
	for _, row := range rows {
		pvByScene[row.SceneId] = row.Pv
	}
	out := make([]RankRow, 0, len(channel.All()))
	for _, id := range channel.All() {
		out = append(out, RankRow{
			ChannelID:   id,
			ChannelName: channel.DisplayName(id),
			Pv:          pvByScene[id],
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Pv != out[j].Pv {
			return out[i].Pv > out[j].Pv
		}
		return out[i].ChannelID < out[j].ChannelID
	})
	return out, nil
}

// GetPVTrend returns daily PV for the last `days` Shanghai calendar days (including today).
// shortCode nil/empty => site-wide; otherwise filter by code. Missing days are filled with 0.
func (r *Repo) GetPVTrend(ctx context.Context, days int, shortCode *string) ([]TrendPoint, error) {
	if days <= 0 {
		days = 7
	}
	start, end, dayList := LastNDaysRangeUTC(time.Now(), days)
	var codePtr *string
	if shortCode != nil && *shortCode != "" {
		codePtr = shortCode
	}
	rows, err := r.clicks.DailyPVBetween(ctx, start, end, codePtr)
	if err != nil {
		return nil, err
	}
	pvByDay := make(map[string]int64, len(rows))
	for _, row := range rows {
		pvByDay[FormatDate(row.Day)] = row.Pv
	}
	out := make([]TrendPoint, 0, len(dayList))
	for _, d := range dayList {
		key := FormatDate(d)
		out = append(out, TrendPoint{Date: key, Pv: pvByDay[key]})
	}
	return out, nil
}
