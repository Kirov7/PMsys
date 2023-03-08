package dao

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/database/gorms"
)

type ProjectNodeDao struct {
	conn *gorms.GormConn
}

func NewProjectNodeDao() *ProjectNodeDao {
	return &ProjectNodeDao{
		conn: gorms.New(),
	}
}

func (m *ProjectNodeDao) FindAll(ctx context.Context) (pms []*data.ProjectNode, err error) {
	session := m.conn.Session(ctx)
	err = session.Model(&data.ProjectNode{}).Find(&pms).Error
	return
}
