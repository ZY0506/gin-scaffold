package interfaces

import (
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户个人中心路由（需要登录）
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler, authMiddleware gin.HandlerFunc) {
	user := r.Group("/user")
	user.Use(authMiddleware)
	{
		user.POST("/change-password", handler.ChangePassword)
		user.PUT("/profile", handler.UpdateProfile)
		user.GET("/profile", handler.GetUserInfo)
		user.POST("/avatar", handler.UploadAvatar)
		user.DELETE("/account", handler.DeleteAccount)
	}
}

type AdminRouter struct {
	handler *UserHandler
}

func NewAdminRouter(handler *UserHandler) *AdminRouter {
	return &AdminRouter{handler: handler}
}

// RegisterAdminRoutes 注册管理端用户路由
// 接收已配置好中间件的 admin 路由组
func (r *AdminRouter) RegisterAdminRoutes(admin *gin.RouterGroup) {
	admin.GET("/users", r.handler.List)
	admin.GET("/users/:id", r.handler.GetByID)
	admin.POST("/users", r.handler.Create)
	admin.PUT("/users/:id", r.handler.Update)
	admin.PATCH("/users/:id/status", r.handler.ToggleStatus)
}
