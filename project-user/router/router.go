package router

import (
	"github.com/Kirov7/project-common/discovery"
	"github.com/Kirov7/project-common/logs"
	"github.com/Kirov7/project-grpc/user/login"
	"github.com/Kirov7/project-user/config"
	loginServiceV1 "github.com/Kirov7/project-user/pkg/service/login.service.v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
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

type gRPCConfig struct {
	Addr         string
	RegisterFunc func(server *grpc.Server)
}

func RegisterGrpc() *grpc.Server {
	c := gRPCConfig{
		Addr: config.AppConf.GrpcConfig.Addr,
		RegisterFunc: func(g *grpc.Server) {
			login.RegisterLoginServiceServer(g, loginServiceV1.NewLoginService())
		},
	}
	//cacheInterceptor := interceptor.NewCacheInterceptor()
	//s := grpc.NewServer(cacheInterceptor.CacheInterceptor())
	s := grpc.NewServer()
	c.RegisterFunc(s)
	listen, err := net.Listen("tcp", c.Addr)
	if err != nil {
		log.Printf("listen port %s fail\n", c.Addr)
	}
	go func() {
		log.Printf("grpc server started as %s \n", c.Addr)
		err = s.Serve(listen)
		if err != nil {
			log.Printf("server started error: %s\n", err)
			return
		}
	}()
	return s
}

func RegisterEtcdServer() {
	etcdRegister := discovery.NewResolver(config.AppConf.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	info := discovery.Server{
		Name:    config.AppConf.GrpcConfig.Name,
		Addr:    config.AppConf.GrpcConfig.Addr,
		Version: config.AppConf.GrpcConfig.Version,
		Weight:  config.AppConf.GrpcConfig.Weight,
	}
	// 创建etcd注册器
	r := discovery.NewRegister(config.AppConf.EtcdConfig.Addrs, logs.LG)
	// 注册服务
	_, err := r.Register(info, 2)
	if err != nil {
		log.Fatalln(err)
	}
}
