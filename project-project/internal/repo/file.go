package repo

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
)

type FileRepo interface {
	Save(ctx context.Context, file *data.File) error
	FindByIds(background context.Context, ids []int64) (list []*data.File, err error)
}
