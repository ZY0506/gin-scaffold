package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	adminHandler "github.com/ZY0506/gin-scaffold/internal/modules/admin/interfaces"
	authHandler "github.com/ZY0506/gin-scaffold/internal/modules/auth/interfaces"
	blHandler "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/interfaces"
	userHandler "github.com/ZY0506/gin-scaffold/internal/modules/user/interfaces"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// Register 注册所有路由
func Register(
	r *gin.Engine,
	authH *authHandler.AuthHandler,
	authMW gin.HandlerFunc,
	userH *userHandler.UserHandler,
	userAdminRouter *userHandler.AdminRouter,
	blAdminRouter *blHandler.AdminRouter,
	adminRouter *adminHandler.AdminRouter,
	adminPublicHandler *adminHandler.AdminHandler,
	adminMiddlewares ...gin.HandlerFunc,
) {
	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")

	// 认证模块（公开 + 需要登录的认证操作）
	authHandler.RegisterAuthRoutes(v1, authH, authMW)

	// 用户模块（个人中心，需要登录）
	userHandler.RegisterUserRoutes(v1, userH, authMW)

	// 管理端公开路由（无需登录）
	adminHandler.RegisterPublicRoutes(v1, adminPublicHandler)

	// 管理端受保护路由组（JWT + Casbin + 操作日志等中间件）
	admin := v1.Group("/admin")
	admin.Use(adminMiddlewares...)
	{
		// 用户管理
		userAdminRouter.RegisterAdminRoutes(admin)

		// 黑名单管理
		blAdminRouter.RegisterAdminRoutes(admin)

		// 管理员管理 + 操作日志
		adminRouter.RegisterAdminRoutes(admin)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 自定义 404
	r.NoRoute(func(c *gin.Context) {
		response.Error(c, 404, errors.ErrNotFound, "请求的资源不存在")
	})
}
