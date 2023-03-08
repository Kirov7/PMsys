package project

import (
	"context"
	"encoding/json"
	"github.com/Kirov7/project-api/pkg/model"
	"github.com/Kirov7/project-api/pkg/model/auth"
	"github.com/Kirov7/project-api/pkg/model/project"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	authRpc "github.com/Kirov7/project-grpc/project/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"time"
)

type HandlerAuth struct {
}

func NewAuth() *HandlerAuth {
	return &HandlerAuth{}
}

func (a *HandlerAuth) authList(c *gin.Context) {
	result := &common.Result{}
	organizationCode := c.GetString("organizationCode")
	var page = &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &authRpc.AuthRequest{
		OrganizationCode: organizationCode,
		Page:             page.Page,
		PageSize:         page.PageSize,
	}
	response, err := AuthServiceClient.AuthList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var authList []*auth.ProjectAuth
	copier.Copy(&authList, response.List)
	if authList == nil {
		authList = []*auth.ProjectAuth{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total": response.Total,
		"list":  authList,
		"page":  page.Page,
	}))
}

func (a *HandlerAuth) apply(c *gin.Context) {
	result := &common.Result{}
	var req *auth.ProjectAuthReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	var nodes []string
	if req.Nodes != "" {
		err := json.Unmarshal([]byte(req.Nodes), &nodes)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
			return
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &authRpc.AuthRequest{
		Action: req.Action,
		AuthId: req.Id,
		Nodes:  nodes,
	}
	applyResponse, err := AuthServiceClient.Apply(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*project.ProjectNodeAuthTree
	copier.Copy(&list, applyResponse.List)
	var checkedList []string
	copier.Copy(&checkedList, applyResponse.CheckedList)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":        list,
		"checkedList": checkedList,
	}))
}

func (a *HandlerAuth) GetAuthNodes(c *gin.Context) ([]string, error) {
	memberId := c.GetInt64("memberId")
	msg := &authRpc.AuthRequest{
		MemberId: memberId,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := AuthServiceClient.AuthNodesByMemberId(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		return nil, errs.NewError(errs.ErrorCode(code), msg)
	}
	return response.List, err
}
