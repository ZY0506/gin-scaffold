package interfaces

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/middleware"
	"github.com/ZY0506/gin-scaffold/internal/modules/user/application"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

// ChangePassword 修改密码
// POST /api/v1/user/change-password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	var req application.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	if err := h.svc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// UpdateProfile 修改个人信息
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	var req application.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, nil)
}

// GetUserInfo 获取个人信息
// GET /api/v1/user/profile
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	resp, err := h.svc.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	response.Success(c, resp)
}

// DeleteAccount 注销账户
// DELETE /api/v1/user/account
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	if err := h.svc.DeleteAccount(c.Request.Context(), userID); err != nil {
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

type UserHandler struct {
	svc *application.UserService
}

func NewUserHandler(svc *application.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List 用户列表
func (h *UserHandler) List(c *gin.Context) {
	var req application.UserListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	users, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, bizErrors.ErrInternal, "查询失败")
		return
	}

	response.Page(c, users, total, int64(req.Page), int64(req.PageSize))
}

// GetByID 用户详情
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "无效的用户ID")
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
			response.Error(c, 404, bizErrors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, bizErrors.ErrInternal, "查询失败")
		return
	}

	response.Success(c, user)
}

// Create 创建用户（管理员）
func (h *UserHandler) Create(c *gin.Context) {
	var req application.AdminCreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.CreateByAdmin(c.Request.Context(), &req)
	if err != nil {
		if bizErrors.IsCode(err, bizErrors.ErrUsernameExist) || bizErrors.IsCode(err, bizErrors.ErrEmailExist) {
			response.Error(c, 400, err.(*bizErrors.Error).Code, err.Error())
			return
		}
		response.Error(c, 500, bizErrors.ErrInternal, "创建失败")
		return
	}

	response.Success(c, user)
}

// Update 更新用户（管理员）
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "无效的用户ID")
		return
	}

	var req application.AdminUpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.UpdateByAdmin(c.Request.Context(), uint(id), &req)
	if err != nil {
		if bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
			response.Error(c, 404, bizErrors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, bizErrors.ErrInternal, "更新失败")
		return
	}

	response.Success(c, user)
}

// ToggleStatus 启用/禁用用户
func (h *UserHandler) ToggleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "无效的用户ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, bizErrors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.ToggleStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		if bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
			response.Error(c, 404, bizErrors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, bizErrors.ErrInternal, "操作失败")
		return
	}

	response.Success(c, nil)
}
