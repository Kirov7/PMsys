package dao

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/database/gorms"
)

type TaskStagesTemplateDao struct {
	conn *gorms.GormConn
}

func NewTaskStagesTemplateDao() *TaskStagesTemplateDao {
	return &TaskStagesTemplateDao{
		conn: gorms.New(),
	}
}

func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]data.TaskStagesTemplate, error) {
	var tsts []data.TaskStagesTemplate
	session := t.conn.Session(ctx)
	err := session.Model(&data.TaskStagesTemplate{}).Where("project_template_code in ?", ids).Find(&tsts).Error
	return tsts, err
}

func (t *TaskStagesTemplateDao) FindByProjectTPLId(ctx context.Context, projectTemplateCode int) (list []*data.TaskStagesTemplate, err error) {
	session := t.conn.Session(ctx)
	err = session.Model(&data.TaskStagesTemplate{}).Where("project_template_code=?", projectTemplateCode).Order("sort desc, id asc").Find(&list).Error
	return list, err
}
