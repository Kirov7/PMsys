package dao

import (
	"context"
	"github.com/Kirov7/project-user/config"
	"github.com/go-redis/redis/v8"
	"time"
)

var Rc *RedisCache

type RedisCache struct {
	rdb *redis.Client
}

func init() {
	rdb := redis.NewClient(config.AppConf.ReadRedisOptions())
	Rc = &RedisCache{rdb: rdb}
}

func (rc *RedisCache) Set(ctx context.Context, key, value string, expire time.Duration) error {
	err := rc.rdb.Set(ctx, key, value, expire).Err()
	return err
}

func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.rdb.Get(ctx, key).Result()
	return result, err
}

func (rc *RedisCache) HSet(ctx context.Context, key string, field string, value string) error {
	err := rc.rdb.HSet(ctx, key, field, value).Err()
	return err
}

func (rc *RedisCache) HKeys(ctx context.Context, key string) ([]string, error) {
	result, err := rc.rdb.HKeys(ctx, key).Result()
	return result, err
}

func (rc *RedisCache) Delete(ctx context.Context, keys []string) error {
	err := rc.rdb.Del(ctx, keys...).Err()
	return err
}
