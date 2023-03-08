package transaction

import (
	database "github.com/Kirov7/project-project/internal/database"
)

// Transaction 事务操作 需要注入数据库连接
type Transaction interface {
	Action(f func(conn database.DbConn) error) error
}
