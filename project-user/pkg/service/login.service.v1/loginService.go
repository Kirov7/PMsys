package login_service_v1

import (
	"context"
	"encoding/json"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-common/jwts"
	"github.com/Kirov7/project-common/tms"
	loginRpc "github.com/Kirov7/project-grpc/user/login"
	"github.com/Kirov7/project-user/config"
	"github.com/Kirov7/project-user/internal/dao"
	"github.com/Kirov7/project-user/internal/data/member"
	"github.com/Kirov7/project-user/internal/data/organization"
	gorms "github.com/Kirov7/project-user/internal/database"
	"github.com/Kirov7/project-user/internal/database/transaction"
	"github.com/Kirov7/project-user/internal/repo"
	"github.com/Kirov7/project-user/pkg/model"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
	"time"
)

type LoginService struct {
	loginRpc.UnimplementedLoginServiceServer
	cache            repo.Cache
	memberRepo       repo.MemberRepo
	organizationRepo repo.OrganizationRepo
	transaction      transaction.Transaction
}

func NewLoginService() *LoginService {
	return &LoginService{
		cache:            dao.Rc,
		memberRepo:       dao.NewMemberDao(),
		organizationRepo: dao.NewOrganizationDao(),
		transaction:      dao.NewTransaction(),
	}
}

func (s *LoginService) GetCaptcha(ctx context.Context, req *loginRpc.CaptchaRequest) (*loginRpc.CaptchaResponse, error) {
	// 1. 获取参数
	mobile := req.Mobile
	// 2. 校验参数
	if !common.VerifyMobile(mobile) {
		return nil, errs.GrpcError(model.NoLegalMobile)
	}
	// 3. 生成验证码 (随机4位1000-9999或者6位100000-999999)
	code := "753159"
	// 4. 调用短信平台 (三方 放入协程中执行 接口可以快速响应)
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功,发送短信 INFO")
		// 缓存系统后期未来可能会更换,可以进一步抽象
		// 5. 存储验证码 redis 当中 过期时间5分钟
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := s.cache.Set(c, model.RegisterRedisKey+mobile, code, 5*time.Minute)
		if err != nil {
			log.Printf("验证码存入redis出错,casue by: %v", err)
		}
	}()
	return &loginRpc.CaptchaResponse{Code: code}, nil
}

