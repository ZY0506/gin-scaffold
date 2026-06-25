package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

const codeKeyPrefix = "verify_code:"

type RedisCodeStore struct {
	rdb *redis.Client
}

func NewRedisCodeStore(rdb *redis.Client) *RedisCodeStore {
	return &RedisCodeStore{rdb: rdb}
}

func (s *RedisCodeStore) key(email string) string {
	return codeKeyPrefix + email
}

func (s *RedisCodeStore) Set(ctx context.Context, email, code string, ttl int) error {
	return s.rdb.Set(ctx, s.key(email), code, time.Duration(ttl)*time.Second).Err()
}

func (s *RedisCodeStore) Get(ctx context.Context, email string) (string, error) {
	code, err := s.rdb.Get(ctx, s.key(email)).Result()
	if err == redis.Nil {
		return "", nil // 不存在或已过期
	}
	if err != nil {
		return "", bizErrors.Wrap(err, bizErrors.ErrRedis, "获取验证码失败")
	}
	return code, nil
}

func (s *RedisCodeStore) Del(ctx context.Context, email string) error {
	return s.rdb.Del(ctx, s.key(email)).Err()
}
