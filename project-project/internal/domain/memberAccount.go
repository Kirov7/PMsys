package domain

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/data"
	"github.com/Kirov7/project-project/internal/repo"
	"github.com/Kirov7/project-project/pkg/model"
	"time"
)

type AccountDomain struct {
	memberAccountRepo repo.AccountRepo
	departmentDomain  *DepartmentDomain
	userRpcDomain     *UserRpcDomain
}

func NewAccountDomain() *AccountDomain {
	return &AccountDomain{
		memberAccountRepo: dao.NewMemberAccountDao(),
		userRpcDomain:     NewUserRpcDomain(),
		departmentDomain:  NewDepartmentDomain(),
	}
}

func (d *AccountDomain) AccountList(organizationCode string, memberId int64, page int64, pageSize int64, departmentCode string, searchType int32) ([]*data.MemberAccountDisplay, int64, *errs.BError) {
	condition := ""
	organizationCodeId := encrypts.DecryptNoErr(organizationCode)
	departmentCodeId := encrypts.DecryptNoErr(departmentCode)
	switch searchType {
	case model.Active:
		condition = "status = 1"
	case model.System:
		condition = "department_code = NULL"
	case model.Baned:
		condition = "status = 0"
	case model.ActiveAndCurDept:
		condition = fmt.Sprintf("status = 1 and department_code = %d", departmentCodeId)
	default:
		condition = "status = 1"
	}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, total, err := d.memberAccountRepo.FindList(c, condition, organizationCodeId, departmentCodeId, page, pageSize)
	if err != nil {
		return nil, 0, model.DBError
	}
	var dList []*data.MemberAccountDisplay
	for _, v := range list {
		display := v.ToDisplay()
		memberInfo, _ := d.userRpcDomain.MemberInfo(c, v.MemberCode)
		display.Avatar = memberInfo.Avatar
		if v.DepartmentCode > 0 {
			department, err := d.departmentDomain.FindDepartmentById(v.DepartmentCode)
			if err != nil {
				return nil, 0, err
			}
			display.Departments = department.Name
		}
		dList = append(dList, display)
	}
	return dList, total, nil
}

func (d *AccountDomain) FindAccount(memId int64) (*data.MemberAccount, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberAccount, err := d.memberAccountRepo.FindByMemberId(c, memId)
	if err != nil {
		return nil, model.DBError
	}
	return memberAccount, nil
}
