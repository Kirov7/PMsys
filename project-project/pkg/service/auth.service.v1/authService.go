package auth_service_v1

import (
	"context"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	authRpc "github.com/Kirov7/project-grpc/project/auth"
	"github.com/Kirov7/project-project/internal/dao"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/database/transaction"
	"github.com/Kirov7/project-project/internal/domain"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/jinzhu/copier"
)

type AuthService struct {
	authRpc.UnimplementedAuthServiceServer
	cache             repo.Cache
	transaction       transaction.Transaction
	projectAuthDomain *domain.ProjectAuthDomain
}

func NewAuthService() *AuthService {
	return &AuthService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

func (a *AuthService) AuthList(ctx context.Context, msg *authRpc.AuthRequest) (*authRpc.ListAuthMessageResponse, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	listPage, total, err := a.projectAuthDomain.AuthListPage(organizationCode, msg.Page, msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var prList []*authRpc.ProjectAuth
	copier.Copy(&prList, listPage)
	return &authRpc.ListAuthMessageResponse{List: prList, Total: total}, nil
}

func (a *AuthService) Apply(ctx context.Context, msg *authRpc.AuthRequest) (*authRpc.ApplyResponse, error) {
	if msg.Action == "getnode" {
		//获取列表
		list, checkedList, err := a.projectAuthDomain.AllNodeAndAuth(msg.AuthId)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*authRpc.ProjectNodeMessage
		copier.Copy(&prList, list)
		return &authRpc.ApplyResponse{List: prList, CheckedList: checkedList}, nil
	}
	if msg.Action == "save" {
		// 先删除project_auth_node表,再新增, 需要使用事务
		nodes := msg.Nodes
		authId := msg.AuthId
		err := a.transaction.Action(func(conn database.DbConn) error {
			err := a.projectAuthDomain.Save(conn, authId, nodes)
			return err
		})
		if err != nil {
			return nil, errs.GrpcError(err.(*errs.BError))
		}
	}
	return &authRpc.ApplyResponse{}, nil
}

func (a *AuthService) AuthNodesByMemberId(ctx context.Context, msg *authRpc.AuthRequest) (*authRpc.AuthNodesResponse, error) {
	list, err := a.projectAuthDomain.AuthNodes(msg.MemberId)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	return &authRpc.AuthNodesResponse{List: list}, nil
}
