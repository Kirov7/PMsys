package interceptor

import (
	"context"
	"encoding/json"
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-grpc/user/login"
	"github.com/Kirov7/project-user/internal/dao"
	"github.com/Kirov7/project-user/internal/repo"
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
	cacheMap["/login.LoginService/OrgList"] = &login.OrgListResponse{}
	cacheMap["/login.LoginService/FindMemInfoById"] = &login.MemberMessage{}
	return &CacheInterceptor{cache: dao.Rc, cacheMap: cacheMap}
}

func (i *CacheInterceptor) CacheInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		return
	})
}
