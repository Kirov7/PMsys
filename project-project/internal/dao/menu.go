package dao

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/database/gorms"
)

type MenuDao struct {
	conn *gorms.GormConn
}

func NewMenuDao() *MenuDao {
	return &MenuDao{
		conn: gorms.New(),
	}
}

func (m *MenuDao) FindMenus(ctx context.Context) (pms []*data.ProjectMenu, err error) {
	session := m.conn.Session(ctx)
	err = session.Order("pid, sort asc, id asc").Find(&pms).Error
	return pms, err
}
