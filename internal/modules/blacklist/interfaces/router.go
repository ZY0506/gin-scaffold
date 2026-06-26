package interfaces

import "github.com/gin-gonic/gin"

type AdminRouter struct {
	handler  *BlacklistHandler
	authMW   gin.HandlerFunc
	casbinMW gin.HandlerFunc
}

func NewAdminRouter(handler *BlacklistHandler, authMW, casbinMW gin.HandlerFunc) *AdminRouter {
	return &AdminRouter{
		handler:  handler,
		authMW:   authMW,
		casbinMW: casbinMW,
	}
}

// RegisterAdminRoutes 注册管理端黑名单路由
func (r *AdminRouter) RegisterAdminRoutes(group *gin.RouterGroup) {
	admin := group.Group("/admin")
	admin.Use(r.authMW, r.casbinMW)
	{
		admin.POST("/blacklist", r.handler.Create)
		admin.GET("/blacklist", r.handler.List)
		admin.PUT("/blacklist/:id", r.handler.Update)
		admin.DELETE("/blacklist/:id", r.handler.Deactivate)
	}
}
