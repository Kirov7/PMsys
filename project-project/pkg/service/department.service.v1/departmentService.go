package department_service_v1

import (
	"context"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	departmentRpc "github.com/Kirov7/project-grpc/project/department"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/database/transaction"
	"github.com/Kirov7/project-project/internal/domain"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/jinzhu/copier"
)

type DepartmentService struct {
	departmentRpc.UnimplementedDepartmentServiceServer
	cache            repo.Cache
	transaction      transaction.Transaction
	departmentDomain *domain.DepartmentDomain
}

func NewDepartmentService() *DepartmentService {
	return &DepartmentService{
		cache:            dao.Rc,
		transaction:      dao.NewTransaction(),
		departmentDomain: domain.NewDepartmentDomain(),
	}
}

func (d *DepartmentService) Save(ctx context.Context, msg *departmentRpc.DepartmentRequest) (*departmentRpc.DepartmentMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	var departmentCode int64
	if msg.DepartmentCode != "" {
		departmentCode = encrypts.DecryptNoErr(msg.DepartmentCode)
	}
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode = encrypts.DecryptNoErr(msg.ParentDepartmentCode)
	}
	dp, err := d.departmentDomain.Save(organizationCode, departmentCode, parentDepartmentCode, msg.Name)
	if err != nil {
		return &departmentRpc.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &departmentRpc.DepartmentMessage{}
	copier.Copy(res, dp)
	return res, nil
}

func (d *DepartmentService) List(ctx context.Context, msg *departmentRpc.DepartmentRequest) (*departmentRpc.ListDepartmentMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode = encrypts.DecryptNoErr(msg.ParentDepartmentCode)
	}
	dps, total, err := d.departmentDomain.List(organizationCode, parentDepartmentCode, msg.Page, msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var list []*departmentRpc.DepartmentMessage
	copier.Copy(&list, dps)
	return &departmentRpc.ListDepartmentMessage{List: list, Total: total}, nil
}

func (d *DepartmentService) Read(ctx context.Context, msg *departmentRpc.DepartmentRequest) (*departmentRpc.DepartmentMessage, error) {
	//organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	departmentCode := encrypts.DecryptNoErr(msg.DepartmentCode)
	dp, err := d.departmentDomain.FindDepartmentById(departmentCode)
	if err != nil {
		return &departmentRpc.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &departmentRpc.DepartmentMessage{}
	copier.Copy(res, dp.ToDisplay())
	return res, nil
}
