package svc

import (
	"context"
	"fmt"
	"time"

	"shortlink/internal/assistant"
	"shortlink/internal/config"
	"shortlink/internal/stats"
	"shortlink/model"
	"shortlink/pkg/base62"
	"shortlink/pkg/codec"
	"shortlink/pkg/rediskey"
	"shortlink/sequence"

	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
)

type ServiceContext struct {
	Config            config.Config
	ShortUrlModel     model.ShortUrlMapModel
	ClickEventModel   model.ClickEventModel
	Stats             *stats.Repo
	Chat              *assistant.ChatService
	Sequence          sequence.Sequence
	ShortURLBlacklist map[string]struct{}
	Filter            *bloom.Filter
	Redis             *redis.Redis
	Encoder           *base62.Encoder
	Mixer             *codec.Mixer
	SingleFlight      syncx.SingleFlight
	RedirectLimiter   *limit.PeriodLimit
	CreateLimiter     *limit.PeriodLimit
	AssistantLimiter  *limit.PeriodLimit
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.ShortCodeLength <= 0 {
		c.ShortCodeLength = 6
	}
	if c.MaxExpireDays <= 0 {
		c.MaxExpireDays = 3650
	}
	if c.DefaultExpireDays < 0 {
		c.DefaultExpireDays = 90
	}
	if c.RedirectRatePerMin <= 0 {
		c.RedirectRatePerMin = 120
	}
	if c.CreateRatePerMin <= 0 {
		c.CreateRatePerMin = 60
	}
	if c.AssistantRatePerMin <= 0 {
		c.AssistantRatePerMin = 20
	}
	if c.AssistantTimeoutSec <= 0 {
		c.AssistantTimeoutSec = 90
	}
	if c.AssistantSessionTTLSec <= 0 {
		c.AssistantSessionTTLSec = 7200
	}
	if c.AssistantMaxHistory <= 0 {
		c.AssistantMaxHistory = 20
	}
	if c.LLM.BaseURL == "" {
		c.LLM.BaseURL = "https://api.deepseek.com"
	}
	if c.Base62Chars == "" {
		c.Base62Chars = base62.DefaultCharset
	}
	if c.Base62Salt == "" {
		c.Base62Salt = "default-salt"
	}
	if c.SequenceType == "" {
		c.SequenceType = "mysql"
	}

	conn := sqlx.NewMysql(c.ShortUrlDB.DSN)

	bl := make(map[string]struct{}, len(c.ShortURLBlacklist))
	for _, v := range c.ShortURLBlacklist {
		bl[v] = struct{}{}
	}

	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.CacheRedis[0].Host,
		Type: redis.NodeType,
	})

	// ~16M bits: demo-scale capacity for ~1M codes @ ~1% FPR ballpark
	filter := bloom.New(rds, rediskey.BloomKey, 1<<24)

	enc, err := base62.NewEncoder(c.Base62Chars, c.ShortCodeLength)
	if err != nil {
		panic(fmt.Sprintf("base62 encoder: %v", err))
	}
	mixer, err := codec.NewMixerFromSalt(c.Base62Salt, codec.ModulusForLength(c.ShortCodeLength))
	if err != nil {
		panic(fmt.Sprintf("codec mixer: %v", err))
	}

	var seq sequence.Sequence
	switch c.SequenceType {
	case "redis":
		seq = sequence.NewRedis(c.CacheRedis[0].Host)
	default:
		seq = sequence.NewMySQL(c.SequenceDB.DSN)
	}

	schemaCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stats.EnsureSchema(schemaCtx, conn); err != nil {
		logx.Errorf("EnsureSchema click_events: %v", err)
	}

	clickModel := model.NewClickEventModel(conn)
	statsRepo := stats.NewRepo(clickModel)
	llmCfg := assistant.LLMConfig{
		APIKey:  c.LLM.APIKey,
		BaseURL: c.LLM.BaseURL,
		Model:   c.LLM.Model,
	}
	chatSvc := &assistant.ChatService{
		Repo:     statsRepo,
		Sessions: assistant.NewSessionStore(rds, c.AssistantMaxHistory, c.AssistantSessionTTLSec),
		LLM:      llmCfg,
	}
	return &ServiceContext{
		Config:            c,
		ShortUrlModel:     model.NewShortUrlMapModel(conn),
		ClickEventModel:   clickModel,
		Stats:             statsRepo,
		Chat:              chatSvc,
		Sequence:          seq,
		ShortURLBlacklist: bl,
		Filter:            filter,
		Redis:             rds,
		Encoder:           enc,
		Mixer:             mixer,
		SingleFlight:      syncx.NewSingleFlight(),
		RedirectLimiter:   limit.NewPeriodLimit(60, c.RedirectRatePerMin, rds, rediskey.PrefixRateRedirect),
		CreateLimiter:     limit.NewPeriodLimit(60, c.CreateRatePerMin, rds, rediskey.PrefixRateCreate),
		AssistantLimiter:  limit.NewPeriodLimit(60, c.AssistantRatePerMin, rds, rediskey.PrefixRateAssistant),
	}
}
