package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ClickEvent struct {
	Id          uint64    `db:"id"`
	ShortCode   string    `db:"short_code"`
	SceneId     string    `db:"scene_id"`
	VisitorHash string    `db:"visitor_hash"`
	ClickedAt   time.Time `db:"clicked_at"`
}

type ClickEventModel interface {
	Insert(ctx context.Context, data *ClickEvent) (sql.Result, error)
	InsertBatch(ctx context.Context, rows []*ClickEvent) error
	CountByShortCode(ctx context.Context, shortCode string) (int64, error)
	CountDistinctVisitor(ctx context.Context, shortCode string) (int64, error)
	MaxClickedAt(ctx context.Context, shortCode string) (*time.Time, error)
	ChannelPVBetween(ctx context.Context, startUTC, endUTC time.Time) ([]ChannelPVRow, error)
	DailyPVBetween(ctx context.Context, startUTC, endUTC time.Time, shortCode *string) ([]DailyPVRow, error)
	CountAll(ctx context.Context) (int64, error)
	DeleteByShortCodes(ctx context.Context, codes []string) error
}

type ChannelPVRow struct {
	SceneId string `db:"scene_id"`
	Pv      int64  `db:"pv"`
}

type DailyPVRow struct {
	Day time.Time `db:"day"`
	Pv  int64     `db:"pv"`
}

type defaultClickEventModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewClickEventModel(conn sqlx.SqlConn) ClickEventModel {
	return &defaultClickEventModel{conn: conn, table: "`click_events`"}
}

func (m *defaultClickEventModel) Insert(ctx context.Context, data *ClickEvent) (sql.Result, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (`short_code`,`scene_id`,`visitor_hash`,`clicked_at`) VALUES (?,?,?,?)",
		m.table,
	)
	return m.conn.ExecCtx(ctx, query, data.ShortCode, data.SceneId, data.VisitorHash, data.ClickedAt.UTC())
}

func (m *defaultClickEventModel) InsertBatch(ctx context.Context, rows []*ClickEvent) error {
	if len(rows) == 0 {
		return nil
	}
	const chunk = 200
	for i := 0; i < len(rows); i += chunk {
		end := i + chunk
		if end > len(rows) {
			end = len(rows)
		}
		if err := m.insertChunk(ctx, rows[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (m *defaultClickEventModel) insertChunk(ctx context.Context, rows []*ClickEvent) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (`short_code`,`scene_id`,`visitor_hash`,`clicked_at`) VALUES ",
		m.table,
	)
	args := make([]any, 0, len(rows)*4)
	for i, row := range rows {
		if i > 0 {
			query += ","
		}
		query += "(?,?,?,?)"
		args = append(args, row.ShortCode, row.SceneId, row.VisitorHash, row.ClickedAt.UTC())
	}
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *defaultClickEventModel) CountByShortCode(ctx context.Context, shortCode string) (int64, error) {
	var n int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `short_code`=?", m.table)
	err := m.conn.QueryRowCtx(ctx, &n, query, shortCode)
	return n, err
}

func (m *defaultClickEventModel) CountDistinctVisitor(ctx context.Context, shortCode string) (int64, error) {
	var n int64
	query := fmt.Sprintf("SELECT COUNT(DISTINCT `visitor_hash`) FROM %s WHERE `short_code`=?", m.table)
	err := m.conn.QueryRowCtx(ctx, &n, query, shortCode)
	return n, err
}

func (m *defaultClickEventModel) MaxClickedAt(ctx context.Context, shortCode string) (*time.Time, error) {
	// go-zero QueryRowCtx scans aggregates more reliably into a named struct field.
	var row struct {
		MaxAt sql.NullTime `db:"max_at"`
	}
	query := fmt.Sprintf("SELECT MAX(`clicked_at`) AS `max_at` FROM %s WHERE `short_code`=?", m.table)
	err := m.conn.QueryRowCtx(ctx, &row, query, shortCode)
	if err != nil {
		return nil, err
	}
	if !row.MaxAt.Valid {
		return nil, nil
	}
	utc := row.MaxAt.Time.UTC()
	return &utc, nil
}

func (m *defaultClickEventModel) ChannelPVBetween(ctx context.Context, startUTC, endUTC time.Time) ([]ChannelPVRow, error) {
	var rows []ChannelPVRow
	query := fmt.Sprintf(
		"SELECT `scene_id`, COUNT(*) AS `pv` FROM %s WHERE `clicked_at`>=? AND `clicked_at`<? GROUP BY `scene_id` ORDER BY `pv` DESC, `scene_id` ASC",
		m.table,
	)
	err := m.conn.QueryRowsCtx(ctx, &rows, query, startUTC.UTC(), endUTC.UTC())
	return rows, err
}

// DailyPVBetween groups by Asia/Shanghai calendar day; startUTC/endUTC must be the UTC span of those days.
func (m *defaultClickEventModel) DailyPVBetween(ctx context.Context, startUTC, endUTC time.Time, shortCode *string) ([]DailyPVRow, error) {
	var rows []DailyPVRow
	base := fmt.Sprintf(
		"SELECT DATE(CONVERT_TZ(`clicked_at`, '+00:00', '+08:00')) AS `day`, COUNT(*) AS `pv` FROM %s WHERE `clicked_at`>=? AND `clicked_at`<?",
		m.table,
	)
	var err error
	if shortCode != nil && *shortCode != "" {
		query := base + " AND `short_code`=? GROUP BY `day` ORDER BY `day` ASC"
		err = m.conn.QueryRowsCtx(ctx, &rows, query, startUTC.UTC(), endUTC.UTC(), *shortCode)
	} else {
		query := base + " GROUP BY `day` ORDER BY `day` ASC"
		err = m.conn.QueryRowsCtx(ctx, &rows, query, startUTC.UTC(), endUTC.UTC())
	}
	return rows, err
}

func (m *defaultClickEventModel) CountAll(ctx context.Context) (int64, error) {
	var n int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	err := m.conn.QueryRowCtx(ctx, &n, query)
	return n, err
}

func (m *defaultClickEventModel) DeleteByShortCodes(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE `short_code` IN (", m.table)
	args := make([]any, len(codes))
	for i, c := range codes {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = c
	}
	query += ")"
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}
