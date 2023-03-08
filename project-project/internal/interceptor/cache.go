package interceptor

import (
	"context"
	"encoding/json"
	"github.com/Kirov7/project-common/encrypts"
	projectRpc "github.com/Kirov7/project-grpc/project/project"
	"github.com/Kirov7/project-project/internal/dao"
	"github.com/Kirov7/project-project/internal/repo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type CacheInterceptor struct {
	cache    repo.Cache
	cacheMap map[string]any
}

func NewCacheInterceptor() *CacheInterceptor {
	cacheMap := make(map[string]any)
	cacheMap["/project.ProjectService/FindProjectByMemId"] = &projectRpc.ProjectResponse{}
	return &CacheInterceptor{cache: dao.Rc, cacheMap: cacheMap}
}

func (i *CacheInterceptor) CacheInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		i = NewCacheInterceptor()
		respType := i.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		respJson, err := i.cache.Get(c, info.FullMethod+"::"+cacheKey)
		if err == nil && respJson != "" {
			json.Unmarshal([]byte(respJson), &respType)
			zap.L().Info(info.FullMethod + ":  used cache")
			return respType, nil
		}
		resp, err = handler(ctx, req)
		if resp == nil {
			return
		}
		bytes, _ := json.Marshal(resp)
		i.cache.Set(c, info.FullMethod+"::"+cacheKey, string(bytes), 5*time.Minute)
		zap.L().Info(info.FullMethod + ":  added to cache")
		/*
			if strings.HasPrefix(info.FullMethod, "/task") {
				i.cache.HSet(c, "task", info.FullMethod+"::"+cacheKey,"")
			}
		*/

		return
	})
}
