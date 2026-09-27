package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ShortUrlMap struct {
	Id        uint64         `db:"id"`
	SceneId   string         `db:"scene_id"`
	UrlHash   string         `db:"url_hash"`
	LongUrl   string         `db:"long_url"`
	ShortCode string         `db:"short_code"`
	Status    int64          `db:"status"`
	ExpireAt  sql.NullTime   `db:"expire_at"`
	CreatedBy string         `db:"created_by"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type ShortUrlMapModel interface {
	Insert(ctx context.Context, data *ShortUrlMap) (sql.Result, error)
	FindOneBySceneHash(ctx context.Context, sceneId, urlHash string) (*ShortUrlMap, error)
	FindOneByShortCode(ctx context.Context, shortCode string) (*ShortUrlMap, error)
}

type defaultShortUrlMapModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewShortUrlMapModel(conn sqlx.SqlConn) ShortUrlMapModel {
	return &defaultShortUrlMapModel{conn: conn, table: "`short_url_map`"}
}

func (m *defaultShortUrlMapModel) Insert(ctx context.Context, data *ShortUrlMap) (sql.Result, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (`scene_id`,`url_hash`,`long_url`,`short_code`,`status`,`expire_at`,`created_by`) VALUES (?,?,?,?,?,?,?)",
		m.table,
	)
	return m.conn.ExecCtx(ctx, query,
		data.SceneId, data.UrlHash, data.LongUrl, data.ShortCode, data.Status, data.ExpireAt, data.CreatedBy,
	)
}

func (m *defaultShortUrlMapModel) FindOneBySceneHash(ctx context.Context, sceneId, urlHash string) (*ShortUrlMap, error) {
	var resp ShortUrlMap
	query := fmt.Sprintf(
		"SELECT `id`,`scene_id`,`url_hash`,`long_url`,`short_code`,`status`,`expire_at`,`created_by`,`created_at`,`updated_at` FROM %s WHERE `scene_id`=? AND `url_hash`=? LIMIT 1",
		m.table,
	)
	err := m.conn.QueryRowCtx(ctx, &resp, query, sceneId, urlHash)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultShortUrlMapModel) FindOneByShortCode(ctx context.Context, shortCode string) (*ShortUrlMap, error) {
	var resp ShortUrlMap
	query := fmt.Sprintf(
		"SELECT `id`,`scene_id`,`url_hash`,`long_url`,`short_code`,`status`,`expire_at`,`created_by`,`created_at`,`updated_at` FROM %s WHERE `short_code`=? LIMIT 1",
		m.table,
	)
	err := m.conn.QueryRowCtx(ctx, &resp, query, shortCode)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// IsDuplicateErr reports MySQL unique constraint violation.
func IsDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "Error 1062")
}
