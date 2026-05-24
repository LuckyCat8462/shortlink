// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"shorturl/internal/config"
	"shorturl/model"
	"shorturl/sequence"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config            config.Config
	ShortUrlModel     model.ShortUrlMapModel
	Sequence          sequence.Sequence
	ShortURLBlacklist map[string]struct{}
	// 	布隆过滤器
	Filter *bloom.Filter
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.ShortUrlDB.DSN)
	// 构造黑名单map时，从config中读取黑名单列表
	m := make(map[string]struct{}, len(c.ShortURLBlacklist))
	for _, v := range c.ShortURLBlacklist {
		m[v] = struct{}{}
	}
	// 初始化布隆过滤器
	// 初始化 bitset
	store := redis.New(c.CacheRedis[0].Host, func(r *redis.Redis) {
		r.Type = redis.NodeType
	})
	// 声明一个bitset
	filter := bloom.New(store, "test_key", 20*(1<<20))
	// 加载已有的短连接数据
	// filter := bloom.NewWithEstimates(1000000, 0.01)
	return &ServiceContext{
		Config:            c,
		ShortUrlModel:     model.NewShortUrlMapModel(conn, c.CacheRedis),
		Sequence:          sequence.NewMySQL(c.SequenceDB.DSN),
		ShortURLBlacklist: m,
		Filter:            filter,
	}
}

// 加载已有的数据到布隆过滤器
func loadDataToBloomFilter() {

}
