// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"shorturl/internal/svc"
	"shorturl/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShorturlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShorturlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShorturlLogic {
	return &ShorturlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShorturlLogic) Shorturl(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
