package main

import (
	srv "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-project/config"
	"github.com/Kirov7/project-project/router"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 路由注册
	router.InitRouter(r)
	// 初始化rpc客户端
	router.InitUserRpcClient()
	// grpc服务注册
	grpcServer := router.RegisterGrpc()
	// grpc服务注册到etcd
	router.RegisterEtcdServer()
	stop := func() {
		grpcServer.Stop()
	}
	srv.Run(r, config.AppConf.ServerConfig.Name, config.AppConf.ServerConfig.Addr, stop)
}