// Register todo 添加账户模块
func (s *LoginService) Register(ctx context.Context, req *loginRpc.RegisterRequest) (*loginRpc.RegisterResponse, error) {
	// 1. 校验验证码
	c := context.Background()
	redisCode, err := s.cache.Get(c, model.RegisterRedisKey+req.Mobile)
	if err == redis.Nil {
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil {
		zap.L().Error("Register Cache get error", zap.Error(err))
		return nil, errs.GrpcError(model.CacheError)
	}
	if redisCode != req.Captcha {
		return nil, errs.GrpcError(model.CaptchaError)
	}
	// 2. 校验业务逻辑 (邮箱/账号/手机号是否注册)
	exit, err := s.memberRepo.GetMemberByEmail(c, req.Email)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.EmailExist)
	}

	exit, err = s.memberRepo.GetMemberByAccount(c, req.Name)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.AccountExist)
	}

	exit, err = s.memberRepo.GetMemberByMobile(c, req.Mobile)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exit {
		return nil, errs.GrpcError(model.MobileExist)
	}
	// 3. 执行业务 将数据存储member表 生成一个数据 存入organization表
	pwd := encrypts.Md5(req.Password)
	mem := &member.Member{
		Account:       req.Name,
		Password:      pwd,
		Name:          req.Name,
		Mobile:        req.Mobile,
		Email:         req.Email,
		Avatar:        "https://source.unsplash.com/collection/94734566/1920x1080",
		CreateTime:    time.Now().UnixMilli(),
		LastLoginTime: time.Now().UnixMilli(),
		Status:        model.StatusNormal,
	}
	err = s.transaction.Action(func(conn gorms.DbConn) error {
		err = s.memberRepo.SaveMember(conn, c, mem)
		if err != nil {
			zap.L().Error("Register DB get error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}

		org := &organization.Organization{
			Name:       mem.Name + "个人组织",
			MemberId:   mem.Id,
			CreateTime: time.Now().UnixMilli(),
			Personal:   int32(model.Personal),
			Avatar:     "https://source.unsplash.com/collection/94734566/1920x1080",
		}
		err = s.organizationRepo.SaveOrganization(conn, c, org)
		if err != nil {
			zap.L().Error("register SaveOrganization db err", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 5. 返回
	return &loginRpc.RegisterResponse{}, nil
}

func (s *LoginService) Login(ctx context.Context, req *loginRpc.LoginRequest) (*loginRpc.LoginResponse, error) {
	c := context.Background()
	// 1. 查询账号密码是否正确
	pwd := encrypts.Md5(req.Password)
	mem, err := s.memberRepo.FindMember(c, req.Account, pwd)
	if err != nil {
		zap.L().Error("Login db FindMember err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if mem == nil {
		zap.L().Error("cant find account err", zap.Error(err))
		return nil, errs.GrpcError(model.AccountOrPasswordError)
	}
	code, _ := encrypts.EncryptInt64(mem.Id, model.AESKey)
	memMsg := &loginRpc.MemberMessage{
		Id:            mem.Id,
		Name:          mem.Name,
		Mobile:        mem.Mobile,
		RealName:      mem.Name,
		Account:       mem.Account,
		Status:        int32(mem.Status),
		LastLoginTime: tms.FormatByMill(mem.LastLoginTime),
		CreateTime:    tms.FormatByMill(mem.CreateTime),
		Address:       mem.Address,
		Province:      int32(mem.Province),
		City:          int32(mem.City),
		Area:          int32(mem.Area),
		Email:         mem.Email,
		Code:          code,
	}
	// 2. 根据用户id查询组织
	orgs, err := s.organizationRepo.FindOrganizationByMemId(c, mem.Id)
	if err != nil {
		zap.L().Error("Login db FindOrganizationByMemId err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgMsgs []*loginRpc.OrganizationMessage
	for _, org := range orgs {
		code, _ := encrypts.EncryptInt64(org.Id, model.AESKey)
		orgMessage := &loginRpc.OrganizationMessage{
			Id:          org.Id,
			Name:        org.Name,
			Avatar:      org.Avatar,
			Description: org.Description,
			MemberId:    org.MemberId,
			CreateTime:  tms.FormatByMill(org.CreateTime),
			Personal:    org.Personal,
			Address:     org.Address,
			Province:    org.Province,
			City:        org.City,
			Area:        org.Area,
			Code:        code,
		}
		orgMsgs = append(orgMsgs, orgMessage)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	// 3. 用jwt生成token
	memIdStr := strconv.FormatInt(mem.Id, 10)
	acExp := time.Duration(config.AppConf.JwtConfig.AccessExp*3600*24) * time.Second
	rfExp := time.Duration(config.AppConf.JwtConfig.RefreshExp*3600*24) * time.Second

	token := jwts.CreateToken(memIdStr, config.AppConf.JwtConfig.AccessSecret, config.AppConf.JwtConfig.RefreshSecret, acExp, rfExp, req.Ip)
	tokenList := &loginRpc.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenType:      "bearer",
		AccessTokenExp: token.AccessExp,
	}
	// todo 放入缓存 member orgs
	go func() {
		memJson, _ := json.Marshal(mem)
		s.cache.Set(context.Background(), model.Member+memIdStr, string(memJson), acExp)
		orgJson, _ := json.Marshal(orgs)
		s.cache.Set(context.Background(), model.MemberOrganization+memIdStr, string(orgJson), acExp)
	}()
	return &loginRpc.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgMsgs,
		TokenList:        tokenList,
	}, nil
}

func (s *LoginService) TokenVerify(ctx context.Context, req *loginRpc.LoginRequest) (*loginRpc.LoginResponse, error) {
	token := req.Token
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}
	parseToken, err := jwts.ParseToken(token, config.AppConf.JwtConfig.AccessSecret, req.Ip)
	if err != nil {
		zap.L().Error("Login TokenVerify err", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	//id, _ := strconv.ParseInt(parseToken, 10, 64)
	memJson, err := s.cache.Get(context.Background(), model.Member+parseToken)
	if err != nil {
		zap.L().Error("Login TokenVerify cache get member err", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if memJson == "" {
		zap.L().Error("Login TokenVerify cache get member expire")
		return nil, errs.GrpcError(model.NoLogin)
	}
	memberById := &member.Member{}
	json.Unmarshal([]byte(memJson), memberById)
	//memberById, err := s.memberRepo.FindMemberById(context.Background(), id)
	//if err != nil {
	//	zap.L().Error("Login db FindMemberById err", zap.Error(err))
	//	return nil, errs.GrpcError(model.DBError)
	//}
	code, _ := encrypts.EncryptInt64(memberById.Id, model.AESKey)
	memMsg := &loginRpc.MemberMessage{
		Id:            memberById.Id,
		Name:          memberById.Name,
		Mobile:        memberById.Mobile,
		RealName:      memberById.Name,
		Account:       memberById.Account,
		Status:        int32(memberById.Status),
		LastLoginTime: tms.FormatByMill(memberById.LastLoginTime),
		Address:       memberById.Address,
		Province:      int32(memberById.Province),
		City:          int32(memberById.City),
		Area:          int32(memberById.Area),
		Email:         memberById.Email,
		Code:          code,
	}

	orgsJson, err := s.cache.Get(context.Background(), model.MemberOrganization+parseToken)
	if err != nil {
		zap.L().Error("Login TokenVerify cache get organization err", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if orgsJson == "" {
		zap.L().Error("Login TokenVerify cache get organization expire")
		return nil, errs.GrpcError(model.NoLogin)
	}
	orgs := []*organization.Organization{}
	json.Unmarshal([]byte(orgsJson), &orgs)

	//orgs, err := s.organizationRepo.FindOrganizationByMemId(context.Background(), memberById.Id)
	//if err != nil {
	//	zap.L().Error("Login db FindOrganizationByMemId err", zap.Error(err))
	//	return nil, errs.GrpcError(model.DBError)
	//}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return &loginRpc.LoginResponse{Member: memMsg}, nil
}

func (s *LoginService) OrgList(ctx context.Context, msg *loginRpc.OrgListRequest) (*loginRpc.OrgListResponse, error) {
	memId := msg.MemId
	orgs, err := s.organizationRepo.FindOrganizationByMemId(ctx, memId)
	if err != nil {
		zap.L().Error("MyOrgList FindOrganizationByMemId err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgsMessage []*loginRpc.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, org := range orgsMessage {
		org.Code, _ = encrypts.EncryptInt64(org.Id, model.AESKey)
	}
	return &loginRpc.OrgListResponse{OrganizationList: orgsMessage}, nil
}

func (s *LoginService) FindMemInfoById(ctx context.Context, msg *loginRpc.MemIdRequest) (*loginRpc.MemberMessage, error) {
	memberById, err := s.memberRepo.FindMemberById(context.Background(), msg.MemId)
	if err != nil {
		zap.L().Error("FindMemInfoById db FindMemberById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	memMsg := &loginRpc.MemberMessage{}
	copier.Copy(memMsg, memberById)
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AESKey)
	orgs, err := s.organizationRepo.FindOrganizationByMemId(context.Background(), memberById.Id)
	if err != nil {
		zap.L().Error("FindMemInfoById db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return memMsg, nil
}

func (s *LoginService) FindMemInfoByIds(ctx context.Context, msg *loginRpc.MemIdRequest) (*loginRpc.MemberMessageList, error) {
	memberList, err := s.memberRepo.FindMemberByIds(context.Background(), msg.MemIds)
	if err != nil {
		zap.L().Error("FindMemInfoByIds db memberRepo.FindMemberByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if memberList == nil || len(memberList) <= 0 {
		return &loginRpc.MemberMessageList{MemberList: nil}, nil
	}
	memMap := make(map[int64]*member.Member)
	for _, mem := range memberList {
		memMap[mem.Id] = mem
	}
	var memMsgs []*loginRpc.MemberMessage
	copier.Copy(&memMsgs, memberList)
	for _, memMsg := range memMsgs {
		m := memMap[memMsg.Id]
		memMsg.CreateTime = tms.FormatByMill(m.CreateTime)
		memMsg.Code = encrypts.EncryptNoErr(memMsg.Id)
	}

	return &loginRpc.MemberMessageList{MemberList: memMsgs}, nil
}
