package domain

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
)

type ProjectNodeDomain struct {
	projectNodeRepo repo.ProjectNodeRepo
}

func NewProjectNodeDomain() *ProjectNodeDomain {
	return &ProjectNodeDomain{
		projectNodeRepo: dao.NewProjectNodeDao(),
	}
}

func (d *ProjectNodeDomain) TreeList() ([]*data.ProjectNodeTree, *errs.BError) {
	nodes, err := d.projectNodeRepo.FindAll(context.Background())
	if err != nil {
		return nil, model.DBError
	}
	treeList := data.ToNodeTreeList(nodes)
	return treeList, nil
}

func (d *ProjectNodeDomain) AllNodeList() ([]*data.ProjectNode, *errs.BError) {
	nodes, err := d.projectNodeRepo.FindAll(context.Background())
	if err != nil {
		return nil, model.DBError
	}
	return nodes, nil
}
