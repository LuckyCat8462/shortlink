package sequence

import (
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const sqlReplaceIntoStub = `REPLACE INTO sequence (stub) VALUES ('a');`

// 建立MySQL连接执行replace into 语句
type MySQL struct {
	conn sqlx.SqlConn
}

func NewMySQL(dsn string) Sequence {
	return &MySQL{
		conn: sqlx.NewMysql(dsn),
	}
}

// 取下一个号
func (m *MySQL) Next() (id uint64, err error) {
	// 预编译
	var stmt sqlx.StmtSession
	stmt, err = m.conn.Prepare(sqlReplaceIntoStub)
	if err != nil {
		logx.Errorw("conn.prepare failed", logx.LogField{Key: "err", Value: err.Error()})
		return 0, err
	}
	defer stmt.Close()

	// 	执行
	var rest sql.Result
	rest, err = stmt.Exec()
	if err != nil {
		logx.Errorw("stmt.Exec() failed", logx.LogField{Key: "err", Value: err.Error()})
		return 0, err
	}
	var lastid int64
	lastid, err = rest.LastInsertId()
	if err != nil {
		logx.Errorw("rest.LastInsertId() failed", logx.LogField{Key: "err", Value: err.Error()})
		return 0, err
	}
	id = uint64(lastid)
	return id, nil
}
