package project_service_v1

import (
	"context"
	"fmt"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-common/tms"
	projectRpc "github.com/Kirov7/project-grpc/project/project"
	"github.com/Kirov7/project-grpc/user/login"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/data"
	database "github.com/Kirov7/project-project/internal/database"
	"github.com/Kirov7/project-project/internal/database/transaction"
	"github.com/Kirov7/project-project/internal/domain"
	"github.com/Kirov7/project-project/internal/repo"
	rpc "github.com/Kirov7/project-project/internal/rpc"
	"github.com/Kirov7/project-project/pkg/model"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type ProjectService struct {
	projectRpc.UnimplementedProjectServiceServer
	cache                  repo.Cache
	transaction            transaction.Transaction
	menuRepo               repo.MenuRepo
	projectRepo            repo.ProjectRepo
	projectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	projectLogRepo         repo.ProjectLogRepo
	taskRepo               repo.TaskRepo
	projectNodeDomain      *domain.ProjectNodeDomain
	taskDomain             *domain.TaskDomain
}

func NewProjectService() *ProjectService {
	return &ProjectService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransaction(),
		menuRepo:               dao.NewMenuDao(),
		projectRepo:            dao.NewProjectDao(),
		projectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskRepo:               dao.NewTaskDao(),
		projectNodeDomain:      domain.NewProjectNodeDomain(),
		taskDomain:             domain.NewTaskDomain(),
	}
}

