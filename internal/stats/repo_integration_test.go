package stats

import (
	"context"
	"os"
	"testing"
	"time"

	"shortlink/internal/channel"
	"shortlink/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Run with:
//
//	SHORTLINK_TEST_DSN='root:pass@tcp(127.0.0.1:3306)/shortlink?parseTime=true&loc=UTC&charset=utf8mb4' go test ./internal/stats/ -run Integration -count=1
func TestIntegrationSeedAndAggregations(t *testing.T) {
	dsn := os.Getenv("SHORTLINK_TEST_DSN")
	if dsn == "" {
		t.Skip("set SHORTLINK_TEST_DSN to run integration test")
	}
	conn := sqlx.NewMysql(dsn)
	links := model.NewShortUrlMapModel(conn)
	clicks := model.NewClickEventModel(conn)
	repo := NewRepo(clicks)
	ctx := context.Background()

	if err := SeedDemo(ctx, links, clicks, nil); err != nil {
		t.Fatalf("SeedDemo: %v", err)
	}

	yesterday := FormatDate(YesterdayShanghai(time.Now()))
	rank, err := repo.GetChannelRanking(ctx, yesterday)
	if err != nil {
		t.Fatal(err)
	}
	if len(rank) == 0 || rank[0].ChannelID != channel.SMS {
		t.Fatalf("yesterday top=%+v want sms", rank)
	}
	if rank[0].Pv < rank[1].Pv {
		t.Fatalf("sms should lead: %+v", rank)
	}

	trend, err := repo.GetPVTrend(ctx, 7, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(trend) != 7 {
		t.Fatalf("trend len=%d", len(trend))
	}
	var sum int64
	for _, p := range trend {
		sum += p.Pv
	}
	if sum == 0 {
		t.Fatal("expected non-zero site trend")
	}

	code := DemoCodeSMS
	pv, err := repo.GetCodePV(ctx, code)
	if err != nil || pv == 0 {
		t.Fatalf("pv=%d err=%v", pv, err)
	}
	uv, err := repo.GetCodeUVAndLastClick(ctx, code)
	if err != nil || uv.Uv == 0 || uv.LastClickAt == nil {
		t.Fatalf("uv=%+v err=%v", uv, err)
	}

	filtered, err := repo.GetPVTrend(ctx, 7, &code)
	if err != nil {
		t.Fatal(err)
	}
	var fsum int64
	for _, p := range filtered {
		fsum += p.Pv
	}
	if fsum == 0 || fsum > sum {
		t.Fatalf("filtered=%d site=%d", fsum, sum)
	}
}
