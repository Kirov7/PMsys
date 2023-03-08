package project

import (
	"context"
	"github.com/Kirov7/project-api/pkg/model/account"
	"github.com/Kirov7/project-api/pkg/model/auth"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	accountRpc "github.com/Kirov7/project-grpc/project/account"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"time"
)

type HandlerAccount struct {
}

func NewAccount() *HandlerAccount {
	return &HandlerAccount{}
}

func (a *HandlerAccount) account(c *gin.Context) {
	result := &common.Result{}
	var req account.AccountReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &accountRpc.AccountRequest{
		MemberId:         c.GetInt64("memberId"),
		OrganizationCode: c.GetString("organizationCode"),
		Page:             int64(req.Page),
		PageSize:         int64(req.PageSize),
		SearchType:       int32(req.SearchType),
		DepartmentCode:   req.DepartmentCode,
	}
	response, err := AccountServiceClient.Account(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*account.MemberAccount
	copier.Copy(&list, response.AccountList)
	if list == nil {
		list = []*account.MemberAccount{}
	}
	var authList []*auth.ProjectAuth
	copier.Copy(&authList, response.AuthList)
	if authList == nil {
		authList = []*auth.ProjectAuth{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total":    response.Total,
		"page":     req.Page,
		"list":     list,
		"authList": authList,
	}))
}
