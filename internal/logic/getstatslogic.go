package logic

import (
	"context"
	"errors"
	"strconv"

	"shortlink/internal/svc"
	"shortlink/internal/types"
	"shortlink/model"
	"shortlink/pkg/rediskey"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatsLogic {
	return &GetStatsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetStatsLogic) GetStats(req *types.GetStatsReq) (*types.GetStatsResp, error) {
	_, err := l.svcCtx.ShortUrlModel.FindOneByShortCode(l.ctx, req.ShortCode)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, &httpError{code: 404, msg: "not found"}
		}
		return nil, err
	}

	var pv int64
	if s, err := l.svcCtx.Redis.Get(rediskey.PV(req.ShortCode)); err == nil && s != "" {
		pv, _ = strconv.ParseInt(s, 10, 64)
	}
	uv, _ := l.svcCtx.Redis.Pfcount(rediskey.UV(req.ShortCode))
	var last *string
	if s, err := l.svcCtx.Redis.Get(rediskey.LastClick(req.ShortCode)); err == nil && s != "" {
		last = &s
	}

	return &types.GetStatsResp{
		ShortCode:   req.ShortCode,
		Pv:          pv,
		Uv:          uv,
		LastClickAt: last,
	}, nil
}
