package repo

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key, value string, expire time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	HSet(ctx context.Context, key string, field string, value string) error
	HKeys(ctx context.Context, key string) ([]string, error)
}
