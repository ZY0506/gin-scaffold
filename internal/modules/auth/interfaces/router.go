package interfaces

import (
	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/modules/auth/application"
)

// RegisterAuthRoutes 注册认证模块路由
// 参数:
//   - r: API 路由组 (如 /api/v1)
//   - handler: 认证 HTTP 处理器
//   - authMiddleware: JWT 认证中间件，用于保护需要登录的路由
func RegisterAuthRoutes(r *gin.RouterGroup, handler *AuthHandler, authMiddleware gin.HandlerFunc) {
	// ===== 公开接口（无需登录） =====
	auth := r.Group("/auth")
	{
		// 发送验证码
		auth.POST("/send-code", handler.SendCode)
		// 用户注册
		auth.POST("/register", handler.Register)
		// 用户登录
		auth.POST("/login", handler.Login)
		// 刷新 Token
		auth.POST("/refresh", handler.RefreshToken)
		// 重置密码
		auth.POST("/reset-password", handler.ResetPassword)
	}

	// ===== 需要登录的接口 =====
	authProtected := r.Group("/auth").Use(authMiddleware)
	{
		// 登出
		authProtected.POST("/logout", handler.Logout)
	}
}

// RegisterAuthRoutesWithDI 带依赖注入的路由注册（便捷方法，在路由层完成依赖组装）
// 适用于 Wire / 手动 DI 的场景
func RegisterAuthRoutesWithDI(
	r *gin.RouterGroup,
	svc *application.AuthService,
	authMiddleware gin.HandlerFunc,
) {
	handler := NewAuthHandler(svc)
	RegisterAuthRoutes(r, handler, authMiddleware)
}
