package logic

import (
	"context"
	"errors"
	"time"

	"shortlink/internal/svc"
	"shortlink/internal/types"
	"shortlink/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLinkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLinkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLinkLogic {
	return &GetLinkLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetLinkLogic) GetLink(req *types.GetLinkReq) (*types.GetLinkResp, error) {
	row, err := l.svcCtx.ShortUrlModel.FindOneByShortCode(l.ctx, req.ShortCode)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, &httpError{code: 404, msg: "not found"}
		}
		return nil, err
	}
	var expStr *string
	if row.ExpireAt.Valid {
		s := row.ExpireAt.Time.Format(time.RFC3339)
		expStr = &s
	}
	return &types.GetLinkResp{
		ShortCode: row.ShortCode,
		LongUrl:   row.LongUrl,
		SceneId:   row.SceneId,
		Status:    int(row.Status),
		ExpireAt:  expStr,
		CreatedAt: row.CreatedAt.Format(time.RFC3339),
	}, nil
}
