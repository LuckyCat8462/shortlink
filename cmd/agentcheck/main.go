package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"shortlink/internal/assistant"
	"shortlink/internal/channel"
	"shortlink/internal/stats"
	"shortlink/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func main() {
	loadDotEnv(".env")
	loadKeyFile("key.md")
	if os.Getenv("SHORTLINK_LLM_API_KEY") == "" || os.Getenv("SHORTLINK_LLM_MODEL") == "" {
		log.Fatal("set SHORTLINK_LLM_API_KEY and SHORTLINK_LLM_MODEL (fill .env, or put the key alone in key.md)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	dsn := hostDSN(os.Getenv("SHORTLINK_DSN"))
	conn := sqlx.NewMysql(dsn)
	if err := stats.EnsureSchema(ctx, conn); err != nil {
		log.Fatalf("EnsureSchema: %v", err)
	}
	repo := stats.NewRepo(model.NewClickEventModel(conn))

	ag, err := assistant.NewStatsAgent(ctx, repo, assistant.LLMConfigFromEnv())
	if err != nil {
		log.Fatalf("NewStatsAgent: %v", err)
	}

	q := "昨天哪个推广渠道点击更高？"
	if len(os.Args) > 1 {
		q = strings.Join(os.Args[1:], " ")
	}
	fmt.Println("Q:", q)

	resp, err := assistant.Ask(ctx, ag, nil, q)
	if err != nil {
		log.Fatalf("Ask: %v", err)
	}
	answer := strings.TrimSpace(resp.Content)
	fmt.Println("A:", answer)

	yesterday := stats.FormatDate(stats.YesterdayShanghai(time.Now()))
	rank, err := repo.GetChannelRanking(ctx, yesterday)
	if err != nil || len(rank) == 0 {
		log.Fatalf("baseline ranking: %v", err)
	}
	top := rank[0]
	fmt.Printf("baseline: date=%s top=%s(%s) pv=%d\n", yesterday, top.ChannelName, top.ChannelID, top.Pv)

	ok := true
	if !strings.Contains(answer, top.ChannelName) && !strings.Contains(answer, "短信") {
		fmt.Println("CHECK FAIL: answer missing expected channel name", top.ChannelName)
		ok = false
	}
	pvStr := fmt.Sprintf("%d", top.Pv)
	if !strings.Contains(answer, pvStr) {
		fmt.Println("CHECK WARN: answer missing exact pv", pvStr, "(model may paraphrase; inspect manually)")
	}
	if top.ChannelID != channel.SMS {
		fmt.Println("CHECK FAIL: seed expectation broken, top is", top.ChannelID)
		ok = false
	}
	if ok {
		fmt.Println("B4 OK: agent used real ranking (短信 leading yesterday)")
	} else {
		os.Exit(1)
	}
}

func hostDSN(dsn string) string {
	if dsn == "" || strings.Contains(dsn, "@tcp(mysql:") {
		return "root:shortlink_dev_pass@tcp(127.0.0.1:3307)/shortlink?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	return dsn
}

// loadKeyFile loads a raw API key from a one-line file (e.g. key.md) if env is empty.
func loadKeyFile(path string) {
	if os.Getenv("SHORTLINK_LLM_API_KEY") != "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	key := strings.TrimSpace(string(b))
	if key == "" {
		return
	}
	_ = os.Setenv("SHORTLINK_LLM_API_KEY", key)
}

// loadDotEnv loads KEY=VAL lines from path into the process env (does not override existing).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		if _, exists := os.LookupEnv(k); exists {
			continue
		}
		_ = os.Setenv(k, v)
	}
}
