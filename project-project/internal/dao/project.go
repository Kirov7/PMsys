package dao

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/database/gorms"
	"gorm.io/gorm"
)

type ProjectDao struct {
	conn *gorms.GormConn
}

func NewProjectDao() *ProjectDao {
	return &ProjectDao{conn: gorms.New()}
}

func (p *ProjectDao) FindProjectByMemId(ctx context.Context, memId int64, condition string, page int64, size int64) ([]*data.ProjectAndMember, int64, error) {
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select a.*,b.project_code,b.member_code,b.join_time,b.is_owner,b.authorize from project a, project_member b where a.id=b.project_code and b.member_code=? %s order by sort limit ?,?", condition)
	db := session.Raw(sql, memId, index, size)
	var mp []*data.ProjectAndMember
	err := db.Scan(&mp).Error
	if err != nil {
		return mp, 0, err
	}
	var total int64
	//session.Model(&project.ProjectAndMember{}).Where("member_code=?", memId).Count(&total)
	query := fmt.Sprintf("select count(*) from project a, project_member b where a.id = b.project_code and b.member_code=? %s", condition)
	tx := session.Raw(query, memId)
	err = tx.Scan(&total).Error
	return mp, total, err
}

func (p ProjectDao) FindCollectProjectByMemId(ctx context.Context, memId int64, page int64, size int64) ([]*data.ProjectAndMember, int64, error) {
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	sql := fmt.Sprintf("select * from project where id in (select project_code from project_collection where member_code = ?) order by sort limit ?,?")
	db := session.Raw(sql, memId, index, size)
	var mp []*data.ProjectAndMember
	err := db.Scan(&mp).Error
	if err != nil {
		return mp, 0, err
	}

	var total int64
	//session.Model(&project.ProjectAndMember{}).Where("member_code=?", memId).Count(&total)
	query := fmt.Sprintf("member_code=?")
	err = session.Model(&data.ProjectCollection{}).Where(query, memId).Count(&total).Error
	return mp, total, err
}

func (p *ProjectDao) SaveProject(ctx context.Context, conn database.DbConn, project *data.Project) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tx(ctx).Save(&project).Error
}

func (p *ProjectDao) SaveProjectMember(ctx context.Context, conn database.DbConn, pm *data.ProjectMember) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tx(ctx).Save(&pm).Error
}

func (p *ProjectDao) FindProjectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (*data.ProjectAndMember, error) {
	var pm *data.ProjectAndMember
	session := p.conn.Session(ctx)
	sql := fmt.Sprintf("select a.*,b.project_code,b.member_code,b.join_time,b.is_owner,b.authorize from project a, project_member b where a.id=b.project_code and b.member_code=? and b.project_code=? limit 1")
	raw := session.Raw(sql, memberId, projectCode)
	err := raw.Scan(&pm).Error
	return pm, err
}

func (p *ProjectDao) FindCollectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (bool, error) {
	var count int64
	session := p.conn.Session(ctx)
	sql := fmt.Sprintf("select count(*) from project_collection where member_code=? and project_code=?")
	raw := session.Raw(sql, memberId, projectCode)
	err := raw.Scan(&count).Error
	return count > 0, err
}

func (p *ProjectDao) UpdateDeletedProject(ctx context.Context, projectCode int64, deleted bool) error {
	var err error
	session := p.conn.Session(ctx)
	if deleted {
		err = session.Model(&data.Project{}).Where("id=?", projectCode).Update("deleted", 1).Error
	} else {
		err = session.Model(&data.Project{}).Where("id=?", projectCode).Update("deleted", 0).Error
	}
	return err
}

func (p *ProjectDao) SaveProjectCollect(ctx context.Context, pc *data.ProjectCollection) error {
	return p.conn.Session(ctx).Save(&pc).Error
}

func (p *ProjectDao) DeleteProjectCollect(ctx context.Context, projectCode int64, memberId int64) error {
	return p.conn.Session(ctx).Where("member_code=? and project_code=?", memberId, projectCode).Delete(&data.ProjectCollection{}).Error
}

func (p *ProjectDao) UpdateProject(ctx context.Context, project *data.Project) error {
	return p.conn.Session(ctx).Updates(&project).Error
}

func (p *ProjectDao) FindMemberByProjectCode(ctx context.Context, projectCode int64) (list []*data.ProjectMember, total int64, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectMember{}).Where("project_code=?", projectCode).Find(&list).Error
	if err != nil {
		return
	}
	err = session.Model(&data.ProjectMember{}).Where("project_code=?", projectCode).Count(&total).Error
	return
}

func (p *ProjectDao) FindProjectById(ctx context.Context, projectCode int64) (pj *data.Project, err error) {
	err = p.conn.Session(ctx).Where("id=?", projectCode).Find(&pj).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return
}

func (p *ProjectDao) FindProjectByIds(ctx context.Context, pids []int64) (list []*data.Project, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.Project{}).Where("id in (?)", pids).Find(&list).Error
	return
}
