package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"shortlink/internal/cachex"
	"shortlink/internal/svc"
	"shortlink/internal/types"
	"shortlink/model"
	"shortlink/pkg/connect"
	"shortlink/pkg/urltool"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLinkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLinkLogic {
	return &CreateLinkLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateLinkLogic) CreateLink(req *types.CreateLinkReq) (*types.CreateLinkResp, error) {
	longURL, err := urltool.NormalizeLongURL(req.LongUrl)
	if err != nil {
		return nil, errBadRequest(err.Error())
	}

	expireDays := l.svcCtx.Config.DefaultExpireDays
	if req.ExpireDays != nil {
		expireDays = *req.ExpireDays
	}
	if expireDays < 0 || expireDays > l.svcCtx.Config.MaxExpireDays {
		return nil, errBadRequest("invalid expireDays")
	}

	if l.svcCtx.Config.CheckLongURL {
		if ok := connect.Get(longURL); !ok {
			return nil, errBadRequest("longUrl is not reachable")
		}
	}

	sceneID := req.SceneId
	urlHash := urltool.URLHash(sceneID, longURL)

	if cached, err := cachex.GetIdem(l.svcCtx.Redis, urlHash); err == nil && cached != nil {
		return l.buildResp(cached.ShortCode, cached.LongUrl, cached.ExpireAt, false), nil
	}

	if row, err := l.svcCtx.ShortUrlModel.FindOneBySceneHash(l.ctx, sceneID, urlHash); err == nil {
		l.writeCaches(row)
		return l.respFromRow(row, false), nil
	} else if !errors.Is(err, model.ErrNotFound) {
		logx.Errorw("FindOneBySceneHash failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}

	var expireAt sql.NullTime
	var expirePtr *time.Time
	if expireDays > 0 {
		t := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)
		expireAt = sql.NullTime{Time: t, Valid: true}
		expirePtr = &t
	}

	const maxRetry = 10
	var code string
	generated := false
	for i := 0; i < maxRetry; i++ {
		seq, err := l.svcCtx.Sequence.Next()
		if err != nil {
			return nil, err
		}
		mixed := l.svcCtx.Mixer.Mix(seq)
		candidate := l.svcCtx.Encoder.EncodeFixed(mixed)
		if _, banned := l.svcCtx.ShortURLBlacklist[candidate]; banned {
			continue
		}
		code = candidate
		generated = true
		break
	}
	if !generated {
		return nil, errors.New("generate short code failed")
	}

	row := &model.ShortUrlMap{
		SceneId:   sceneID,
		UrlHash:   urlHash,
		LongUrl:   longURL,
		ShortCode: code,
		Status:    1,
		ExpireAt:  expireAt,
		CreatedBy: "api",
	}

	if _, err := l.svcCtx.ShortUrlModel.Insert(l.ctx, row); err != nil {
		if model.IsDuplicateErr(err) {
			if existing, e2 := l.svcCtx.ShortUrlModel.FindOneBySceneHash(l.ctx, sceneID, urlHash); e2 == nil {
				l.writeCaches(existing)
				return l.respFromRow(existing, false), nil
			}
			// short_code conflict - rare; ask client to retry
			return nil, fmt.Errorf("short code conflict, please retry")
		}
		logx.Errorw("Insert failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}

	if err := l.svcCtx.Filter.Add([]byte(code)); err != nil {
		logx.Errorw("Bloom.Add failed", logx.LogField{Key: "err", Value: err.Error()})
	}
	l.writeCaches(row)
	_ = expirePtr
	return l.respFromRow(row, true), nil
}

func (l *CreateLinkLogic) writeCaches(row *model.ShortUrlMap) {
	now := time.Now()
	var exp *time.Time
	if row.ExpireAt.Valid {
		t := row.ExpireAt.Time
		exp = &t
	}
	ttl := cachex.ComputeTTL(exp, now)
	_ = cachex.SetJump(l.svcCtx.Redis, row.ShortCode, cachex.JumpCache{
		LongUrl:  row.LongUrl,
		Status:   row.Status,
		ExpireAt: exp,
		UrlHash:  row.UrlHash,
		SceneId:  row.SceneId,
	}, ttl)
	_ = cachex.SetIdem(l.svcCtx.Redis, row.UrlHash, cachex.IdemCache{
		ShortCode: row.ShortCode,
		LongUrl:   row.LongUrl,
		ExpireAt:  exp,
		Status:    row.Status,
	}, ttl)
}

func (l *CreateLinkLogic) respFromRow(row *model.ShortUrlMap, created bool) *types.CreateLinkResp {
	var exp *time.Time
	if row.ExpireAt.Valid {
		t := row.ExpireAt.Time
		exp = &t
	}
	return l.buildResp(row.ShortCode, row.LongUrl, exp, created)
}

func (l *CreateLinkLogic) buildResp(code, longURL string, exp *time.Time, created bool) *types.CreateLinkResp {
	var expStr *string
	if exp != nil {
		s := exp.Format(time.RFC3339)
		expStr = &s
	}
	domain := l.svcCtx.Config.ShortDomain
	return &types.CreateLinkResp{
		ShortUrl:  fmt.Sprintf("%s/%s", trimSlash(domain), code),
		ShortCode: code,
		LongUrl:   longURL,
		ExpireAt:  expStr,
		Created:   created,
	}
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

type httpError struct {
	code int
	msg  string
}

func (e *httpError) Error() string { return e.msg }

func errBadRequest(msg string) error {
	return &httpError{code: http.StatusBadRequest, msg: msg}
}

func HTTPStatusFromError(err error) (int, bool) {
	var he *httpError
	if errors.As(err, &he) {
		return he.code, true
	}
	return 0, false
}
