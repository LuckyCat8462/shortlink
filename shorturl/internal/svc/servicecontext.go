// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
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
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.ShortUrlDB.DSN)
	// 构造黑名单map时，从config中读取黑名单列表
	m := make(map[string]struct{}, len(c.ShortURLBlacklist))
	for _, v := range c.ShortURLBlacklist {
		m[v] = struct{}{}
	}
	return &ServiceContext{
		Config:            c,
		ShortUrlModel:     model.NewShortUrlMapModel(conn),
		Sequence:          sequence.NewMySQL(c.SequenceDB.DSN),
		ShortURLBlacklist: m,
	}
}
