package repo

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
)

type ProjectNodeRepo interface {
	FindAll(ctx context.Context) (list []*data.ProjectNode, err error)
}
