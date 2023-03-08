package dao

import (
	database "github.com/Kirov7/project-user/internal/database"
	"github.com/Kirov7/project-user/internal/database/gorms"
)

type TransactionImpl struct {
	conn database.DbConn
}

func NewTransaction() *TransactionImpl {
	return &TransactionImpl{
		conn: gorms.NewTransaction(),
	}
}

func (t *TransactionImpl) Action(f func(conn database.DbConn) error) error {
	t.conn.Begin()
	err := f(t.conn)
	if err != nil {
		t.conn.Rollback()
		return err
	}
	t.conn.Commit()
	return nil
}
