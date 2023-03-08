package router

import (
	"github.com/gin-gonic/gin"
)

type Router interface {
	Route(r *gin.Engine)
}

// RegisterRouter (一般不建议结构体前缀包含包名)
type RegisterRouter struct {
}

func NewRegisterRouter() *RegisterRouter {
	return &RegisterRouter{}
}

func (*RegisterRouter) Route(ro Router, r *gin.Engine) {
	ro.Route(r)
}

var routers []Router

func InitRouter(r *gin.Engine) {
	//rg := New()
	//rg.Route(&user.RouterUser{}, r)
	for _, ro := range routers {
		ro.Route(r)
	}
}

func Register(ro ...Router) {
	routers = append(routers, ro...)
}