func (s *ProjectService) Index(context.Context, *projectRpc.IndexRequest) (*projectRpc.IndexResponse, error) {
	pms, err := s.menuRepo.FindMenus(context.Background())
	if err != nil {
		zap.L().Error("Login db FindMenus err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	childs := data.CovertChild(pms)
	var mms []*projectRpc.MenuMessage
	copier.Copy(&mms, childs)
	return &projectRpc.IndexResponse{Menus: mms}, nil
}

func (s *ProjectService) FindProjectByMemId(ctx context.Context, req *projectRpc.ProjectRequest) (*projectRpc.ProjectResponse, error) {
	id := req.MemberId
	page := req.Page
	pageSize := req.PageSize
	var pms []*data.ProjectAndMember
	var total int64
	var err error
	if req.SelectBy == "" || req.SelectBy == "my" {
		pms, total, err = s.projectRepo.FindProjectByMemId(ctx, id, "and deleted = 0", page, pageSize)

	}
	if req.SelectBy == "archive" {
		pms, total, err = s.projectRepo.FindProjectByMemId(ctx, id, "and archive = 1", page, pageSize)

	}
	if req.SelectBy == "deleted" {
		pms, total, err = s.projectRepo.FindProjectByMemId(ctx, id, "and deleted = 1", page, pageSize)

	}
	if req.SelectBy == "collect" {
		pms, total, err = s.projectRepo.FindCollectProjectByMemId(context.Background(), id, page, pageSize)
		for _, pm := range pms {
			pm.Collected = model.Collected
		}
	} else {
		collectPms, _, err := s.projectRepo.FindCollectProjectByMemId(context.Background(), id, page, pageSize)
		if err != nil {
			zap.L().Error("project FindCollectProjectByMemId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		var cMap = make(map[int64]*data.ProjectAndMember)
		for _, cpm := range collectPms {
			cMap[cpm.Id] = cpm
		}
		for _, pm := range pms {
			if cMap[pm.ProjectCode] != nil {
				pm.Collected = model.Collected
			}
		}
	}
	if err != nil {
		zap.L().Error("project FindCollectProjectByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return &projectRpc.ProjectResponse{Pm: []*projectRpc.ProjectMessage{}, Total: total}, nil
	}

	var pmm []*projectRpc.ProjectMessage
	copier.Copy(&pmm, pms)
	for _, v := range pmm {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		pam := data.ToMap(pms)[v.Id]
		v.AccessControlType = pam.GetAccessControlType()
		v.OrganizationCode, _ = encrypts.EncryptInt64(pam.OrganizationCode, model.AESKey)
		v.JoinTime = tms.FormatByMill(pam.JoinTime)
		v.OwnerName = req.MemberName
		v.Order = int32(pam.Sort)
		v.CreateTime = tms.FormatByMill(pam.CreateTime)
	}
	return &projectRpc.ProjectResponse{Pm: pmm, Total: total}, nil
}

func (s *ProjectService) FindProjectTemplate(ctx context.Context, req *projectRpc.ProjectRequest) (*projectRpc.ProjectTemplateResponse, error) {
	// 1. 根据viewType查询项目模板表 得到list
	organizationCodeStr, _ := encrypts.Decrypt(req.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	page := req.Page
	pageSize := req.PageSize
	var pts []data.ProjectTemplate
	var total int64
	var err error

	if req.ViewType == -1 {
		pts, total, err = s.projectTemplateRepo.FindProjectTemplateAll(ctx, organizationCode, page, pageSize)
	}
	if req.ViewType == 1 {
		pts, total, err = s.projectTemplateRepo.FindProjectTemplateSystem(ctx, page, pageSize)
	}
	if req.ViewType == 0 {
		pts, total, err = s.projectTemplateRepo.FindProjectTemplateCustom(ctx, req.MemberId, organizationCode, page, pageSize)
	}
	if err != nil {
		zap.L().Error("FindProjectTemplate error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	// 2. 模型转换, 拿到模板id列表，去任务步骤模板表去进行查询
	tsts, err := s.taskStagesTemplateRepo.FindInProTemIds(ctx, data.ToProjectTemplateIds(pts))
	if err != nil {
		zap.L().Error("FindProjectTemplate FindInProTemIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var ptAll []*data.ProjectTemplateAll
	for _, v := range pts {
		ptAll = append(ptAll, v.Convert(data.CovertProjectMap(tsts)[v.Id]))
	}

	// 3. 组装数据
	var ptRsp []*projectRpc.ProjectTemplateMessage
	copier.Copy(&ptRsp, ptAll)
	return &projectRpc.ProjectTemplateResponse{Ptm: ptRsp, Total: total}, nil
}

func (s *ProjectService) SaveProject(ctx context.Context, req *projectRpc.ProjectRequest) (*projectRpc.SaveProjectResponse, error) {
	organizationCodeStr, _ := encrypts.Decrypt(req.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(organizationCodeStr, 10, 64)
	templateCodeStr, _ := encrypts.Decrypt(req.TemplateCode, model.AESKey)
	templateCode, _ := strconv.ParseInt(templateCodeStr, 10, 64)

	// 获取模板信息
	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	stageTemplateList, err := s.taskStagesTemplateRepo.FindByProjectTPLId(c, int(templateCode))
	if err != nil {
		zap.L().Error("SaveProject taskStagesTemplateRepo.FindByProjectTPLId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	pr := &data.Project{
		Name:              req.Name,
		Description:       req.Description,
		TemplateCode:      int(templateCode),
		CreateTime:        time.Now().UnixMilli(),
		Cover:             "https://picsum.photos/200",
		Deleted:           model.NoDeleted,
		Archive:           model.NoArchive,
		OrganizationCode:  organizationCode,
		AccessControlType: model.Open,
		TaskBoardTheme:    model.Simple,
	}
	var resp *projectRpc.SaveProjectResponse
	err = s.transaction.Action(func(conn database.DbConn) error {
		// 1. 保存项目表
		err := s.projectRepo.SaveProject(ctx, conn, pr)
		if err != nil {
			zap.L().Error("SaveProject Save error", zap.Error(err))
			return model.DBError
		}
		pm := &data.ProjectMember{
			ProjectCode: pr.Id,
			MemberCode:  req.MemberId,
			JoinTime:    time.Now().UnixMilli(),
			IsOwner:     req.MemberId,
			Authorize:   "",
		}
		fmt.Println("pm:", pm)

		// 2. 保存项目和成员的关联表
		err = s.projectRepo.SaveProjectMember(ctx, conn, pm)
		if err != nil {
			zap.L().Error("SaveProject SaveProjectMember error", zap.Error(err))
			return model.DBError
		}

		// 3. 生成任务步骤
		for index, template := range stageTemplateList {
			taskStage := &data.TaskStages{
				Name:        template.Name,
				ProjectCode: pr.Id,
				Sort:        index + 1,
				Description: "",
				CreateTime:  time.Now().UnixMilli(),
				Deleted:     model.NoDeleted,
			}
			err := s.taskStagesRepo.SaveTaskStages(ctx, conn, taskStage)
			if err != nil {
				zap.L().Error("SaveProject taskStagesRepo.SaveTaskStages error", zap.Error(err))
				return model.DBError
			}
		}
		return nil
	})
	code, _ := encrypts.EncryptInt64(pr.Id, model.AESKey)
	resp = &projectRpc.SaveProjectResponse{
		Id:               pr.Id,
		Code:             code,
		OrganizationCode: organizationCodeStr,
		Name:             pr.Name,
		Cover:            pr.Cover,
		CreateTime:       tms.FormatByMill(pr.CreateTime),
		TaskBoardTheme:   pr.TaskBoardTheme,
	}
	if err != nil {
		return nil, errs.GrpcError(err.(*errs.BError))
	}
	return resp, nil
}

//2. 项目和成员的关联表 查到项目的拥有者 去member表查名字
//3. 查收藏表 判断收藏状态
func (s *ProjectService) FindProjectDetail(ctx context.Context, req *projectRpc.ProjectRequest) (*projectRpc.ProjectDetailResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(req.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	memberId := req.MemberId
	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// 1. 查项目表
	projectAndMember, err := s.projectRepo.FindProjectByPidAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindProjectByPIdAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	ownerId := projectAndMember.IsOwner
	member, err := rpc.LoginServiceClient.FindMemInfoById(c, &login.MemIdRequest{MemId: ownerId})
	if err != nil {
		zap.L().Error("project rpc FindProjectDetail FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	// 去user模块去找了
	// TODO 优化 收藏的时候 可以放入redis
	isCollect, err := s.projectRepo.FindCollectByPidAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindCollectByPidAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if isCollect {
		projectAndMember.Collected = model.Collected
	}
	var detailMsg = &projectRpc.ProjectDetailResponse{}
	copier.Copy(detailMsg, projectAndMember)
	detailMsg.OwnerAvatar = member.Avatar
	detailMsg.OwnerName = member.Name
	detailMsg.Code, _ = encrypts.EncryptInt64(projectAndMember.Id, model.AESKey)
	detailMsg.AccessControlType = projectAndMember.GetAccessControlType()
	detailMsg.OrganizationCode, _ = encrypts.EncryptInt64(projectAndMember.OrganizationCode, model.AESKey)
	detailMsg.Order = int32(projectAndMember.Sort)
	detailMsg.CreateTime = tms.FormatByMill(projectAndMember.CreateTime)
	return detailMsg, nil
}

func (s *ProjectService) UpdateDeletedProject(ctx context.Context, msg *projectRpc.UpdateDeletedRequest) (*projectRpc.UpdateDeletedResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.projectRepo.UpdateDeletedProject(c, projectCode, msg.Deleted)
	if err != nil {
		zap.L().Error("project RecycleProject DeleteProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &projectRpc.UpdateDeletedResponse{}, nil
}

func (s *ProjectService) UpdateCollectProject(ctx context.Context, msg *projectRpc.UpdateCollectRequest) (*projectRpc.UpdateCollectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var err error
	if "collect" == msg.CollectType {
		pc := &data.ProjectCollection{
			ProjectCode: projectCode,
			MemberCode:  msg.MemberId,
			CreateTime:  time.Now().UnixMilli(),
		}
		err = s.projectRepo.SaveProjectCollect(c, pc)
	}
	if "cancel" == msg.CollectType {
		err = s.projectRepo.DeleteProjectCollect(c, projectCode, msg.MemberId)
	}
	if err != nil {
		zap.L().Error("project UpdateCollectProject SaveProjectCollect error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &projectRpc.UpdateCollectResponse{}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, msg *projectRpc.UpdateProjectRequest) (*projectRpc.UpdateProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	proj := &data.Project{
		Id:                 projectCode,
		Name:               msg.Name,
		Description:        msg.Description,
		Cover:              msg.Cover,
		TaskBoardTheme:     msg.TaskBoardTheme,
		Prefix:             msg.Prefix,
		Private:            int(msg.Private),
		OpenPrefix:         int(msg.OpenPrefix),
		OpenBeginTime:      int(msg.OpenBeginTime),
		OpenTaskPrivate:    int(msg.OpenTaskPrivate),
		Schedule:           msg.Schedule,
		AutoUpdateSchedule: int(msg.AutoUpdateSchedule),
	}
	err := s.projectRepo.UpdateProject(c, proj)
	if err != nil {
		zap.L().Error("project UpdateProject::UpdateProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &projectRpc.UpdateProjectResponse{}, nil
}

func (s *ProjectService) GetLogBySelfProject(ctx context.Context, msg *projectRpc.ProjectRequest) (*projectRpc.ProjectLogResponse, error) {
	//根据用户id查询当前的用户的日志表

	projectLogs, total, err := s.projectLogRepo.FindLogByMemberCode(context.Background(), msg.MemberId, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindLogByMemberCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//查询项目信息
	pIdList := make([]int64, len(projectLogs))
	mIdList := make([]int64, len(projectLogs))
	taskIdList := make([]int64, len(projectLogs))
	for _, v := range projectLogs {
		pIdList = append(pIdList, v.ProjectCode)
		mIdList = append(mIdList, v.MemberCode)
		taskIdList = append(taskIdList, v.SourceCode)
	}
	projects, err := s.projectRepo.FindProjectByIds(context.Background(), pIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindProjectByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pMap := make(map[int64]*data.Project)
	for _, v := range projects {
		pMap[v.Id] = v
	}
	messageList, _ := rpc.LoginServiceClient.FindMemInfoByIds(context.Background(), &login.MemIdRequest{MemIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.MemberList {
		mMap[v.Id] = v
	}
	tasks, err := s.taskRepo.FindTaskByIds(context.Background(), taskIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindTaskByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tMap := make(map[int64]*data.Task)
	for _, v := range tasks {
		tMap[v.Id] = v
	}
	var list []*data.IndexProjectLogDisplay
	for _, v := range projectLogs {
		display := v.ToIndexDisplay()
		display.ProjectName = pMap[v.ProjectCode].Name
		display.MemberAvatar = mMap[v.MemberCode].Avatar
		display.MemberName = mMap[v.MemberCode].Name
		display.TaskName = tMap[v.SourceCode].Name
		list = append(list, display)
	}
	var msgList []*projectRpc.ProjectLogMessage
	copier.Copy(&msgList, list)
	return &projectRpc.ProjectLogResponse{List: msgList, Total: total}, nil
}

func (s *ProjectService) NodeList(c context.Context, msg *projectRpc.ProjectRequest) (*projectRpc.ProjectNodeListResponse, error) {
	list, err := s.projectNodeDomain.TreeList()
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var nodes []*projectRpc.ProjectNodeMessage
	copier.Copy(&nodes, list)
	return &projectRpc.ProjectNodeListResponse{Nodes: nodes}, nil
}

func (s *ProjectService) FindProjectByMemberId(ctx context.Context, msg *projectRpc.ProjectRequest) (*projectRpc.FindProjectByMemberIdResponse, error) {
	isProjectCode := false
	var projectId int64
	if msg.ProjectCode != "" {
		projectId = encrypts.DecryptNoErr(msg.ProjectCode)
		isProjectCode = true
	}
	isTaskCode := false
	var taskId int64
	if msg.TaskCode != "" {
		taskId = encrypts.DecryptNoErr(msg.TaskCode)
		isTaskCode = true
	}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if !isProjectCode && isTaskCode {
		projectCode, ok, bError := s.taskDomain.FindProjectIdByTaskId(taskId)
		if bError != nil {
			return nil, bError
		}
		if !ok {
			return &projectRpc.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectId = projectCode
		isProjectCode = true
	}
	if isProjectCode {
		//根据projectid和memberid查询
		pm, err := s.projectRepo.FindProjectByPidAndMemId(c, projectId, msg.MemberId)
		if err != nil {
			return nil, model.DBError
		}
		if pm == nil {
			return &projectRpc.FindProjectByMemberIdResponse{
				Project:  nil,
				IsOwner:  false,
				IsMember: false,
			}, nil
		}
		projectMessage := &projectRpc.ProjectMessage{}
		copier.Copy(projectMessage, pm)
		isOwner := false
		if pm.IsOwner == 1 {
			isOwner = true
		}
		return &projectRpc.FindProjectByMemberIdResponse{
			Project:  projectMessage,
			IsOwner:  isOwner,
			IsMember: true,
		}, nil
	}
	return &projectRpc.FindProjectByMemberIdResponse{}, nil
}
