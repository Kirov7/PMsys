package domain

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
)

type MenuDomain struct {
	menuRepo repo.MenuRepo
}

func (d *MenuDomain) MenuTreeList() ([]*data.ProjectMenuChild, *errs.BError) {
	menus, err := d.menuRepo.FindMenus(context.Background())
	if err != nil {
		return nil, model.DBError
	}
	menuChildren := data.CovertChild(menus)
	return menuChildren, nil
}

func NewMenuDomain() *MenuDomain {
	return &MenuDomain{
		menuRepo: dao.NewMenuDao(),
	}
}
