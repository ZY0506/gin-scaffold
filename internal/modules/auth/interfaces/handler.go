package interfaces

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/middleware"
	"github.com/ZY0506/gin-scaffold/internal/modules/auth/application"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// AuthHandler HTTP 处理器，处理认证相关的 HTTP 请求
type AuthHandler struct {
	svc *application.AuthService
}

// NewAuthHandler 创建认证 HTTP 处理器
func NewAuthHandler(svc *application.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// SendCode 发送验证码
// POST /api/v1/auth/send-code
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req application.SendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	if err := h.svc.SendCode(c.Request.Context(), req.Email); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req application.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	if err := h.svc.Register(c.Request.Context(), &req); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	ip := c.ClientIP()
	resp, err := h.svc.Login(c.Request.Context(), &req, ip)
	if err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新双 Token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req application.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, resp)
}

// Logout 登出
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	jti := middleware.GetJTI(c)
	if jti == "" {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到令牌标识")
		return
	}

	if err := h.svc.Logout(c.Request.Context(), jti); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// ResetPassword 重置密码
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req application.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// extractCode 从 error 中提取业务错误码
func extractCode(err error) int {
	var bizErr *bizErrors.Error
	if errors.As(err, &bizErr) {
		return bizErr.Code
	}
	return bizErrors.ErrInternal
}

// extractMsg 从 error 中提取用户可读的错误消息
func extractMsg(err error) string {
	var bizErr *bizErrors.Error
	if errors.As(err, &bizErr) {
		return bizErr.Msg
	}
	return "internal server error"
}
