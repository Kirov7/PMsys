package domain

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
)

type ProjectAuthNodeDomain struct {
	projectAuthNodeRepo repo.ProjectAuthNodeRepo
}

func NewProjectAuthNodeDomain() *ProjectAuthNodeDomain {
	return &ProjectAuthNodeDomain{
		projectAuthNodeRepo: dao.NewProjectAuthNodeDao(),
	}
}

func (d ProjectAuthNodeDomain) FindNodeStringList(authId int64) ([]string, *errs.BError) {
	list, err := d.projectAuthNodeRepo.FindNodeStringList(context.Background(), authId)
	if err != nil {
		return nil, model.DBError
	}
	return list, nil
}

func (d ProjectAuthNodeDomain) Save(conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeRepo.DeleteByAuthId(context.Background(), conn, authId)
	if err != nil {
		return model.DBError
	}
	err = d.projectAuthNodeRepo.Save(context.Background(), conn, authId, nodes)
	if err != nil {
		return model.DBError
	}
	return nil
}
