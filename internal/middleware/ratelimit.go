package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/ZY0506/gin-scaffold/config"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// RateLimit 单机令牌桶限流中间件
// 基于 golang.org/x/time/rate 实现，适用于单实例部署
// 多实例部署请使用 RateLimitRedis
func RateLimit(cfg *config.RateLimitConfig) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, errors.ErrRateLimit, "请求频率过高，请稍后再试")
			return
		}
		c.Next()
	}
}
