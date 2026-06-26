package interfaces

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册管理端公开路由（无需登录）
func RegisterPublicRoutes(r *gin.RouterGroup, h *AdminHandler) {
	r.POST("/admin/login", h.Login)
}

// AdminRouter 管理端受保护路由（需 JWT + Casbin）
type AdminRouter struct {
	handler *AdminHandler
}

func NewAdminRouter(handler *AdminHandler) *AdminRouter {
	return &AdminRouter{handler: handler}
}

// RegisterAdminRoutes 注册所有管理端受保护路由
// 接收已配置好中间件的 admin 路由组
func (r *AdminRouter) RegisterAdminRoutes(admin *gin.RouterGroup) {
	// 管理员管理
	admin.GET("/admins", r.handler.List)
	admin.GET("/admins/:id", r.handler.GetByID)
	admin.POST("/admins", r.handler.Create)
	admin.PUT("/admins/:id", r.handler.Update)

	// 操作日志
	admin.GET("/operation-logs", r.handler.ListOperationLogs)
}
