package project

import (
	"context"
	"github.com/Kirov7/project-api/pkg/model/department"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	departmentRpc "github.com/Kirov7/project-grpc/project/department"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"time"
)

type HandlerDepartment struct {
}

func NewDepartment() *HandlerDepartment {
	return &HandlerDepartment{}
}

func (d *HandlerDepartment) save(c *gin.Context) {
	result := &common.Result{}
	var req *department.DepartmentReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &departmentRpc.DepartmentRequest{
		Name:                 req.Name,
		DepartmentCode:       req.DepartmentCode,
		ParentDepartmentCode: req.ParentDepartmentCode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	departmentMessage, err := DepartmentServiceClient.Save(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var res = &department.Department{}
	copier.Copy(res, departmentMessage)
	c.JSON(http.StatusOK, result.Success(res))
}

func (d *HandlerDepartment) list(c *gin.Context) {
	result := &common.Result{}
	var req *department.DepartmentReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &departmentRpc.DepartmentRequest{
		Page:                 req.Page,
		PageSize:             req.PageSize,
		ParentDepartmentCode: req.Pcode,
		OrganizationCode:     c.GetString("organizationCode"),
	}
	listDepartmentMessage, err := DepartmentServiceClient.List(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*department.Department
	copier.Copy(&list, listDepartmentMessage.List)
	if list == nil {
		list = []*department.Department{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"total": listDepartmentMessage.Total,
		"page":  req.Page,
		"list":  list,
	}))
}

func (d *HandlerDepartment) read(c *gin.Context) {
	result := &common.Result{}
	departmentCode := c.PostForm("departmentCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &departmentRpc.DepartmentRequest{
		DepartmentCode:   departmentCode,
		OrganizationCode: c.GetString("organizationCode"),
	}
	departmentMessage, err := DepartmentServiceClient.Read(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var res = &department.Department{}
	copier.Copy(res, departmentMessage)
	c.JSON(http.StatusOK, result.Success(res))
}
