package dao

import (
	"context"
	"github.com/Kirov7/project-user/internal/data/organization"
	database "github.com/Kirov7/project-user/internal/database"
	"github.com/Kirov7/project-user/internal/database/gorms"
)

type OrganizationDao struct {
	conn *gorms.GormConn
}

func NewOrganizationDao() *OrganizationDao {
	return &OrganizationDao{
		conn: gorms.New(),
	}
}

func (o *OrganizationDao) FindOrganizationByMemId(ctx context.Context, memId int64) ([]organization.Organization, error) {
	var orgs []organization.Organization
	err := o.conn.Session(ctx).Where("member_id=?", memId).Find(&orgs).Error
	return orgs, err
}

func (o *OrganizationDao) SaveOrganization(conn database.DbConn, ctx context.Context, org *organization.Organization) error {
	o.conn = conn.(*gorms.GormConn)
	err := o.conn.Tx(ctx).Create(org).Error
	return err
}
