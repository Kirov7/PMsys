package project

import (
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var ignores = []string{
	"project/login/register",
	"project/login",
	"project/login/getCaptcha",
	"project/organization",
	"project/auth/apply"}

func Auth() func(*gin.Context) {
	return func(c *gin.Context) {
		//zap.L().Info("开始做授权认证")
		result := &common.Result{}
		//当用户登录认证通过，获取到用户信息，查询用户权限所拥有的节点信息
		//根据请求的uri路径 进行匹配
		uri := c.Request.RequestURI
		a := NewAuth()
		nodes, err := a.GetAuthNodes(c)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}
		for _, v := range ignores {
			if strings.Contains(uri, v) {
				c.Next()
				return
			}
		}
		for _, v := range nodes {
			if strings.Contains(uri, v) {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusOK, result.Fail(403, "无操作权限"))
		c.Abort()
	}
}
