package project

import (
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ProjectAuth() func(*gin.Context) {
	return func(c *gin.Context) {
		// 如果此用户不是项目成员直接返回无权限
		//zap.L().Info("项目操作授权认证")
		result := &common.Result{}
		//在接口有权限的基础上，做项目权限，不是这个项目的成员，无权限查看项目和操作项目
		//检查是否有projectCode和taskCode这两个参数
		isProjectAuth := false
		projectCode := c.PostForm("projectCode")
		if projectCode != "" {
			isProjectAuth = true
		}
		taskCode := c.PostForm("taskCode")
		if taskCode != "" {
			isProjectAuth = true
		}
		if isProjectAuth {
			p := NewHandlerProject()
			pr, isMember, isOwner, err := p.FindProjectByMemberId(c.GetInt64("memberId"), projectCode, taskCode)
			if err != nil {
				code, msg := errs.ParseGrpcError(err)
				c.JSON(http.StatusOK, result.Fail(code, msg))
				c.Abort()
				return
			}
			if !isMember {
				c.JSON(http.StatusOK, result.Fail(403, "不是项目成员，无操作权限"))
				c.Abort()
				return
			}
			if pr.Private == 1 {
				//私有项目
				if isOwner {
					c.Next()
					return
				} else {
					c.JSON(http.StatusOK, result.Fail(403, "私有项目，无操作权限"))
					c.Abort()
					return
				}
			}
		}
	}
}
