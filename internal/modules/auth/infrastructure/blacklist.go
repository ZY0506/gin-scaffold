package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

const blacklistKeyPrefix = "token_blacklist:"

type RedisTokenBlacklist struct {
	rdb *redis.Client
}

func NewRedisTokenBlacklist(rdb *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{rdb: rdb}
}

// Add 将 JTI 加入黑名单，TTL 等于 Token 剩余有效期
func (b *RedisTokenBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	key := blacklistKeyPrefix + jti
	return b.rdb.Set(ctx, key, 1, ttl).Err()
}

// Exists 检查 JTI 是否在黑名单中
func (b *RedisTokenBlacklist) Exists(ctx context.Context, jti string) (bool, error) {
	key := blacklistKeyPrefix + jti
	n, err := b.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, bizErrors.Wrap(err, bizErrors.ErrRedis, "检查令牌黑名单失败")
	}
	return n > 0, nil
}
