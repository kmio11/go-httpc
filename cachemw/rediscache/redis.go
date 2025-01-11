package rediscache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/kmio11/go-httpc/cachemw"
)

var _ cachemw.CacheWithTTL = (*RedisCache)(nil)

type RedisCache struct {
	cache *cache.Cache
}

type Option func(*RedisCache)

func New(redisCache *cache.Cache, opts ...Option) *RedisCache {
	r := &RedisCache{
		cache: redisCache,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	var b []byte
	if err := r.cache.Get(ctx, key, &b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *RedisCache) set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	return r.set(ctx, key, value, 0)
}

func (r *RedisCache) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.set(ctx, key, value, ttl)
}
