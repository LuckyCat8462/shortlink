package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	"unicode"

	"shortlink/internal/cachex"
	"shortlink/internal/svc"
	"shortlink/internal/types"
	"shortlink/model"
	"shortlink/pkg/rediskey"

	"github.com/zeromicro/go-zero/core/logx"
)

var ErrNotFound = errors.New("not found")

type RedirectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRedirectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedirectLogic {
	return &RedirectLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Resolve returns long URL or ErrNotFound (unified 404).
func (l *RedirectLogic) Resolve(req *types.RedirectReq) (string, error) {
	code := req.ShortCode
	if !validShortCode(code, l.svcCtx.Config.ShortCodeLength) {
		return "", ErrNotFound
	}
	if _, banned := l.svcCtx.ShortURLBlacklist[code]; banned {
		return "", ErrNotFound
	}

	exist, err := l.svcCtx.Filter.Exists([]byte(code))
	if err != nil {
		logx.Errorw("Bloom.Exists failed", logx.LogField{Key: "err", Value: err.Error()})
	} else if !exist {
		return "", ErrNotFound
	}

	if cached, err := cachex.GetJump(l.svcCtx.Redis, code); err == nil && cached != nil {
		if long, ok := l.checkMapping(code, cached.LongUrl, cached.Status, cached.ExpireAt); ok {
			return long, nil
		}
		return "", ErrNotFound
	}

	v, err := l.svcCtx.SingleFlight.Do("load:"+code, func() (any, error) {
		row, err := l.svcCtx.ShortUrlModel.FindOneByShortCode(l.ctx, code)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		var exp *time.Time
		if row.ExpireAt.Valid {
			t := row.ExpireAt.Time
			exp = &t
		}
		ttl := cachex.ComputeTTL(exp, time.Now())
		_ = cachex.SetJump(l.svcCtx.Redis, row.ShortCode, cachex.JumpCache{
			LongUrl:  row.LongUrl,
			Status:   row.Status,
			ExpireAt: exp,
			UrlHash:  row.UrlHash,
			SceneId:  row.SceneId,
		}, ttl)
		return row, nil
	})
	if err != nil {
		return "", ErrNotFound
	}
	row, ok := v.(*model.ShortUrlMap)
	if !ok || row == nil {
		return "", ErrNotFound
	}
	var exp *time.Time
	if row.ExpireAt.Valid {
		t := row.ExpireAt.Time
		exp = &t
	}
	if long, ok := l.checkMapping(code, row.LongUrl, row.Status, exp); ok {
		return long, nil
	}
	return "", ErrNotFound
}

func (l *RedirectLogic) checkMapping(code, longURL string, status int64, expireAt *time.Time) (string, bool) {
	if status != 1 {
		cachex.DelJump(l.svcCtx.Redis, code)
		return "", false
	}
	if expireAt != nil && !time.Now().Before(*expireAt) {
		cachex.DelJump(l.svcCtx.Redis, code)
		return "", false
	}
	return longURL, true
}

// RecordClick runs asynchronously; must not block redirect.
func (l *RedirectLogic) RecordClick(code, clientIP string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("RecordClick panic: %v", r)
			}
		}()
		ipHash := hashIP(clientIP)
		now := time.Now().Format(time.RFC3339)
		if _, err := l.svcCtx.Redis.Incr(rediskey.PV(code)); err != nil {
			logx.Errorw("pv incr failed", logx.LogField{Key: "err", Value: err.Error()})
		}
		if _, err := l.svcCtx.Redis.Pfadd(rediskey.UV(code), ipHash); err != nil {
			logx.Errorw("uv pfadd failed", logx.LogField{Key: "err", Value: err.Error()})
		}
		if err := l.svcCtx.Redis.Set(rediskey.LastClick(code), now); err != nil {
			logx.Errorw("lastClick set failed", logx.LogField{Key: "err", Value: err.Error()})
		}
	}()
}

func validShortCode(code string, length int) bool {
	if length <= 0 {
		length = 6
	}
	if len(code) != length {
		return false
	}
	for _, r := range code {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:8])
}
