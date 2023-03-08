package dao

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/database/gorms"
)

type ProjectAuthDao struct {
	conn *gorms.GormConn
}

func (p *ProjectAuthDao) FindAuthList(ctx context.Context, orgCode int64) (list []*data.ProjectAuth, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).Where("organization_code=? and status=1", orgCode).Find(&list).Error
	return
}

func NewProjectAuthDao() *ProjectAuthDao {
	return &ProjectAuthDao{
		conn: gorms.New(),
	}
}

func (p *ProjectAuthDao) FindAuthListPage(ctx context.Context, orgCode int64, page int64, pageSize int64) (list []*data.ProjectAuth, total int64, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Limit(int(pageSize)).
		Offset(int((page - 1) * pageSize)).
		Find(&list).Error
	err = session.Model(&data.ProjectAuth{}).
		Where("organization_code=?", orgCode).
		Count(&total).Error
	return
}
