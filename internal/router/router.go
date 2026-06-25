package router

import (
	"github.com/gin-gonic/gin"

	authHandler "github.com/ZY0506/gin-scaffold/internal/modules/auth/interfaces"
	blHandler "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/interfaces"
	userHandler "github.com/ZY0506/gin-scaffold/internal/modules/user/interfaces"
)

// Register 注册所有路由
func Register(
	r *gin.Engine,
	authH *authHandler.AuthHandler,
	authMW gin.HandlerFunc,
	userH *userHandler.UserHandler,
	userAdminRouter *userHandler.AdminRouter,
	blAdminRouter *blHandler.AdminRouter,
) {
	// API v1 路由组
	v1 := r.Group("/api/v1")

	// 认证模块（公开 + 需要登录的认证操作）
	authHandler.RegisterAuthRoutes(v1, authH, authMW)

	// 用户模块（个人中心，需要登录）
	userHandler.RegisterUserRoutes(v1, userH, authMW)

	// 用户管理端路由（JWT + Casbin）
	userAdminRouter.RegisterAdminRoutes(v1)

	// 黑名单管理端路由（JWT + Casbin）
	blAdminRouter.RegisterAdminRoutes(v1)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
