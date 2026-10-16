package assistant

import (
	"context"
	"encoding/json"
	"fmt"

	"shortlink/internal/stats"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type codeArg struct {
	ShortCode string `json:"shortCode" jsonschema:"description=6-char short code, e.g. demos1"`
}

type rankingArg struct {
	Date string `json:"date" jsonschema:"description=Shanghai calendar day YYYY-MM-DD"`
}

type trendArg struct {
	Days      int    `json:"days,omitempty" jsonschema:"description=Number of days, default 7"`
	ShortCode string `json:"shortCode,omitempty" jsonschema:"description=Optional short code filter; empty = site-wide"`
}

func NewStatsTools(repo *stats.Repo) ([]tool.BaseTool, error) {
	if repo == nil {
		return nil, fmt.Errorf("nil stats repo")
	}

	t1, err := utils.InferTool(
		"get_code_pv",
		"Get total PV (click count) for one short code from MySQL click_events.",
		func(ctx context.Context, in *codeArg) (string, error) {
			pv, err := repo.GetCodePV(ctx, in.ShortCode)
			if err != nil {
				return "", err
			}
			return mustJSON(map[string]any{"shortCode": in.ShortCode, "pv": pv}), nil
		},
	)
	if err != nil {
		return nil, err
	}

	t2, err := utils.InferTool(
		"get_code_uv_and_last_click",
		"Get exact UV and last click time for one short code.",
		func(ctx context.Context, in *codeArg) (string, error) {
			out, err := repo.GetCodeUVAndLastClick(ctx, in.ShortCode)
			if err != nil {
				return "", err
			}
			return mustJSON(out), nil
		},
	)
	if err != nil {
		return nil, err
	}

	t3, err := utils.InferTool(
		"get_channel_ranking",
		"Rank promo channels by PV on a Shanghai calendar day (YYYY-MM-DD). Use for questions like which channel had more clicks yesterday.",
		func(ctx context.Context, in *rankingArg) (string, error) {
			rows, err := repo.GetChannelRanking(ctx, in.Date)
			if err != nil {
				return "", err
			}
			return mustJSON(rows), nil
		},
	)
	if err != nil {
		return nil, err
	}

	t4, err := utils.InferTool(
		"get_pv_trend",
		"Daily PV trend for last N Shanghai days (default 7). Omit shortCode for site-wide.",
		func(ctx context.Context, in *trendArg) (string, error) {
			days := in.Days
			if days <= 0 {
				days = 7
			}
			var code *string
			if in.ShortCode != "" {
				code = &in.ShortCode
			}
			points, err := repo.GetPVTrend(ctx, days, code)
			if err != nil {
				return "", err
			}
			return mustJSON(points), nil
		},
	)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{t1, t2, t3, t4}, nil
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"marshal failed"}`
	}
	return string(b)
}