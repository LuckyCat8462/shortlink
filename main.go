package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"shortlink/internal/config"
	"shortlink/internal/handler"
	"shortlink/internal/stats"
	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

// 声明 flag 参数
var configFile = flag.String("f", "etc/shorturl-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	applyEnvOverrides(&c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	if c.SeedDemo {
		seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := stats.SeedDemo(seedCtx, ctx.ShortUrlModel, ctx.ClickEventModel, ctx.Filter); err != nil {
			logx.Errorf("SeedDemo failed: %v", err)
		}
	}
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// applyEnvOverrides lets Compose / cloud inject secrets without committing real yaml.
func applyEnvOverrides(c *config.Config) {
	if v := os.Getenv("SHORTLINK_DSN"); v != "" {
		c.ShortUrlDB.DSN = v
		c.SequenceDB.DSN = v
	}
	if v := os.Getenv("SHORTLINK_REDIS"); v != "" {
		if len(c.CacheRedis) == 0 {
			var node cache.NodeConf
			node.Host = v
			c.CacheRedis = cache.CacheConf{node}
		} else {
			c.CacheRedis[0].Host = v
		}
	}
	if v := os.Getenv("SHORTLINK_API_KEY"); v != "" {
		c.ApiKey = v
	}
	if v := os.Getenv("SHORTLINK_BASE62_SALT"); v != "" {
		c.Base62Salt = v
	}
	if v := os.Getenv("SHORTLINK_DOMAIN"); v != "" {
		c.ShortDomain = v
	}
	if v := os.Getenv("SHORTLINK_CLICK_LOG"); v != "" {
		c.ClickLogEnabled = parseBoolEnv(v)
	}
	if v := os.Getenv("SHORTLINK_SEED_DEMO"); v != "" {
		c.SeedDemo = parseBoolEnv(v)
	}
	if v := os.Getenv("SHORTLINK_LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("SHORTLINK_LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}
	if v := os.Getenv("SHORTLINK_LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
}

func parseBoolEnv(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
