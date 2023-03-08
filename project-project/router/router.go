package router

import (
	"github.com/Kirov7/project-common/discovery"
	"github.com/Kirov7/project-common/logs"
	accountRpc "github.com/Kirov7/project-grpc/project/account"
	authRpc "github.com/Kirov7/project-grpc/project/auth"
	departmentRpc "github.com/Kirov7/project-grpc/project/department"
	menuRpc "github.com/Kirov7/project-grpc/project/menu"
	projectRpc "github.com/Kirov7/project-grpc/project/project"
	taskRpc "github.com/Kirov7/project-grpc/project/task"
	"github.com/Kirov7/project-project/config"
	rpc "github.com/Kirov7/project-project/internal/rpc"
	accountServiceV1 "github.com/Kirov7/project-project/pkg/service/account.service.v1"
	authServiceV1 "github.com/Kirov7/project-project/pkg/service/auth.service.v1"
	departmentServiceV1 "github.com/Kirov7/project-project/pkg/service/department.service.v1"
	menuServiceV1 "github.com/Kirov7/project-project/pkg/service/menu.service.v1"
	projectServiceV1 "github.com/Kirov7/project-project/pkg/service/project.service.v1"
	taskServiceV1 "github.com/Kirov7/project-project/pkg/service/task.service.v1"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
)

type Router interface {
	Route(r *gin.Engine)
}

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
			projectRpc.RegisterProjectServiceServer(g, projectServiceV1.NewProjectService())
			taskRpc.RegisterTaskServiceServer(g, taskServiceV1.NewTaskService())
			accountRpc.RegisterAccountServiceServer(g, accountServiceV1.NewAccountService())
			departmentRpc.RegisterDepartmentServiceServer(g, departmentServiceV1.NewDepartmentService())
			authRpc.RegisterAuthServiceServer(g, authServiceV1.NewAuthService())
			menuRpc.RegisterMenuServiceServer(g, menuServiceV1.NewMenuService())
		},
	}
	//cache := interceptor.NewCacheInterceptor().CacheInterceptor()
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
	r := discovery.NewRegister(config.AppConf.EtcdConfig.Addrs, logs.LG)
	_, err := r.Register(info, 2)
	if err != nil {
		log.Fatalln(err)
	}
}

func InitUserRpcClient() {
	rpc.InitRpcUserClient()
}
