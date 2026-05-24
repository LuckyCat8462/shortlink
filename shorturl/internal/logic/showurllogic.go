// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shorturl/internal/svc"
	"shorturl/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShowUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShowUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShowUrlLogic {
	return &ShowUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShowUrlLogic) ShowUrl(req *types.ShowUrlRequest) (resp *types.ShowUrlResponse, err error) {
	// 查看短链接
	// 1.对方输入一个短链接，重定向到真实的长链接
	// 1.1.查缓存之前就要使用布隆过滤器，不存在的短链，直接返回404，不需要后续处理，减少消耗
	// a.基于内存版本,服务重启后消失，每次重启都要重新加载已有的短连接

	// b.基于redis版本，go-zero自带
	exist, err := l.svcCtx.Filter.Exists([]byte(req.ShortURL))
	if err != nil {
		logx.Errorw("Bloom Filet err: %v", logx.LogField{Value: err, Key: "err"})
	}
	if !exist {
		return nil, errors.New("404 not exist")
	}
	fmt.Println("开始查询缓存")
	// 1.2.根据短链接查询长链接，single flight
	// 同时有10w个请求baidu.com/test01，single flight可以合并并发的请求，第一个请求去查询，其余所有请求等待第一个请求的结果，不需要自己去查询
	shortURL, err := l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{String: req.ShortURL, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("404-short链接不存在")
		}
		logx.Errorw("ShowUrl findOneBySurl err: %v", logx.LogField{Value: err, Key: "err"})
		return nil, err
	}
	// 1.3. 返回查询到的长连接，在调用handler层返回重定向响应
	return &types.ShowUrlResponse{
		LongURL: shortURL.Lurl.String,
	}, nil
}
