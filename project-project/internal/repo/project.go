package repo

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
)

type ProjectRepo interface {
	FindProjectByMemId(ctx context.Context, memId int64, condition string, page int64, size int64) ([]*data.ProjectAndMember, int64, error)
	FindCollectProjectByMemId(ctx context.Context, memId int64, page int64, size int64) ([]*data.ProjectAndMember, int64, error)
	SaveProject(ctx context.Context, conn database.DbConn, project *data.Project) error
	SaveProjectMember(ctx context.Context, conn database.DbConn, pm *data.ProjectMember) error
	FindProjectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (*data.ProjectAndMember, error)
	FindCollectByPidAndMemId(ctx context.Context, projectCode int64, memberId int64) (bool, error)
	UpdateDeletedProject(ctx context.Context, projectCode int64, deleted bool) error
	SaveProjectCollect(ctx context.Context, pc *data.ProjectCollection) error
	DeleteProjectCollect(ctx context.Context, projectCode int64, memberId int64) error
	UpdateProject(ctx context.Context, project *data.Project) error
	FindMemberByProjectCode(ctx context.Context, projectCode int64) ([]*data.ProjectMember, int64, error)
	FindProjectById(ctx context.Context, projectCode int64) (*data.Project, error)
	FindProjectByIds(ctx context.Context, pids []int64) ([]*data.Project, error)
}

type ProjectTemplateRepo interface {
	FindProjectTemplateSystem(ctx context.Context, page int64, size int64) ([]data.ProjectTemplate, int64, error)
	FindProjectTemplateCustom(ctx context.Context, memId int64, organizationCode int64, page int64, size int64) ([]data.ProjectTemplate, int64, error)
	FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]data.ProjectTemplate, int64, error)
}
