package task_service_v1

import (
	"context"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-common/tms"
	taskRpc "github.com/Kirov7/project-grpc/project/task"
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

type TaskService struct {
	taskRpc.UnimplementedTaskServiceServer
	cache                  repo.Cache
	transaction            transaction.Transaction
	projectRepo            repo.ProjectRepo
	ProjectTemplateRepo    repo.ProjectTemplateRepo
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	taskRepo               repo.TaskRepo
	projectLogRepo         repo.ProjectLogRepo
	taskWorkTimeRepo       repo.TaskWorkTimeRepo
	fileRepo               repo.FileRepo
	sourceLinkRepo         repo.SourceLinkRepo
	userRpcDomain          *domain.UserRpcDomain
}

func NewTaskService() *TaskService {
	return &TaskService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransaction(),
		projectRepo:            dao.NewProjectDao(),
		ProjectTemplateRepo:    dao.NewProjectTemplateDao(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		taskRepo:               dao.NewTaskDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskWorkTimeRepo:       dao.NewTaskWorkTimeDao(),
		fileRepo:               dao.NewFileDao(),
		sourceLinkRepo:         dao.NewSourceLinkDao(),
		userRpcDomain:          domain.NewUserRpcDomain(),
	}
}

func (t *TaskService) TaskStages(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskStagesResponse, error) {
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	page := msg.Page
	pageSize := msg.PageSize
	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	taskStages, total, err := t.taskStagesRepo.FindByProjectCode(c, projectCode, page, pageSize)
	if err != nil {
		zap.L().Error("project task TaskStages FindByProjectCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tsMap := data.ToTaskStagesMap(taskStages)

	var resp []*taskRpc.TaskStagesMessage
	copier.Copy(&resp, taskStages)
	if resp == nil {
		return &taskRpc.TaskStagesResponse{List: resp, Total: 0}, nil
	}
	for _, v := range resp {
		stages := tsMap[int(v.Id)]
		v.Code, _ = encrypts.EncryptInt64(int64(v.Id), model.AESKey)
		v.CreateTime = tms.FormatByMill(stages.CreateTime)
		v.ProjectCode = msg.ProjectCode
	}
	return &taskRpc.TaskStagesResponse{
		List:  resp,
		Total: total,
	}, nil
}

func (t *TaskService) MemberProjectList(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.MemberProjectResponse, error) {
	projectCodeStr, _ := encrypts.Decrypt(msg.ProjectCode, model.AESKey)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)

	c, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	projectMembers, total, err := t.projectRepo.FindMemberByProjectCode(c, projectCode)
	if err != nil {
		zap.L().Error("project task TaskStages FindByProjectCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 拿上用户Id列表请求用户信息
	if projectMembers == nil || len(projectMembers) == 0 {
		return &taskRpc.MemberProjectResponse{List: nil, Total: 0}, nil
	}
	var mIds []int64
	pmMap := make(map[int64]*data.ProjectMember)
	for _, member := range projectMembers {
		mIds = append(mIds, member.MemberCode)
		pmMap[member.MemberCode] = member
	}

	// 请求用户信息
	//req := &login.MemIdRequest{
	//	MemIds: mIds,
	//}
	//memberMessageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, req)
	memberList, _, err := t.userRpcDomain.MemberList(mIds)
	if err != nil {
		zap.L().Error("project MemberProjectList userRpcDomain.MemberList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}

	var list []*taskRpc.MemberProjectMessage
	for _, m := range memberList {
		owner := pmMap[m.Id].IsOwner
		mpm := &taskRpc.MemberProjectMessage{
			MemberCode: m.Id,
			Name:       m.Name,
			Avatar:     m.Avatar,
			Email:      m.Email,
			Code:       m.Code,
		}
		if m.Id == owner {
			mpm.IsOwner = model.Owner
		}
		list = append(list, mpm)
	}
	return &taskRpc.MemberProjectResponse{
		List:  list,
		Total: total,
	}, nil
}

func (t *TaskService) TaskList(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskListResponse, error) {
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskList, err := t.taskRepo.FindTaskByStageCode(c, int(stageCode))
	if err != nil {
		zap.L().Error("project task TaskList FindTaskByStageCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var taskDisplayList []*data.TaskDisplay
	var mIds []int64
	for _, v := range taskList {
		display := v.ToTaskDisplay()
		if v.Private == model.Private {
			// 隐私模式
			tm, err := t.taskRepo.FindTaskMemberByTaskId(ctx, v.Id, msg.MemberId)
			if err != nil {
				zap.L().Error("project task TaskList FindTaskMemberByTaskId error", zap.Error(err))
				return nil, errs.GrpcError(model.DBError)
			}
			if tm == nil {
				display.CanRead = model.NoCanRead
			} else {
				display.CanRead = model.CanRead
			}
		}
		taskDisplayList = append(taskDisplayList, display)
		mIds = append(mIds, v.AssignTo)
	}
	if mIds == nil || len(mIds) <= 0 {
		return &taskRpc.TaskListResponse{List: nil}, nil
	}
	//msgList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.MemIdRequest{MemIds: mIds})
	_, memberMap, err := t.userRpcDomain.MemberList(mIds)
	if err != nil {
		zap.L().Error("project task TaskList userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}
	//memberMap := make(map[int64]*login.MemberMessage)
	//for _, v := range memberList {
	//	memberMap[v.Id] = v
	//}

	for _, v := range taskDisplayList {
		message := memberMap[encrypts.DecryptNoErr(v.AssignTo)]
		e := data.Executor{
			Name:   message.Name,
			Avatar: message.Avatar,
		}
		v.Executor = e
	}
	var taskMessageList []*taskRpc.TaskMessage
	copier.Copy(&taskMessageList, taskDisplayList)
	return &taskRpc.TaskListResponse{List: taskMessageList}, nil
}

func (t *TaskService) SaveTask(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskMessage, error) {
	// 检查业务逻辑
	if msg.Name == "" {
		return nil, errs.GrpcError(model.TaskNameNotNull)
	}
	stageCode := encrypts.DecryptNoErr(msg.StageCode)
	taskStages, err := t.taskStagesRepo.FindById(ctx, int(stageCode))
	if err != nil {
		zap.L().Error("project task SaveTask taskStagesRepo.FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskStages == nil {
		return nil, errs.GrpcError(model.TaskStagesNotNull)
	}
	projectCode := encrypts.DecryptNoErr(msg.ProjectCode)
	findProject, err := t.projectRepo.FindProjectById(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask projectRepo.FindProjectById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if findProject == nil || findProject.Deleted == model.Deleted {
		return nil, errs.GrpcError(model.ProjectAlreadyDeleted)
	}
	maxIdNum, err := t.taskRepo.FindTaskMaxIdNum(ctx, projectCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskMaxIdNum error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	maxSort, err := t.taskRepo.FindTaskSort(ctx, projectCode, stageCode)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskSort error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	assignTo := encrypts.DecryptNoErr(msg.AssignTo)
	ts := &data.Task{
		Name:        msg.Name,
		CreateTime:  time.Now().UnixMilli(),
		CreateBy:    msg.MemberId,
		AssignTo:    assignTo,
		ProjectCode: projectCode,
		StageCode:   int(stageCode),
		IdNum:       int(maxIdNum + 1),
		Private:     findProject.OpenTaskPrivate,
		Sort:        int(maxSort + 1024),
		BeginTime:   time.Now().UnixMilli(),
		EndTime:     time.Now().Add(2 * 24 * time.Hour).UnixMilli(),
	}
	err = t.transaction.Action(func(conn database.DbConn) error {
		err = t.taskRepo.SaveTask(ctx, conn, ts)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTask error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		tm := &data.TaskMember{
			MemberCode: assignTo,
			TaskCode:   ts.Id,
			JoinTime:   time.Now().UnixMilli(),
			IsOwner:    model.Owner,
		}
		if assignTo == msg.MemberId {
			tm.IsExecutor = model.Executor
		}
		err = t.taskRepo.SaveTaskMember(ctx, conn, tm)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTaskMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	display := ts.ToTaskDisplay()
	member, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.MemIdRequest{MemId: assignTo})
	if err != nil {
		return nil, err
	}
	display.Executor = data.Executor{
		Name:   member.Name,
		Avatar: member.Avatar,
		Code:   member.Code,
	}
	// 添加任务动态
	go createProjectLog(t.projectLogRepo, ts.ProjectCode, ts.Id, ts.Name, ts.AssignTo, "create", "task")

	tm := &taskRpc.TaskMessage{}
	copier.Copy(tm, display)
	return tm, nil
}

func (t *TaskService) TaskSort(ctx context.Context, msg *taskRpc.TaskSortRequest) (*taskRpc.TaskSortResponse, error) {
	preTaskCode := encrypts.DecryptNoErr(msg.PreTaskCode)
	toStageCode := encrypts.DecryptNoErr(msg.ToStageCode)
	if msg.PreTaskCode == msg.NextTaskCode {
		return &taskRpc.TaskSortResponse{}, nil
	}
	err := t.sortTask(preTaskCode, msg.NextTaskCode, toStageCode)
	if err != nil {
		return nil, err
	}
	return &taskRpc.TaskSortResponse{}, nil
}

func (t *TaskService) sortTask(preTaskCode int64, nextTaskCode string, toStageCode int64) error {
	//1. 从小到大排
	//2. 原有的顺序  比如 1 2 3 4 5 4排到2前面去 4的序号在1和2 之间 如果4是最后一个 保证 4比所有的序号都打 如果 排到第一位 直接置为0

	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ts, err := t.taskRepo.FindTaskById(c, preTaskCode)
	if err != nil {
		zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
		return errs.GrpcError(model.DBError)
	}
	err = t.transaction.Action(func(conn database.DbConn) error {
		//如果相等是不需要进行改变的
		ts.StageCode = int(toStageCode)
		if nextTaskCode != "" {
			//意味着要进行排序的替换
			nextTaskCode := encrypts.DecryptNoErr(nextTaskCode)
			next, err := t.taskRepo.FindTaskById(c, nextTaskCode)
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			// next.Sort 要找到比它小的那个任务
			prepre, err := t.taskRepo.FindTaskByStageCodeLtSort(c, next.StageCode, next.Sort)
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskByStageCodeLtSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			if prepre != nil {
				ts.Sort = (prepre.Sort + next.Sort) / 2
			}
			if prepre == nil {
				ts.Sort = 0
			}
			//sort := ts.Sort
			//ts.Sort = next.Sort
			//next.Sort = sort
			//err = t.taskRepo.UpdateTaskSort(c, conn, next)
			//if err != nil {
			//	zap.L().Error("project task TaskSort taskRepo.UpdateTaskSort error", zap.Error(err))
			//	return errs.GrpcError(model.DBError)
			//}
		} else {
			maxSort, err := t.taskRepo.FindTaskSort(c, ts.ProjectCode, int64(ts.StageCode))
			if err != nil {
				zap.L().Error("project task TaskSort taskRepo.FindTaskSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			ts.Sort = int(maxSort + 1024)
		}
		if ts.Sort < 50 {
			//重置排序
			err = t.resetSort(toStageCode)
			if err != nil {
				zap.L().Error("project task TaskSort resetSort error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			return t.sortTask(preTaskCode, nextTaskCode, toStageCode)
		}
		err = t.taskRepo.UpdateTaskSort(c, conn, ts)
		if err != nil {
			zap.L().Error("project task TaskSort taskRepo.UpdateTaskSort error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	return err
}

func (t *TaskService) resetSort(stageCode int64) error {
	list, err := t.taskRepo.FindTaskByStageCode(context.Background(), int(stageCode))
	if err != nil {
		return err
	}
	return t.transaction.Action(func(conn database.DbConn) error {
		iSort := 1024
		for index, v := range list {
			v.Sort = (index + 1) * iSort
			return t.taskRepo.UpdateTaskSort(context.Background(), conn, v)
		}
		return nil
	})
}

func (t *TaskService) MyTaskList(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.MyTaskListResponse, error) {
	var tsList []*data.Task
	var err error
	var total int64
	if msg.TaskType == model.MyExecute {
		//我执行的
		tsList, total, err = t.taskRepo.FindTaskByAssignTo(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByAssignTo error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == model.MyParticipate {
		//参与的
		tsList, total, err = t.taskRepo.FindTaskByMemberCode(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByMemberCode error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == model.MyCreate {
		//我创建的
		tsList, total, err = t.taskRepo.FindTaskByCreateBy(ctx, msg.MemberId, int(msg.Type), msg.Page, msg.PageSize)
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByCreateBy error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if tsList == nil || len(tsList) <= 0 {
		return &taskRpc.MyTaskListResponse{List: nil, Total: 0}, nil
	}
	var pids []int64
	var mids []int64
	for _, v := range tsList {
		pids = append(pids, v.ProjectCode)
		mids = append(mids, v.AssignTo)
	}
	pListChan := make(chan []*data.Project)
	defer close(pListChan)
	mListChan := make(chan []*login.MemberMessage)
	defer close(mListChan)

	go func() {
		pList, _ := t.projectRepo.FindProjectByIds(ctx, pids)
		pListChan <- pList
	}()
	go func() {
		//mList, _ := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.MemIdRequest{MemIds: mids})
		mList, _, _ := t.userRpcDomain.MemberList(mids)
		mListChan <- mList
	}()

	pList := <-pListChan
	projectMap := data.ToProjectMap(pList)
	mList := <-mListChan
	mMap := make(map[int64]*login.MemberMessage)

	for _, v := range mList {
		mMap[v.Id] = v
	}
	var mtdList []*data.MyTaskDisplay
	for _, v := range tsList {
		memberMessage := mMap[v.AssignTo]
		name := memberMessage.Name
		avatar := memberMessage.Avatar
		mtd := v.ToMyTaskDisplay(projectMap[v.ProjectCode], name, avatar)
		mtdList = append(mtdList, mtd)
	}
	var myMsgs []*taskRpc.MyTaskMessage
	copier.Copy(&myMsgs, mtdList)
	return &taskRpc.MyTaskListResponse{List: myMsgs, Total: total}, nil
}

func (t *TaskService) ReadTask(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskMessage, error) {
	// 根据taskCode查询任务详情 根据任务查询项目详情 根据任务查询步骤详情 查询任务的执行者成员
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskInfo, err := t.taskRepo.FindTaskById(c, taskCode)
	if err != nil {
		zap.L().Error("project task ReadTask taskRepo FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskInfo == nil {
		return &taskRpc.TaskMessage{}, nil
	}
	display := taskInfo.ToTaskDisplay()
	if taskInfo.Private == 1 {
		//代表隐私模式
		taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, taskInfo.Id, msg.MemberId)
		if err != nil {
			zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if taskMember != nil {
			display.CanRead = model.CanRead
		} else {
			display.CanRead = model.NoCanRead
		}
	}
	pj, err := t.projectRepo.FindProjectById(c, taskInfo.ProjectCode)
	display.ProjectName = pj.Name
	taskStages, err := t.taskStagesRepo.FindById(c, taskInfo.StageCode)
	display.StageName = taskStages.Name
	// in ()
	memberMessage, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.MemIdRequest{MemId: taskInfo.AssignTo})
	if err != nil {
		zap.L().Error("project task TaskList LoginServiceClient.FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	e := data.Executor{
		Name:   memberMessage.Name,
		Avatar: memberMessage.Avatar,
	}
	display.Executor = e
	var taskMessage = &taskRpc.TaskMessage{}
	copier.Copy(taskMessage, display)
	return taskMessage, nil
}

func (t *TaskService) ListTaskMember(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskMemberList, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskMembers, total, err := t.taskRepo.FindTaskMemberPage(c, taskCode, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project task TaskList taskRepo.FindTaskMemberPage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var mids []int64
	for _, v := range taskMembers {
		mids = append(mids, v.MemberCode)
	}
	//messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.MemIdRequest{MemIds: mids})
	//mMap := make(map[int64]*login.MemberMessage, len(messageList.MemberList))
	//for _, v := range messageList.MemberList {
	//	mMap[v.Id] = v
	//}
	_, mMap, err := t.userRpcDomain.MemberList(mids)
	if err != nil {
		zap.L().Error("project task ListTaskMember userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}

	var taskMemeberMemssages []*taskRpc.TaskMemberMessage
	for _, v := range taskMembers {
		tm := &taskRpc.TaskMemberMessage{}
		tm.Code = encrypts.EncryptNoErr(v.MemberCode)
		tm.Id = v.Id
		message := mMap[v.MemberCode]
		tm.Name = message.Name
		tm.Avatar = message.Avatar
		tm.IsExecutor = int32(v.IsExecutor)
		tm.IsOwner = int32(v.IsOwner)
		taskMemeberMemssages = append(taskMemeberMemssages, tm)
	}
	return &taskRpc.TaskMemberList{List: taskMemeberMemssages, Total: total}, nil
}

func createProjectLog(logRepo repo.ProjectLogRepo, projectCode int64, taskCode int64, taskName string, toMemberCode int64, logType string, actionType string) {
	remark := ""
	if logType == "create" {
		remark = "创建了任务"
	}
	pl := &data.ProjectLog{
		MemberCode:  toMemberCode,
		SourceCode:  taskCode,
		Content:     taskName,
		Remark:      remark,
		ProjectCode: projectCode,
		CreateTime:  time.Now().UnixMilli(),
		Type:        logType,
		ActionType:  actionType,
		Icon:        "plus",
		IsComment:   0,
		IsRobot:     0,
	}
	logRepo.SaveProjectLog(pl)
}

func (t *TaskService) TaskLog(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskLogList, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	all := msg.All
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var list []*data.ProjectLog
	var total int64
	var err error
	if all == 1 {
		//显示全部
		list, total, err = t.projectLogRepo.FindLogByTaskCode(c, taskCode, int(msg.Comment))
	}
	if all == 0 {
		//分页
		list, total, err = t.projectLogRepo.FindLogByTaskCodePage(c, taskCode, int(msg.Comment), int(msg.Page), int(msg.PageSize))
	}
	if err != nil {
		zap.L().Error("project task TaskLog projectLogRepo.FindLogByTaskCodePage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if total == 0 {
		return &taskRpc.TaskLogList{}, nil
	}
	var displayList []*data.ProjectLogDisplay
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	//messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.MemIdRequest{MemIds: mIdList})
	//mMap := make(map[int64]*login.MemberMessage)
	//for _, v := range messageList.MemberList {
	//	mMap[v.Id] = v
	//}
	_, mMap, err := t.userRpcDomain.MemberList(mIdList)
	if err != nil {
		zap.L().Error("project task TaskLog userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}

	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*taskRpc.TaskLog
	copier.Copy(&l, displayList)
	return &taskRpc.TaskLogList{List: l, Total: total}, nil
}

func (t *TaskService) TaskWorkTimeList(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskWorkTimeResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var list []*data.TaskWorkTime
	var err error
	list, err = t.taskWorkTimeRepo.FindWorkTimeList(c, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindWorkTimeList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(list) == 0 {
		return &taskRpc.TaskWorkTimeResponse{}, nil
	}
	var displayList []*data.TaskWorkTimeDisplay
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	//messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.MemIdRequest{MemIds: mIdList})
	//mMap := make(map[int64]*login.MemberMessage)
	//for _, v := range messageList.MemberList {
	//	mMap[v.Id] = v
	//}
	_, mMap, err := t.userRpcDomain.MemberList(mIdList)

	if err != nil {
		zap.L().Error("project task TaskWorkTimeList userRpcDomain.MemberList error", zap.Error(err))
		return nil, err
	}

	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*taskRpc.TaskWorkTime
	copier.Copy(&l, displayList)
	return &taskRpc.TaskWorkTimeResponse{List: l, Total: int64(len(list))}, nil
}

func (t *TaskService) SaveTaskWorkTime(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.SaveTaskWorkTimeResponse, error) {
	tmt := &data.TaskWorkTime{}
	tmt.BeginTime = msg.BeginTime
	tmt.Num = int(msg.Num)
	tmt.Content = msg.Content
	tmt.TaskCode = encrypts.DecryptNoErr(msg.TaskCode)
	tmt.MemberCode = msg.MemberId
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := t.taskWorkTimeRepo.Save(c, tmt)
	if err != nil {
		zap.L().Error("project task SaveTaskWorkTime taskWorkTimeRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &taskRpc.SaveTaskWorkTimeResponse{}, nil
}

func (t *TaskService) SaveTaskFile(ctx context.Context, msg *taskRpc.TaskFileReqMessage) (*taskRpc.TaskFileResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	//存file表
	f := &data.File{
		PathName:         msg.PathName,
		Title:            msg.FileName,
		Extension:        msg.Extension,
		Size:             int(msg.Size),
		ObjectType:       "",
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		TaskCode:         encrypts.DecryptNoErr(msg.TaskCode),
		ProjectCode:      encrypts.DecryptNoErr(msg.ProjectCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Downloads:        0,
		Extra:            "",
		Deleted:          model.NoDeleted,
		FileType:         msg.FileType,
		FileUrl:          msg.FileUrl,
		DeletedTime:      0,
	}
	err := t.fileRepo.Save(context.Background(), f)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//存入source_link
	sl := &data.SourceLink{
		SourceType:       "file",
		SourceCode:       f.Id,
		LinkType:         "task",
		LinkCode:         taskCode,
		OrganizationCode: encrypts.DecryptNoErr(msg.OrganizationCode),
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Sort:             0,
	}
	err = t.sourceLinkRepo.Save(context.Background(), sl)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &taskRpc.TaskFileResponse{}, nil
}

func (t *TaskService) TaskSources(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.TaskSourceResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	sourceLinks, err := t.sourceLinkRepo.FindByTaskCode(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.FindByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(sourceLinks) == 0 {
		return &taskRpc.TaskSourceResponse{}, nil
	}
	var fIdList []int64
	for _, v := range sourceLinks {
		fIdList = append(fIdList, v.SourceCode)
	}
	files, err := t.fileRepo.FindByIds(context.Background(), fIdList)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.FindByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	fMap := make(map[int64]*data.File)
	for _, v := range files {
		fMap[v.Id] = v
	}
	var list []*data.SourceLinkDisplay
	for _, v := range sourceLinks {
		list = append(list, v.ToDisplay(fMap[v.SourceCode]))
	}
	var slMsg []*taskRpc.TaskSourceMessage
	copier.Copy(&slMsg, list)
	return &taskRpc.TaskSourceResponse{List: slMsg}, nil
}

func (t *TaskService) CreateComment(ctx context.Context, msg *taskRpc.TaskRequest) (*taskRpc.CreateCommentResponse, error) {
	taskCode := encrypts.DecryptNoErr(msg.TaskCode)
	taskById, err := t.taskRepo.FindTaskById(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task CreateComment fileRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pl := &data.ProjectLog{
		MemberCode:   msg.MemberId,
		Content:      msg.CommentContent,
		Remark:       msg.CommentContent,
		Type:         "createComment",
		CreateTime:   time.Now().UnixMilli(),
		SourceCode:   taskCode,
		ActionType:   "task",
		ToMemberCode: 0,
		IsComment:    model.Comment,
		ProjectCode:  taskById.ProjectCode,
		Icon:         "plus",
		IsRobot:      0,
	}
	go t.projectLogRepo.SaveProjectLog(pl)
	return &taskRpc.CreateCommentResponse{}, nil
}
