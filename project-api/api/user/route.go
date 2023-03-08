package user

import (
	"github.com/Kirov7/project-api/router"
	"github.com/gin-gonic/gin"
	"log"
)

func init() {
	log.Println("init user router")
	router.Register(&RouterUser{})
}

type RouterUser struct {
}

func (*RouterUser) Route(r *gin.Engine) {
	// 初始化grpc的客户端连接
	InitRpcUserClient()
	h := NewHandlerUser()
	r.POST("/project/login/getCaptcha", h.getCaptcha)
	r.POST("/project/login/register", h.register)
	r.POST("/project/login", h.login)
	r.POST("/project/organization/_getOrgList", h.orgList)
}
