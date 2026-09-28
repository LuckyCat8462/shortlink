package main

import (
	"flag"
	"fmt"
	"os"

	"shortlink/internal/config"
	"shortlink/internal/handler"
	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
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
}
