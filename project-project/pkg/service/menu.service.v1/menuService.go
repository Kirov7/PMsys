package menu_service_v1

import (
	"context"
	"github.com/Kirov7/project-common/errs"
	menuRpc "github.com/Kirov7/project-grpc/project/menu"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/database/transaction"
	"github.com/Kirov7/project-project/internal/domain"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type MenuService struct {
	menuRpc.UnimplementedMenuServiceServer
	cache       repo.Cache
	transaction transaction.Transaction
	menuDomain  *domain.MenuDomain
}

func NewMenuService() *MenuService {
	return &MenuService{
		cache:       dao.Rc,
		transaction: dao.NewTransaction(),
		menuDomain:  domain.NewMenuDomain(),
	}
}

func (m *MenuService) MenuList(context.Context, *menuRpc.MenuRequest) (*menuRpc.MenuListResponse, error) {
	treeList, err := m.menuDomain.MenuTreeList()
	if err != nil {
		zap.L().Error("MenuList error", zap.Error(err))
		return nil, errs.GrpcError(err)
	}
	var list []*menuRpc.MenuMessage
	copier.Copy(&list, treeList)
	return &menuRpc.MenuListResponse{List: list}, nil
}
