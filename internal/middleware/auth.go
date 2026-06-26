package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authInfra "github.com/ZY0506/gin-scaffold/internal/modules/auth/infrastructure"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// JWTService JWT 鉴权服务接口
// 适配 auth/infrastructure.JWTService
type JWTService interface {
	ValidateToken(tokenString string) (*authInfra.Claims, error)
}

// TokenBlacklist Token 黑名单接口
// 适配 auth/infrastructure.RedisTokenBlacklist
type TokenBlacklist interface {
	Exists(ctx context.Context, jti string) (bool, error)
}

// JWTAuth JWT 鉴权中间件
// 从 Authorization Header 提取 Bearer Token，校验后将用户信息注入 gin.Context
// 注入的键值：user_id (uint), username (string), role (string), jti (string)
func JWTAuth(jwtService JWTService, blacklist TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, bizErrors.ErrUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, http.StatusUnauthorized, bizErrors.ErrTokenInvalid, "无效的认证格式")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 JWT
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			if bizErrors.IsCode(err, bizErrors.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, bizErrors.ErrTokenExpired, "令牌已过期")
			} else {
				response.Error(c, http.StatusUnauthorized, bizErrors.ErrTokenInvalid, "无效的令牌")
			}
			c.Abort()
			return
		}

		// 检查黑名单
		if blacklist != nil {
			exists, err := blacklist.Exists(c.Request.Context(), claims.ID)
			if err != nil {
				response.Error(c, http.StatusUnauthorized, bizErrors.ErrTokenBlacklisted, "令牌校验失败")
				c.Abort()
				return
			}
			if exists {
				response.Error(c, http.StatusUnauthorized, bizErrors.ErrTokenBlacklisted, "令牌已被注销")
				c.Abort()
				return
			}
		}

		// 注入用户信息到 Context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("jti", claims.ID)
		c.Next()
	}
}

// GetUserID 从 gin.Context 获取用户 ID
func GetUserID(c *gin.Context) uint {
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetRole 从 gin.Context 获取用户角色
func GetRole(c *gin.Context) string {
	if v, exists := c.Get("role"); exists {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

// GetJTI 从 gin.Context 获取 JWT ID
func GetJTI(c *gin.Context) string {
	if v, exists := c.Get("jti"); exists {
		if jti, ok := v.(string); ok {
			return jti
		}
	}
	return ""
}
