package transaction

import (
	database "github.com/Kirov7/project-user/internal/database"
)

// Transaction 事务操作 需要注入数据库连接
type Transaction interface {
	Action(func(conn database.DbConn) error) error
}
