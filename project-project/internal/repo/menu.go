package repo

import (
	"context"
	"github.com/Kirov7/project-project/internal/data"
)

type MenuRepo interface {
	FindMenus(ctx context.Context) ([]*data.ProjectMenu, error)
}
