package interfaces

import "github.com/gin-gonic/gin"

type AdminRouter struct {
	handler *BlacklistHandler
}

func NewAdminRouter(handler *BlacklistHandler) *AdminRouter {
	return &AdminRouter{handler: handler}
}

// RegisterAdminRoutes 注册管理端黑名单路由
// 接收已配置好中间件的 admin 路由组
func (r *AdminRouter) RegisterAdminRoutes(admin *gin.RouterGroup) {
	admin.POST("/blacklist", r.handler.Create)
	admin.GET("/blacklist", r.handler.List)
	admin.PUT("/blacklist/:id", r.handler.Update)
	admin.DELETE("/blacklist/:id", r.handler.Deactivate)
}
