package project

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-api/pkg/model"
	"github.com/Kirov7/project-api/pkg/model/project"
	"github.com/Kirov7/project-api/pkg/model/source"
	"github.com/Kirov7/project-api/pkg/model/tasks"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-common/fs"
	"github.com/Kirov7/project-common/tms"
	taskRpc "github.com/Kirov7/project-grpc/project/task"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"

	"time"
)

type HandlerTask struct {
}

func NewHandlerTask() *HandlerTask {
	return &HandlerTask{}
}

func (t *HandlerTask) taskStages(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		MemberId:    c.GetInt64("memberId"),
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	stages, err := TaskServiceClient.TaskStages(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var resp []*tasks.TaskStagesResp
	copier.Copy(&resp, stages.List)
	if resp == nil {
		resp = []*tasks.TaskStagesResp{}
	}
	for _, v := range resp {
		v.TasksLoading = true  //任务加载状态
		v.FixedCreator = false //添加任务按钮定位
		v.ShowTaskCard = false //是否显示创建卡片
		v.Tasks = []int{}
		v.DoneTasks = []int{}
		v.UnDoneTasks = []int{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  resp,
		"total": stages.Total,
		"page":  page.Page,
	}))
}

func (t *HandlerTask) memberProjectList(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	page := &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		ProjectCode: projectCode,
		Page:        page.Page,
		PageSize:    page.PageSize,
	}
	memberResp, err := TaskServiceClient.MemberProjectList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var resp []*project.MemberProjectResp
	copier.Copy(&resp, memberResp.List)
	if resp == nil {
		resp = []*project.MemberProjectResp{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  resp,
		"total": memberResp.Total,
		"page":  page.Page,
	}))
}

func (t *HandlerTask) taskList(c *gin.Context) {
	result := &common.Result{}
	stageCode := c.PostForm("stageCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, err := TaskServiceClient.TaskList(ctx, &taskRpc.TaskRequest{StageCode: stageCode, MemberId: c.GetInt64("memberId")})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var taskDisplayList []*tasks.TaskDisplay
	copier.Copy(&taskDisplayList, list.List)
	// 返回给前端的数据不要为nil
	if taskDisplayList == nil {
		taskDisplayList = []*tasks.TaskDisplay{}
	}
	for _, v := range taskDisplayList {
		if v.Tags == nil {
			v.Tags = []int{}
		}
		if v.ChildCount == nil {
			v.ChildCount = []int{}
		}
	}
	c.JSON(http.StatusOK, result.Success(taskDisplayList))
}

func (t *HandlerTask) saveTask(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSaveReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		ProjectCode: req.ProjectCode,
		Name:        req.Name,
		StageCode:   req.StageCode,
		AssignTo:    req.AssignTo,
		MemberId:    c.GetInt64("memberId"),
	}
	taskMessage, err := TaskServiceClient.SaveTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	c.JSON(http.StatusOK, result.Success(td))
}

func (t *HandlerTask) taskSort(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskSortReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	msg := &taskRpc.TaskSortRequest{
		PreTaskCode:  req.PreTaskCode,
		NextTaskCode: req.NextTaskCode,
		ToStageCode:  req.ToStageCode,
	}
	_, err = TaskServiceClient.TaskSort(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (t *HandlerTask) myTaskList(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.MyTaskReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		MemberId: memberId,
		TaskType: int32(req.TaskType),
		Type:     int32(req.Type),
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	myTaskListResponse, err := TaskServiceClient.MyTaskList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var myTaskList []*tasks.MyTaskDisplay
	copier.Copy(&myTaskList, myTaskListResponse.List)
	if myTaskList == nil {
		myTaskList = []*tasks.MyTaskDisplay{}
	}
	for _, v := range myTaskList {
		v.ProjectInfo = tasks.ProjectInfo{
			Name: v.ProjectName,
			Code: v.ProjectCode,
		}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  myTaskList,
		"total": myTaskListResponse.Total,
	}))
}

func (t *HandlerTask) readTask(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	taskMessage, err := TaskServiceClient.ReadTask(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	td := &tasks.TaskDisplay{}
	copier.Copy(td, taskMessage)
	if td != nil {
		if td.Tags == nil {
			td.Tags = []int{}
		}
		if td.ChildCount == nil {
			td.ChildCount = []int{}
		}
	}
	c.JSON(200, result.Success(td))
}

func (t *HandlerTask) listTaskMember(c *gin.Context) {

	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	page := &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	taskMemberResponse, err := TaskServiceClient.ListTaskMember(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var tms []*tasks.TaskMember
	copier.Copy(&tms, taskMemberResponse.List)
	if tms == nil {
		tms = []*tasks.TaskMember{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskMemberResponse.Total,
		"page":  page.Page,
	}))

}

func (t *HandlerTask) taskLog(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.TaskLogReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode: req.TaskCode,
		MemberId: c.GetInt64("memberId"),
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
		All:      int32(req.All),
		Comment:  int32(req.Comment),
	}
	taskLogResponse, err := TaskServiceClient.TaskLog(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var tms []*tasks.ProjectLogDisplay
	copier.Copy(&tms, taskLogResponse.List)
	if tms == nil {
		tms = []*tasks.ProjectLogDisplay{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  tms,
		"total": taskLogResponse.Total,
		"page":  req.Page,
	}))
}

