package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// CasbinEnforcer Casbin 权限检查接口
// 适配 casbin/v2.SyncedEnforcer / casbin/v2.Enforcer
type CasbinEnforcer interface {
	Enforce(rvals ...interface{}) (bool, error)
}

// CasbinRBAC Casbin RBAC 鉴权中间件
// 从 gin.Context 中提取 role（由 JWTAuth 中间件注入），
// 结合请求路径 (obj) 和 HTTP 方法 (act) 执行权限校验
// 需要先使用 JWTAuth 中间件注入 role
func CasbinRBAC(enforcer CasbinEnforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, bizErrors.ErrForbidden, "权限不足")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr == "" {
			response.Error(c, http.StatusForbidden, bizErrors.ErrForbidden, "权限不足")
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		allowed, err := enforcer.Enforce(roleStr, path, method)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, bizErrors.ErrInternal, "权限校验失败")
			c.Abort()
			return
		}

		if !allowed {
			response.Error(c, http.StatusForbidden, bizErrors.ErrForbidden, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
