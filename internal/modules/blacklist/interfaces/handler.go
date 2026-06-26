package interfaces

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/middleware"
	"github.com/ZY0506/gin-scaffold/internal/modules/blacklist/application"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

type BlacklistHandler struct {
	svc *application.BlacklistService
}

func NewBlacklistHandler(svc *application.BlacklistService) *BlacklistHandler {
	return &BlacklistHandler{svc: svc}
}

// Create 添加黑名单
func (h *BlacklistHandler) Create(c *gin.Context) {
	var req application.AddBlacklistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)
	if err := h.svc.Add(c.Request.Context(), &req, operatorID); err != nil {
		code := errors.ErrInternal
		if errors.IsCode(err, errors.ErrBadRequest) {
			code = errors.ErrBadRequest
		}
		response.Error(c, 400, code, err.Error())
		return
	}

	response.Success(c, nil)
}

// List 黑名单列表
func (h *BlacklistHandler) List(c *gin.Context) {
	var req application.BlacklistListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Page(c, list, total, int64(req.Page), int64(req.PageSize))
}

// Deactivate 解封
func (h *BlacklistHandler) Deactivate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的ID")
		return
	}

	if err := h.svc.Deactivate(c.Request.Context(), uint(id)); err != nil {
		if errors.IsCode(err, errors.ErrBlacklisted) {
			response.Error(c, 404, errors.ErrBlacklisted, "记录不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "操作失败")
		return
	}

	response.Success(c, nil)
}

// Update 修改黑名单记录
func (h *BlacklistHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的ID")
		return
	}

	var req application.UpdateBlacklistReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), uint(id), &req); err != nil {
		if errors.IsCode(err, errors.ErrBlacklisted) {
			response.Error(c, 404, errors.ErrBlacklisted, "记录不存在")
			return
		}
		if errors.IsCode(err, errors.ErrBadRequest) {
			response.Error(c, 400, errors.ErrBadRequest, err.Error())
			return
		}
		response.Error(c, 500, errors.ErrInternal, "操作失败")
		return
	}

	response.Success(c, nil)
}
