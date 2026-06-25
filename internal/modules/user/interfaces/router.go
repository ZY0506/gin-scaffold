package interfaces

import (
	"github.com/gin-gonic/gin"
)

type AdminRouter struct {
	handler      *UserHandler
	authMW       gin.HandlerFunc
	casbinMW     gin.HandlerFunc
}

func NewAdminRouter(handler *UserHandler, authMW, casbinMW gin.HandlerFunc) *AdminRouter {
	return &AdminRouter{
		handler:  handler,
		authMW:   authMW,
		casbinMW: casbinMW,
	}
}

// RegisterAdminRoutes 注册管理端用户路由
// 需要 JWT 认证 + Casbin RBAC 鉴权
func (r *AdminRouter) RegisterAdminRoutes(group *gin.RouterGroup) {
	admin := group.Group("/admin")
	admin.Use(r.authMW, r.casbinMW)
	{
		admin.GET("/users", r.handler.List)
		admin.GET("/users/:id", r.handler.GetByID)
		admin.POST("/users", r.handler.Create)
		admin.PUT("/users/:id", r.handler.Update)
		admin.PATCH("/users/:id/status", r.handler.ToggleStatus)
	}
}
