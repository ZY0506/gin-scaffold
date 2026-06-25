package interfaces

import (
	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/modules/user/application"
)

// RegisterUserRoutes 注册用户个人中心路由（需要登录）
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler, authMiddleware gin.HandlerFunc) {
	user := r.Group("/user")
	user.Use(authMiddleware)
	{
		user.POST("/change-password", handler.ChangePassword)
		user.PUT("/profile", handler.UpdateProfile)
		user.GET("/profile", handler.GetUserInfo)
		user.DELETE("/account", handler.DeleteAccount)
	}
}

// RegisterUserRoutesWithDI 便捷方法
func RegisterUserRoutesWithDI(r *gin.RouterGroup, svc *application.UserService, authMiddleware gin.HandlerFunc) {
	handler := NewUserHandler(svc)
	RegisterUserRoutes(r, handler, authMiddleware)
}

type AdminRouter struct {
	handler  *UserHandler
	authMW   gin.HandlerFunc
	casbinMW gin.HandlerFunc
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
