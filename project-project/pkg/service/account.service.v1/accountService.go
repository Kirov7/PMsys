package account_service_v1

import (
	"context"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	accountRpc "github.com/Kirov7/project-grpc/project/account"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/database/transaction"
	"github.com/Kirov7/project-project/internal/domain"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/jinzhu/copier"
)

type AccountService struct {
	accountRpc.UnimplementedAccountServiceServer
	cache             repo.Cache
	transaction       transaction.Transaction
	accountDomain     *domain.AccountDomain
	projectAuthDomain *domain.ProjectAuthDomain
}

func NewAccountService() *AccountService {
	return &AccountService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		accountDomain:     domain.NewAccountDomain(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

func (a *AccountService) Account(c context.Context, msg *accountRpc.AccountRequest) (*accountRpc.AccountResponse, error) {
	// 去account表查询account
	accountList, total, err := a.accountDomain.AccountList(msg.OrganizationCode, msg.MemberId, msg.Page, msg.PageSize, msg.DepartmentCode, msg.SearchType)
	if err != nil {
		return &accountRpc.AccountResponse{}, errs.GrpcError(err)
	}
	// 再去auth表查询authList
	authList, err := a.projectAuthDomain.AuthList(encrypts.DecryptNoErr(msg.OrganizationCode))
	if err != nil {
		return &accountRpc.AccountResponse{}, errs.GrpcError(err)
	}
	var maList []*accountRpc.MemberAccount
	copier.Copy(&maList, accountList)
	var prList []*accountRpc.ProjectAuth
	copier.Copy(&prList, authList)
	return &accountRpc.AccountResponse{
		AccountList: maList,
		AuthList:    prList,
		Total:       total,
	}, nil
}
