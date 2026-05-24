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
	"shorturl/model"
	"shorturl/pkg/base62"
	"shorturl/pkg/connect"
	"shorturl/pkg/md5"
	"shorturl/pkg/urltool"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// Shorturl 输入长链接，转短链接
func (l *ShorturlLogic) Shorturl(req *types.ConvertRequest) (resp *types.ConvertResponse, err error) {
	// todo: add your logic here and delete this line
	// 1、校验长链接数据
	// 使用第三方库validator包来做参数校验
	// 1.1.数据不能为空
	// 1.2.链接请求必须有效，能访问
	// http.Get(req.LongURL)
	if ok := connect.Get(req.LongURL); !ok {
		return nil, errors.New("无效链接")
	}
	// 1.3.判断是否已经转链过了（数据库中是否已存在该短链接）
	// 1.3.1.为长连接生成md5
	md5Value := md5.Sum([]byte(req.LongURL)) // 此处使用的是项目中封装的md5包
	// 1.3.2.用md5值，去库里查询是否存在
	u, err := l.svcCtx.ShortUrlModel.FindOneByMd5(l.ctx, sql.NullString{String: md5Value, Valid: true})
	if !errors.Is(err, sqlx.ErrNotFound) {
		if err == nil {
			return nil, fmt.Errorf("该链接已被转为%s", u.Surl.String)
		}
		logx.Errorw("shortUrlModel.FindOneByMD5 filed", logx.LogField{
			Key:   "err",
			Value: err.Error(),
		})
		return nil, err
	}
	// 1.4.避免循环转链（输入不能为短链接）
	basePath, err := urltool.GetBasePath(req.LongURL)
	if err != nil {
		logx.Errorw("url.parse filed", logx.LogField{
			Key:   "err",
			Value: err.Error(),
		})
		return nil, err
	}
	u, err = l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{String: basePath, Valid: true})
	if !errors.Is(err, sqlx.ErrNotFound) {
		if err == nil {
			return nil, errors.New("该链接已经是短链了")
		}
		logx.Errorw("shorturlmodel.findonebySurl filed", logx.LogField{
			Key:   "err",
			Value: err.Error(),
		})
		return nil, err
	}

	// 2、取号
	// 基于MySQL实现的发号器
	// 每来一个转链请求，就使用replace into语句，向sequence表插入数据，取出主键id作为号码
	seqID, err := l.svcCtx.Sequence.Next()
	if err != nil {
		logx.Errorw("Sequence.Next() failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	// fmt.Println(seqID)
	_ = seqID // 后续可以使用这个seqID来生成短链接
	// 3、将号码转短链
	shortURL := base62.Int2String(seqID)
	// 3.1.考虑安全性,打乱base62字符串
	// 3.2.考虑关键词，避免使用敏感词，设立黑名单机制，例如health,status,metrics,fuxk,convert
	if _, ok := l.svcCtx.ShortURLBlacklist[shortURL]; ok {
		return nil, fmt.Errorf("短链接不能包含敏感词%s", shortURL)
	}

	// 4、存储长短链映射关系
	if _, err := l.svcCtx.ShortUrlModel.Insert(l.ctx, &model.ShortUrlMap{
		Lurl: sql.NullString{String: req.LongURL, Valid: true},
		Surl: sql.NullString{String: shortURL, Valid: true},
		Md5:  sql.NullString{String: md5Value, Valid: true},
	}); err != nil {
		logx.Errorw("shortUrlModel.Insert filed", logx.LogField{
			Key:   "err",
			Value: err.Error(),
		})
		return nil, err
	}
	// 将生成的短连接，加到布隆过滤器中
	if err := l.svcCtx.Filter.Add([]byte(shortURL)); err != nil {
		logx.Errorw("Filter.Add filed", logx.LogField{
			Key:   "err",
			Value: err.Error(),
		})
	}

	// 5、返回响应
	shortURL = l.svcCtx.Config.ShortDomain + "/" + shortURL
	return &types.ConvertResponse{ShortURL: shortURL}, nil
}
