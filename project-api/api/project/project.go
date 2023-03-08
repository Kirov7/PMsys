package project

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-api/pkg/model"
	"github.com/Kirov7/project-api/pkg/model/menu"
	"github.com/Kirov7/project-api/pkg/model/project"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	projectRpc "github.com/Kirov7/project-grpc/project/project"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"strconv"
	"time"
)

type HandlerProject struct {
}

func NewHandlerProject() *HandlerProject {
	return &HandlerProject{}
}

func (p *HandlerProject) index(ctx *gin.Context) {
	resp := &common.Result{}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	indexRpcReq := &projectRpc.IndexRequest{}
	indexRpcResp, err := ProjectServiceClient.Index(c, indexRpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, resp.Fail(code, msg))
		return
	}
	menus := indexRpcResp.Menus
	var ms []*menu.Menu
	copier.Copy(&ms, menus)
	// 4. 返回结果
	ctx.JSON(http.StatusOK, resp.Success(ms))

}

func (p *HandlerProject) projectList(ctx *gin.Context) {
	// 1. 获取参数
	result := &common.Result{}
	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	memberId := ctx.GetInt64("memberId")
	memberName := ctx.GetString("memberName")
	var page = &model.Page{}
	page.Bind(ctx)
	selectBy := ctx.PostForm("selectBy")
	findProjectByMemIdRpcReq := &projectRpc.ProjectRequest{
		MemberId:   memberId,
		MemberName: memberName,
		Page:       page.Page,
		PageSize:   page.PageSize,
		SelectBy:   selectBy,
	}
	findProjectByMemIdRpcResp, err := ProjectServiceClient.FindProjectByMemId(c, findProjectByMemIdRpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var pam []*project.ProjectAndMember
	copier.Copy(&pam, findProjectByMemIdRpcResp.Pm)
	if pam == nil {
		pam = []*project.ProjectAndMember{}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pam,
		"total": findProjectByMemIdRpcResp.Total,
	}))
}

func (p *HandlerProject) projectTemplate(ctx *gin.Context) {
	result := &common.Result{}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberId := ctx.GetInt64("memberId")
	memberName := ctx.GetString("memberName")
	organizationCode := ctx.GetString("organizationCode")
	var page = &model.Page{}
	page.Bind(ctx)
	viewTypeStr := ctx.PostForm("viewType")
	viewType, _ := strconv.ParseInt(viewTypeStr, 10, 64)
	projectTemplateRsp, err := ProjectServiceClient.FindProjectTemplate(c,
		&projectRpc.ProjectRequest{
			MemberId:         memberId,
			MemberName:       memberName,
			OrganizationCode: organizationCode,
			Page:             page.Page,
			PageSize:         page.PageSize,
			ViewType:         int32(viewType)})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var pts []*project.ProjectTemplate
	copier.Copy(&pts, projectTemplateRsp.Ptm)
	if pts == nil {
		pts = []*project.ProjectTemplate{}
	}
	for _, pt := range pts {
		if pt.TaskStages == nil {
			pt.TaskStages = []*project.TaskStagesOnlyName{}
		}
	}
	ctx.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pts,
		"total": projectTemplateRsp.Total,
	}))
}

func (p *HandlerProject) projectSave(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	memberId := c.GetInt64("memberId")
	organizationCode := c.GetString("organizationCode")
	var req *project.SaveProjectRequest
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusOK, "参数错误"))
		return
	}
	saveProjectRpcReq := &projectRpc.ProjectRequest{
		MemberId:         memberId,
		Name:             req.Name,
		OrganizationCode: organizationCode,
		Description:      req.Description,
		TemplateCode:     req.TemplateCode,
		Id:               int64(req.Id)}
	saveProjectRpcResp, err := ProjectServiceClient.SaveProject(ctx, saveProjectRpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	rsp := &project.SaveProject{
		Id:               saveProjectRpcResp.Id,
		Cover:            saveProjectRpcResp.Cover,
		Name:             saveProjectRpcResp.Name,
		Description:      saveProjectRpcResp.Description,
		Code:             saveProjectRpcResp.Code,
		CreateTime:       saveProjectRpcResp.CreateTime,
		TaskBoardTheme:   saveProjectRpcResp.TaskBoardTheme,
		OrganizationCode: saveProjectRpcResp.OrganizationCode,
	}
	err = copier.Copy(&rsp, saveProjectRpcResp)
	if err != nil {
		fmt.Println("copy err", err)
		return
	}
	c.JSON(http.StatusOK, result.Success(rsp))
}

func (p *HandlerProject) readProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	detail, err := ProjectServiceClient.FindProjectDetail(ctx, &projectRpc.ProjectRequest{ProjectCode: projectCode, MemberId: memberId})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	pd := &project.ProjectDetail{}
	copier.Copy(pd, detail)
	c.JSON(http.StatusOK, result.Success(pd))
}

func (p *HandlerProject) recycleProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &projectRpc.UpdateDeletedRequest{ProjectCode: projectCode, Deleted: true})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) recoveryProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &projectRpc.UpdateDeletedRequest{ProjectCode: projectCode, Deleted: false})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) collectProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	collectType := c.PostForm("type")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateCollectProject(ctx, &projectRpc.UpdateCollectRequest{ProjectCode: projectCode, CollectType: collectType, MemberId: memberId})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) editProject(c *gin.Context) {
	result := &common.Result{}
	var req *project.ProjectEdit
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	editProjectRpcReq := &projectRpc.UpdateProjectRequest{}
	copier.Copy(editProjectRpcReq, req)
	editProjectRpcReq.MemberId = memberId
	_, err = ProjectServiceClient.UpdateProject(ctx, editProjectRpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) getLogBySelfProject(c *gin.Context) {
	result := &common.Result{}
	var page = &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &projectRpc.ProjectRequest{
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	projectLogResponse, err := ProjectServiceClient.GetLogBySelfProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*project.ProjectLog
	copier.Copy(&list, projectLogResponse.List)
	if list == nil {
		list = []*project.ProjectLog{}
	}
	c.JSON(http.StatusOK, result.Success(list))
}

func (p *HandlerProject) nodeList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := ProjectServiceClient.NodeList(ctx, &projectRpc.ProjectRequest{})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*project.ProjectNodeTree
	copier.Copy(&list, response.Nodes)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"nodes": list,
	}))
}

func (p *HandlerProject) FindProjectByMemberId(memberId int64, projectCode string, taskCode string) (*project.Project, bool, bool, *errs.BError) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &projectRpc.ProjectRequest{
		MemberId:    memberId,
		ProjectCode: projectCode,
		TaskCode:    taskCode,
	}
	projectResponse, err := ProjectServiceClient.FindProjectByMemberId(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		return nil, false, false, errs.NewError(errs.ErrorCode(code), msg)
	}
	if projectResponse.Project == nil {
		return nil, false, false, nil
	}
	pr := &project.Project{}
	copier.Copy(pr, projectResponse.Project)
	return pr, true, projectResponse.IsOwner, nil
}
