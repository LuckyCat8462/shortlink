package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

// Config 服务配置，对应 etc/shorturl-api.yaml
type Config struct {
	rest.RestConf

	ShortUrlDB struct { // 短链映射表数据库
		DSN string
	}
	SequenceDB struct { // 发号器数据库
		DSN string
	}
	CacheRedis cache.CacheConf // Redis 缓存

	ShortDomain         string   // 短链域名
	ApiKey              string   // 创建接口鉴权密钥
	ShortCodeLength     int      // 短码长度
	Base62Chars         string   // Base62 字符集
	Base62Salt          string   // 编码盐值
	DefaultExpireDays   int      // 默认过期天数
	MaxExpireDays       int      // 最大过期天数
	CheckLongURL        bool     // 是否校验长链可达
	EnableHostWhitelist bool     // 是否启用长链域名白名单
	ShortURLBlacklist   []string // 禁止作为短码的保留词
	SequenceType        string   // 发号器类型：mysql / redis
	ClickLogEnabled     bool     // 是否写入 click_events（分析助手数据源）
	SeedDemo            bool     // 启动时写入演示短链与点击明细

	RedirectRatePerMin int // 跳转限流（次/分钟）
	CreateRatePerMin   int // 创建限流（次/分钟）
	AssistantRatePerMin int // 分析助手限流（次/分钟）
	AssistantTimeoutSec int // 分析助手单次超时（秒）
	AssistantSessionTTLSec int // session TTL（秒）
	AssistantMaxHistory    int // session 保留消息条数

	LLM struct {
		APIKey  string
		BaseURL string
		Model   string
	}
}
