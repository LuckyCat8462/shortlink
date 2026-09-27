package cachex

import (
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"shortlink/pkg/rediskey"
)

const defaultTTL = 24 * time.Hour

// JumpCache is stored at surl:c:{code}.
type JumpCache struct {
	LongUrl  string     `json:"longUrl"`
	Status   int64      `json:"status"`
	ExpireAt *time.Time `json:"expireAt"`
	UrlHash  string     `json:"urlHash,omitempty"`
	SceneId  string     `json:"sceneId,omitempty"`
}

// IdemCache is stored at surl:h:{hash}.
type IdemCache struct {
	ShortCode string     `json:"shortCode"`
	LongUrl   string     `json:"longUrl"`
	ExpireAt  *time.Time `json:"expireAt"`
	Status    int64      `json:"status"`
}

func ComputeTTL(expireAt *time.Time, now time.Time) time.Duration {
	ttl := defaultTTL
	if expireAt != nil {
		remain := expireAt.Sub(now)
		if remain <= 0 {
			return 0
		}
		if remain < ttl {
			ttl = remain
		}
	}
	return ttl
}

func SetJump(r *redis.Redis, code string, v JumpCache, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	sec := int(ttl.Seconds())
	if sec < 1 {
		sec = 1
	}
	return r.Setex(rediskey.CacheByCode(code), string(b), sec)
}

func GetJump(r *redis.Redis, code string) (*JumpCache, error) {
	s, err := r.Get(rediskey.CacheByCode(code))
	if err != nil {
		return nil, err
	}
	if s == "" {
		return nil, nil
	}
	var v JumpCache
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func DelJump(r *redis.Redis, code string) {
	_, _ = r.Del(rediskey.CacheByCode(code))
}

func SetIdem(r *redis.Redis, hash string, v IdemCache, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	sec := int(ttl.Seconds())
	if sec < 1 {
		sec = 1
	}
	return r.Setex(rediskey.Idempotency(hash), string(b), sec)
}

func GetIdem(r *redis.Redis, hash string) (*IdemCache, error) {
	s, err := r.Get(rediskey.Idempotency(hash))
	if err != nil {
		return nil, err
	}
	if s == "" {
		return nil, nil
	}
	var v IdemCache
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return &v, nil
}
