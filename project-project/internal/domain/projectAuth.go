package domain

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type ProjectAuthDomain struct {
	projectAuthRepo       repo.ProjectAuthRepo
	projectNodeDomain     *ProjectNodeDomain
	projectAuthNodeDomain *ProjectAuthNodeDomain
	accountDomain         *AccountDomain
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		projectAuthRepo:       dao.NewProjectAuthDao(),
		projectNodeDomain:     NewProjectNodeDomain(),
		projectAuthNodeDomain: NewProjectAuthNodeDomain(),
		accountDomain:         NewAccountDomain(),
	}
}

func (d *ProjectAuthDomain) AuthList(orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, err := d.projectAuthRepo.FindAuthList(c, orgCode)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, nil
}

func (d *ProjectAuthDomain) AuthListPage(orgCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, total, err := d.projectAuthRepo.FindAuthListPage(c, orgCode, page, pageSize)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, 0, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, total, nil
}

func (d *ProjectAuthDomain) AllNodeAndAuth(authId int64) ([]*data.ProjectNodeAuthTree, []string, *errs.BError) {
	treeList, err := d.projectNodeDomain.AllNodeList()
	if err != nil {
		return nil, nil, err
	}
	authNodeList, dbErr := d.projectAuthNodeDomain.FindNodeStringList(authId)
	if dbErr != nil {
		return nil, nil, model.DBError
	}
	list := data.ToAuthNodeTreeList(treeList, authNodeList)
	return list, authNodeList, nil
}

func (d *ProjectAuthDomain) Save(conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeDomain.Save(conn, authId, nodes)
	if err != nil {
		return err
	}
	return nil

}

func (d *ProjectAuthDomain) AuthNodes(memberId int64) ([]string, *errs.BError) {
	account, err := d.accountDomain.FindAccount(memberId)
	if err != nil {
		return nil, err
	}
	//c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//defer cancel()
	authorize := account.Authorize
	authId, _ := strconv.ParseInt(authorize, 10, 64)
	authNodeList, dbErr := d.projectAuthNodeDomain.FindNodeStringList(authId)
	if dbErr != nil {
		return nil, model.DBError
	}
	return authNodeList, nil
}
