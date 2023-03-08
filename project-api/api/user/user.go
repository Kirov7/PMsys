package user

import (
	"context"
	"github.com/Kirov7/project-api/pkg/model/user"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-grpc/user/login"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"time"
)

type HandlerUser struct {
}

func NewHandlerUser() *HandlerUser {
	return &HandlerUser{}
}

func (h *HandlerUser) getCaptcha(ctx *gin.Context) {
	resp := &common.Result{}
	mobile := ctx.PostForm("mobile")

	// 发起grpc调用
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	captchaResp, err := LoginServiceClient.GetCaptcha(c, &login.CaptchaRequest{Mobile: mobile})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, resp.Fail(code, msg))
		return
	}

	ctx.JSON(http.StatusOK, resp.Success(captchaResp.Code))
}

func (h *HandlerUser) register(ctx *gin.Context) {
	// 1. 接受参数 参数模型
	resp := &common.Result{}
	var req user.RegisterReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, resp.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	// 2. 校验参数 判断参数是否合法
	if err := req.Verify(); err != nil {
		ctx.JSON(http.StatusOK, resp.Fail(http.StatusBadRequest, err.Error()))
	}
	// 3. 调用user grpc服务 获取响应
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	registerRpcReq := &login.RegisterRequest{
		Mobile:   req.Mobile,
		Name:     req.Name,
		Password: req.Password,
		Captcha:  req.Captcha,
		Email:    req.Email,
	}
	_, err = LoginServiceClient.Register(c, registerRpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, resp.Fail(code, msg))
		return
	}
	// 4. 返回结果
	ctx.JSON(http.StatusOK, resp.Success(""))
}

func (h *HandlerUser) login(ctx *gin.Context) {
	// 1. 接受参数 参数模型
	resp := &common.Result{}
	var req user.LoginReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, resp.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	// 2.调用user grpc 完成登录
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rpcReq := &login.LoginRequest{
		Account:  req.Account,
		Password: req.Password,
		Ip:       getIp(ctx),
	}
	loginRpcResp, err := LoginServiceClient.Login(c, rpcReq)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, resp.Fail(code, msg))
		return
	}

	var organizationList []user.OrganizationList
	for _, organization := range loginRpcResp.OrganizationList {
		organizationMsg := user.OrganizationList{
			Name:        organization.Name,
			Avatar:      organization.Avatar,
			Description: organization.Description,
			CreateTime:  organization.CreateTime,
			Personal:    organization.Personal,
			Address:     organization.Address,
			Province:    organization.Province,
			City:        organization.City,
			Area:        organization.Area,
			Code:        organization.Code,
			OwnerCode:   organization.OwnerCode,
		}
		organizationList = append(organizationList, organizationMsg)
	}

	loginResp := &user.LoginRsp{
		Member: user.Member{
			Name:          loginRpcResp.Member.Name,
			Mobile:        loginRpcResp.Member.Mobile,
			Status:        int(loginRpcResp.Member.Status),
			Code:          loginRpcResp.Member.Code,
			Email:         loginRpcResp.Member.Email,
			LastLoginTime: loginRpcResp.Member.LastLoginTime,
			CreateTime:    loginRpcResp.Member.CreateTime,
		},
		TokenList: user.TokenList{
			AccessToken:    loginRpcResp.TokenList.AccessToken,
			RefreshToken:   loginRpcResp.TokenList.RefreshToken,
			TokenType:      loginRpcResp.TokenList.TokenType,
			AccessTokenExp: loginRpcResp.TokenList.AccessTokenExp,
		},
		OrganizationList: organizationList,
	}
	ctx.JSON(http.StatusOK, resp.Success(loginResp))
}

func (h *HandlerUser) orgList(ctx *gin.Context) {
	result := &common.Result{}
	token := ctx.GetHeader("Authorization")
	//验证用户是否已经登录
	ip := getIp(ctx)
	mem, err2 := LoginServiceClient.TokenVerify(context.Background(), &login.LoginRequest{Token: token, Ip: ip})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	list, err2 := LoginServiceClient.OrgList(context.Background(), &login.OrgListRequest{MemId: mem.Member.Id})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	if list.OrganizationList == nil {
		ctx.JSON(http.StatusOK, result.Success([]*user.OrganizationList{}))
		return
	}
	var orgs []*user.OrganizationList
	copier.Copy(&orgs, list.OrganizationList)
	ctx.JSON(http.StatusOK, result.Success(orgs))
}

func getIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.01"
	}
	return ip
}
