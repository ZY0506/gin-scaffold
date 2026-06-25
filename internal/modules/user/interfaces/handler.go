package interfaces

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/modules/user/application"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

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
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	users, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Page(c, users, total, int64(req.Page), int64(req.PageSize))
}

// GetByID 用户详情
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的用户ID")
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.IsCode(err, errors.ErrUserNotFound) {
			response.Error(c, 404, errors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Success(c, user)
}

// Create 创建用户（管理员）
func (h *UserHandler) Create(c *gin.Context) {
	var req application.AdminCreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.CreateByAdmin(c.Request.Context(), &req)
	if err != nil {
		if errors.IsCode(err, errors.ErrUsernameExist) || errors.IsCode(err, errors.ErrEmailExist) {
			response.Error(c, 400, err.(*errors.Error).Code, err.Error())
			return
		}
		response.Error(c, 500, errors.ErrInternal, "创建失败")
		return
	}

	response.Success(c, user)
}

// Update 更新用户（管理员）
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的用户ID")
		return
	}

	var req application.AdminUpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.UpdateByAdmin(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.IsCode(err, errors.ErrUserNotFound) {
			response.Error(c, 404, errors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "更新失败")
		return
	}

	response.Success(c, user)
}

// ToggleStatus 启用/禁用用户
func (h *UserHandler) ToggleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的用户ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.ToggleStatus(c.Request.Context(), uint(id), req.Status); err != nil {
		if errors.IsCode(err, errors.ErrUserNotFound) {
			response.Error(c, 404, errors.ErrUserNotFound, "用户不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "操作失败")
		return
	}

	response.Success(c, nil)
}
