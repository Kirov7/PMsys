package middle

import (
	"context"
	"github.com/Kirov7/project-api/api/user"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	"github.com/Kirov7/project-grpc/user/login"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func TokenVerify() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		result := &common.Result{}
		// 1. 从Header中获取token
		token := ctx.GetHeader("Authorization")
		// 2. 调用user服务进行token认证
		c, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		ip := getIp(ctx)
		// todo 先去查询node表, 确认不使用登录控制的接口,不做登录认证
		TokenVerifyRpcResp, err := user.LoginServiceClient.TokenVerify(c, &login.LoginRequest{Token: token, Ip: ip})
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			ctx.JSON(http.StatusOK, result.Fail(code, msg))
			ctx.Abort()
			return
		}
		// 3. 处理结果, 认证通过讲信息存入gin的context, 失败则返回未登录
		ctx.Set("memberId", TokenVerifyRpcResp.Member.Id)
		ctx.Set("memberName", TokenVerifyRpcResp.Member.Name)
		ctx.Set("organizationCode", TokenVerifyRpcResp.Member.OrganizationCode)
		ctx.Next()
	}
}

func getIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}
