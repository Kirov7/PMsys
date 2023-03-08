package domain

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
)

type TaskDomain struct {
	taskRepo repo.TaskRepo
}

func NewTaskDomain() *TaskDomain {
	return &TaskDomain{
		taskRepo: dao.NewTaskDao(),
	}
}

func (d *TaskDomain) FindProjectIdByTaskId(taskId int64) (int64, bool, *errs.BError) {
	task, err := d.taskRepo.FindTaskById(context.Background(), taskId)
	if err != nil {
		return 0, false, model.DBError
	}
	if task == nil {
		return 0, false, nil
	}
	return task.ProjectCode, true, nil
}
