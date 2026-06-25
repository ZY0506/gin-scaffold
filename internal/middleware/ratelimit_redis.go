package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/ZY0506/gin-scaffold/config"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// RateLimitRedis 分布式限流中间件（基于 Redis 滑动窗口）
// 使用 Redis Sorted Set + Lua 脚本保证原子性，适用于多实例部署
func RateLimitRedis(cfg *config.RateLimitConfig, rdb *redis.Client) gin.HandlerFunc {
	// Lua 脚本：滑动窗口限流
	// KEYS[1]       - 限流 key（按 IP）
	// ARGV[1]       - 窗口大小（秒）
	// ARGV[2]       - 窗口内最大请求数
	// ARGV[3]       - 当前时间戳（毫秒）
	script := redis.NewScript(`
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		redis.call("zremrangebyscore", key, 0, now - window * 1000)
		local count = redis.call("zcard", key)

		if count < limit then
			redis.call("zadd", key, now, now)
			redis.call("expire", key, window)
			return 1
		end

		return 0
	`)

	return func(c *gin.Context) {
		key := "ratelimit:" + c.ClientIP()
		now := time.Now().UnixMilli()
		window := int64(2) // 2 秒滑动窗口，允许 Rate 个请求
		limit := int64(cfg.Rate)

		allowed, err := script.Run(c.Request.Context(), rdb, []string{key},
			window, limit, now).Int()
		if err != nil || allowed == 0 {
			response.Error(c, http.StatusTooManyRequests, errors.ErrRateLimit, "请求频率过高，请稍后再试")
			return
		}
		c.Next()
	}
}
