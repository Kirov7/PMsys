package main

import (
	_ "github.com/Kirov7/project-api/api/project"
	_ "github.com/Kirov7/project-api/api/user"
	"github.com/Kirov7/project-api/config"
	"github.com/Kirov7/project-api/router"
	srv "github.com/Kirov7/project-common"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	r.StaticFS("/upload", http.Dir("upload"))
	// 路由注册
	router.InitRouter(r)

	// 开启pprof后 默认的访问路径事/debug/pprof
	pprof.Register(r)

	srv.Run(r, config.AppConf.ServerConfig.Name, config.AppConf.ServerConfig.Addr, nil)
}
