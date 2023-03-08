package user

import (
	"github.com/Kirov7/project-api/config"
	"github.com/Kirov7/project-common/discovery"
	"github.com/Kirov7/project-common/logs"
	"github.com/Kirov7/project-grpc/user/login"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
)

var LoginServiceClient login.LoginServiceClient

func InitRpcUserClient() {
	etcdRegister := discovery.NewResolver(config.AppConf.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	conn, err := grpc.Dial("etcd:///user", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	LoginServiceClient = login.NewLoginServiceClient(conn)
}
