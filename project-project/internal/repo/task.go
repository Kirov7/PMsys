package repo

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
)

type TaskStagesTemplateRepo interface {
	FindInProTemIds(ctx context.Context, id []int) ([]data.TaskStagesTemplate, error)
	FindByProjectTPLId(ctx context.Context, projectTemplateCode int) (list []*data.TaskStagesTemplate, err error)
}

type TaskStagesRepo interface {
	SaveTaskStages(ctx context.Context, conn database.DbConn, ts *data.TaskStages) error
	FindByProjectCode(ctx context.Context, projectCode int64, page int64, size int64) (taskList []*data.TaskStages, total int64, err error)
	FindById(ctx context.Context, id int) (*data.TaskStages, error)
}

type TaskRepo interface {
	FindTaskByStageCode(ctx context.Context, stageCode int) ([]*data.Task, error)
	FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberCode int64) (*data.TaskMember, error)
	FindTaskMaxIdNum(ctx context.Context, projectCode int64) (int64, error)
	FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (int64, error)
	SaveTask(ctx context.Context, conn database.DbConn, ts *data.Task) error
	SaveTaskMember(ctx context.Context, conn database.DbConn, tm *data.TaskMember) error
	FindTaskById(ctx context.Context, taskCode int64) (*data.Task, error)
	UpdateTaskSort(ctx context.Context, conn database.DbConn, ts *data.Task) error
	FindTaskByStageCodeLtSort(ctx context.Context, stageCode int, sort int) (*data.Task, error)
	FindTaskByAssignTo(ctx context.Context, memberId int64, done int, page int64, pageSize int64) ([]*data.Task, int64, error)
	FindTaskByMemberCode(ctx context.Context, memberId int64, done int, page int64, pageSize int64) (tList []*data.Task, total int64, err error)
	FindTaskByCreateBy(ctx context.Context, memberId int64, done int, page int64, pageSize int64) (tList []*data.Task, total int64, err error)
	FindTaskMemberPage(ctx context.Context, taskId int64, page int64, size int64) (tList []*data.TaskMember, total int64, err error)
	FindTaskByIds(ctx context.Context, taskIdList []int64) (list []*data.Task, err error)
}
