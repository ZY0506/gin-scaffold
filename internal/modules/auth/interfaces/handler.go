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
// @Summary      发送邮箱验证码
// @Description  发送验证码到指定邮箱，验证码有效期 5 分钟
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Param        req body application.SendCodeReq true "邮箱地址"
// @Success      200 {object} response.Response
// @Router       /auth/send-code [post]
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
// @Summary      用户注册
// @Description  通过邮箱验证码注册新用户
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Param        req body application.RegisterReq true "注册信息"
// @Success      200 {object} response.Response
// @Router       /auth/register [post]
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
// @Summary      用户登录
// @Description  通过用户名/邮箱 + 密码登录，返回双 Token
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Param        req body application.LoginReq true "登录信息"
// @Success      200 {object} response.Response{data=application.AuthTokenResp}
// @Router       /auth/login [post]
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
// @Summary      刷新双 Token
// @Description  使用 Refresh Token 获取新的 Access/Refresh Token 对
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Param        req body application.RefreshTokenReq true "刷新令牌"
// @Success      200 {object} response.Response{data=application.AuthTokenResp}
// @Router       /auth/refresh [post]
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
// @Summary      用户登出
// @Description  将当前 Access Token 加入黑名单
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /auth/logout [post]
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
// @Summary      重置密码
// @Description  通过邮箱验证码重置密码，无需登录
// @Tags         认证模块
// @Accept       json
// @Produce      json
// @Param        req body application.ResetPasswordReq true "重置密码信息"
// @Success      200 {object} response.Response
// @Router       /auth/reset-password [post]
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
