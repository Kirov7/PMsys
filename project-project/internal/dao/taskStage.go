package dao

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/database/gorms"
)

type TaskStagesDao struct {
	conn *gorms.GormConn
}

func NewTaskStagesDao() *TaskStagesDao {
	return &TaskStagesDao{
		conn: gorms.New(),
	}
}

func (t *TaskStagesDao) SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error {
	t.conn = conn.(*gorms.GormConn)
	return t.conn.Tx(ctx).Save(&ts).Error
}

func (t *TaskStagesDao) FindByProjectCode(ctx context.Context, projectCode int64, page int64, size int64) (taskList []*data.TaskStages, total int64, err error) {
	session := t.conn.Session(ctx)
	var stages []*data.TaskStages
	err = session.Model(&data.TaskStages{}).Where("project_code=? and deleted=?", projectCode, 0).Order("sort asc").Limit(int(size)).Offset(int((page - 1) * size)).Find(&stages).Error
	err = session.Model(&data.TaskStages{}).Where("project_code=?", projectCode).Count(&total).Error
	return stages, total, err
}

func (t *TaskStagesDao) FindById(ctx context.Context, id int) (ts *data.TaskStages, err error) {
	err = t.conn.Session(ctx).Where("id=?", id).Find(&ts).Error
	return
}
