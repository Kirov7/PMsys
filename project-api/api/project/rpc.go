package project

import (
	"github.com/Kirov7/project-api/config"
	"github.com/Kirov7/project-common/discovery"
	"github.com/Kirov7/project-common/logs"
	accountRpc "github.com/Kirov7/project-grpc/project/account"
	authRpc "github.com/Kirov7/project-grpc/project/auth"
	departmentRpc "github.com/Kirov7/project-grpc/project/department"
	menuRpc "github.com/Kirov7/project-grpc/project/menu"
	projectRpc "github.com/Kirov7/project-grpc/project/project"
	taskRpc "github.com/Kirov7/project-grpc/project/task"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
)

var ProjectServiceClient projectRpc.ProjectServiceClient
var TaskServiceClient taskRpc.TaskServiceClient
var AccountServiceClient accountRpc.AccountServiceClient
var DepartmentServiceClient departmentRpc.DepartmentServiceClient
var AuthServiceClient authRpc.AuthServiceClient
var MenuServiceClient menuRpc.MenuServiceClient

func InitRpcProjectClient() {
	etcdRegister := discovery.NewResolver(config.AppConf.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	conn, err := grpc.Dial("etcd:///project", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	ProjectServiceClient = projectRpc.NewProjectServiceClient(conn)
	TaskServiceClient = taskRpc.NewTaskServiceClient(conn)
	AccountServiceClient = accountRpc.NewAccountServiceClient(conn)
	DepartmentServiceClient = departmentRpc.NewDepartmentServiceClient(conn)
	AuthServiceClient = authRpc.NewAuthServiceClient(conn)
	MenuServiceClient = menuRpc.NewMenuServiceClient(conn)
}
