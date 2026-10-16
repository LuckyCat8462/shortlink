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

	"github.com/cloudwego/eino/schema"
)

func main() {
	loadDotEnv(".env")

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cm, err := assistant.NewChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sys := assistant.SystemPrompt(time.Now())
	fmt.Println("=== system prompt (first 300 chars) ===")
	fmt.Println(trim(sys, 300))
	fmt.Println("=== probes ===")

	probes := []string{
		"你能帮我做什么？",
		"昨天哪个推广渠道点击更高？你先别查库，凭感觉猜一个。",
		"帮我写一首诗。",
	}

	for _, q := range probes {
		resp, err := cm.Generate(ctx, []*schema.Message{
			{Role: schema.System, Content: sys},
			{Role: schema.User, Content: q},
		})
		if err != nil {
			log.Fatalf("q=%q: %v", q, err)
		}
		fmt.Printf("\nQ: %s\nA: %s\n", q, strings.TrimSpace(resp.Content))
	}
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// loadDotEnv loads KEY=VALUE lines into the process env if not already set.
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
		i := strings.IndexByte(line, '=')
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		if len(v) >= 2 {
			if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
				v = v[1 : len(v)-1]
			}
		}
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}