func (t *HandlerTask) taskWorkTimeList(c *gin.Context) {

	taskCode := c.PostForm("taskCode")
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode: taskCode,
		MemberId: c.GetInt64("memberId"),
	}
	taskWorkTimeResponse, err := TaskServiceClient.TaskWorkTimeList(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var tms []*tasks.TaskWorkTime
	copier.Copy(&tms, taskWorkTimeResponse.List)
	if tms == nil {
		tms = []*tasks.TaskWorkTime{}
	}
	c.JSON(http.StatusOK, result.Success(tms))

}

func (t *HandlerTask) saveTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	var req *tasks.SaveTaskWorkTimeReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode:  req.TaskCode,
		MemberId:  c.GetInt64("memberId"),
		Content:   req.Content,
		Num:       int32(req.Num),
		BeginTime: tms.ParseTime(req.BeginTime),
	}
	_, err = TaskServiceClient.SaveTaskWorkTime(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (t *HandlerTask) uploadFiles(c *gin.Context) {
	result := &common.Result{}
	req := tasks.UploadFileReq{}
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	multipartForm, err := c.MultipartForm()
	if err != nil {
		zap.L().Error("c.MultipartForm() err", zap.Error(err))
		return
	}
	file := multipartForm.File
	key := ""
	if req.TotalChunks == 1 {
		//代表不分片，直接上传
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if !fs.IsExist(path) {
			os.MkdirAll(path, os.ModePerm)
		}
		dst := path + "/" + req.Filename
		key = dst
		header := file["file"][0]
		err := c.SaveUploadedFile(header, dst)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
	}
	if req.TotalChunks > 1 {
		//分片上传 合起来即可
		path := "upload/" + req.ProjectCode + "/" + req.TaskCode + "/" + tms.FormatYMD(time.Now())
		if !fs.IsExist(path) {
			os.MkdirAll(path, os.ModePerm)
		}
		fileName := path + "/" + req.Identifier
		openFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(-999, err.Error()))
			return
		}
		open, err := file["file"][0].Open()
		defer open.Close()
		buf := make([]byte, req.CurrentChunkSize)
		open.Read(buf)
		openFile.Write(buf)
		openFile.Close()
		newpath := path + "/" + req.Filename
		key = newpath
		if req.TotalChunks == req.ChunkNumber {
			//最后一块 重命名文件名
			err := os.Rename(fileName, newpath)
			fmt.Println(err)
		}
	}
	//调用服务 存入file表
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	fileUrl := "http://localhost/" + key
	msg := &taskRpc.TaskFileReqMessage{
		TaskCode:         req.TaskCode,
		ProjectCode:      req.ProjectCode,
		OrganizationCode: c.GetString("organizationCode"),
		PathName:         key,
		FileName:         req.Filename,
		Size:             int64(req.TotalSize),
		Extension:        path.Ext(key),
		FileUrl:          fileUrl,
		FileType:         file["file"][0].Header.Get("Content-Type"),
		MemberId:         c.GetInt64("memberId"),
	}
	if req.TotalChunks == req.ChunkNumber {
		_, err = TaskServiceClient.SaveTaskFile(ctx, msg)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			return
		}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"file":        key,
		"hash":        "",
		"key":         key,
		"url":         "http://localhost/" + key,
		"projectName": req.ProjectName,
	}))
	return
}

func (t *HandlerTask) taskSources(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sources, err := TaskServiceClient.TaskSources(ctx, &taskRpc.TaskRequest{TaskCode: taskCode})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var slList []*source.SourceLink
	copier.Copy(&slList, sources.List)
	if slList == nil {
		slList = []*source.SourceLink{}
	}
	c.JSON(http.StatusOK, result.Success(slList))
}

func (t *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	req := tasks.CommentReq{}
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &taskRpc.TaskRequest{
		TaskCode:       req.TaskCode,
		CommentContent: req.Comment,
		Mentions:       req.Mentions,
		MemberId:       c.GetInt64("memberId"),
	}
	_, err = TaskServiceClient.CreateComment(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success(true))
}
