package sequence

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Redis 使用 Redis 实现的发号器
type Redis struct {
	client *redis.Redis
}

// NewRedis 创建 Redis 发号器实例
func NewRedis(addr string) *Redis {
	conf := redis.RedisConf{
		Host: addr,
		Type: "node",
	}
	return &Redis{
		client: conf.NewRedis(),
	}
}

// Next 获取下一个序列号
// 使用 Redis INCR 命令实现原子自增
func (r *Redis) Next() (seq uint64, err error) {
	result, err := r.client.Incr("shortlink:sequence")
	if err != nil {
		logx.Errorw("redis.Incr failed", logx.LogField{Key: "err", Value: err.Error()})
		return 0, err
	}

	return uint64(result), nil
}
