package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"shortlink/internal/assistant"
	"shortlink/internal/stats"
	"shortlink/model"

	"github.com/cloudwego/eino/components/tool"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := os.Getenv("SHORTLINK_DSN")
	if dsn == "" {
		// 本机直连示例（需先把 compose mysql 映到 3306）
		dsn = "root:shortlink_dev_pass@tcp(127.0.0.1:3307)/shortlink?parseTime=true&loc=UTC&charset=utf8mb4"
	}

	conn := sqlx.NewMysql(dsn)
	if err := stats.EnsureSchema(ctx, conn); err != nil {
		log.Fatalf("EnsureSchema: %v", err)
	}

	repo := stats.NewRepo(model.NewClickEventModel(conn))
	tools, err := assistant.NewStatsTools(repo)
	if err != nil {
		log.Fatalf("NewStatsTools: %v", err)
	}

	yesterday := stats.FormatDate(stats.YesterdayShanghai(time.Now()))
	code := stats.DemoCodeSMS

	cases := []struct {
		name string
		args string
	}{
		{"get_code_pv", fmt.Sprintf(`{"shortCode":%q}`, code)},
		{"get_code_uv_and_last_click", fmt.Sprintf(`{"shortCode":%q}`, code)},
		{"get_channel_ranking", fmt.Sprintf(`{"date":%q}`, yesterday)},
		{"get_pv_trend", `{"days":7}`},
		{"get_pv_trend", fmt.Sprintf(`{"days":7,"shortCode":%q}`, code)},
	}

	byName := map[string]tool.InvokableTool{}
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Fatalf("Info: %v", err)
		}
		inv, ok := t.(tool.InvokableTool)
		if !ok {
			log.Fatalf("tool %s is not InvokableTool", info.Name)
		}
		byName[info.Name] = inv
		fmt.Printf("tool ready: %s — %s\n", info.Name, info.Desc)
	}
	fmt.Println("---")

	for _, c := range cases {
		inv, ok := byName[c.name]
		if !ok {
			log.Fatalf("missing tool %s", c.name)
		}
		out, err := inv.InvokableRun(ctx, c.args)
		if err != nil {
			log.Fatalf("%s(%s): %v", c.name, c.args, err)
		}
		fmt.Printf("%s %s\n=> %s\n\n", c.name, c.args, out)
	}

	fmt.Printf("hint: yesterday(Shanghai)=%s expected top channel=%s\n",
		yesterday, stats.ExpectedYesterdayTopChannel())
}