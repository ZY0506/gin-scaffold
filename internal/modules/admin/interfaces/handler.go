package interfaces

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/modules/admin/application"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

type AdminHandler struct {
	svc      *application.AdminService
	opLogSvc *application.OperationLogService
	jwt      JWTService
}

type JWTService interface {
	GeneratePair(userID uint, role string) (accessToken, refreshToken string, err error)
}

func NewAdminHandler(svc *application.AdminService, opLogSvc *application.OperationLogService, jwt JWTService) *AdminHandler {
	return &AdminHandler{svc: svc, opLogSvc: opLogSvc, jwt: jwt}
}

// Login 管理员登录
// @Summary      管理员登录
// @Description  管理员通过用户名密码登录，返回双 Token
// @Tags         管理端-管理员
// @Accept       json
// @Produce      json
// @Param        req body application.AdminLoginReq true "登录信息"
// @Success      200 {object} response.Response
// @Router       /admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req application.AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	adminID, _, err := h.svc.Login(c.Request.Context(), &req, c.ClientIP())
	if err != nil {
		response.Fail(c, extractCode(err), extractMsg(err))
		return
	}

	accessToken, refreshToken, err := h.jwt.GeneratePair(adminID, "admin")
	if err != nil {
		response.Error(c, 500, errors.ErrInternal, "令牌生成失败")
		return
	}

	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// List 管理员列表
// @Summary      管理员列表
// @Description  分页查询所有管理员账号
// @Tags         管理端-管理员
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页条数" default(20)
// @Success      200 {object} response.PageData{list=[]application.AdminItemResp}
// @Router       /admin/admins [get]
func (h *AdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Page(c, list, total, int64(page), int64(pageSize))
}

// GetByID 管理员详情
// @Summary      管理员详情
// @Description  获取指定管理员的详细信息
// @Tags         管理端-管理员
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "管理员ID"
// @Success      200 {object} response.Response{data=application.AdminItemResp}
// @Router       /admin/admins/{id} [get]
func (h *AdminHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的ID")
		return
	}

	admin, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.IsCode(err, errors.ErrAdminNotFound) {
			response.Error(c, 404, errors.ErrAdminNotFound, "管理员不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Success(c, admin)
}

// Create 创建管理员
// @Summary      创建管理员
// @Description  创建新的管理员账号
// @Tags         管理端-管理员
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req body application.AdminCreateReq true "管理员信息"
// @Success      200 {object} response.Response{data=application.AdminItemResp}
// @Router       /admin/admins [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var req application.AdminCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	admin, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.IsCode(err, errors.ErrAdminUsernameExist) {
			response.Error(c, 400, errors.ErrAdminUsernameExist, "用户名已存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "创建失败")
		return
	}

	response.Success(c, admin)
}

// Update 修改管理员
// @Summary      修改管理员
// @Description  修改管理员昵称、密码、状态
// @Tags         管理端-管理员
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "管理员ID"
// @Param        req body application.AdminUpdateReq true "修改信息"
// @Success      200 {object} response.Response
// @Router       /admin/admins/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "无效的ID")
		return
	}

	var req application.AdminUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.Update(c.Request.Context(), uint(id), &req); err != nil {
		if errors.IsCode(err, errors.ErrAdminNotFound) {
			response.Error(c, 404, errors.ErrAdminNotFound, "管理员不存在")
			return
		}
		response.Error(c, 500, errors.ErrInternal, "更新失败")
		return
	}

	response.Success(c, nil)
}

// ListOperationLogs 操作日志列表
// @Summary      操作日志列表
// @Description  分页查询管理员操作日志
// @Tags         管理端-管理员
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页条数" default(20)
// @Success      200 {object} response.PageData{list=[]domain.OperationLog}
// @Router       /admin/operation-logs [get]
func (h *AdminHandler) ListOperationLogs(c *gin.Context) {
	var req application.OperationLogListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errors.ErrBadRequest, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.opLogSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 500, errors.ErrInternal, "查询失败")
		return
	}

	response.Page(c, list, total, int64(req.Page), int64(req.PageSize))
}

func extractCode(err error) int {
	var bizErr *errors.Error
	if as, ok := err.(*errors.Error); ok {
		bizErr = as
	}
	if bizErr != nil {
		return bizErr.Code
	}
	return errors.ErrInternal
}

func extractMsg(err error) string {
	var bizErr *errors.Error
	if as, ok := err.(*errors.Error); ok {
		bizErr = as
	}
	if bizErr != nil {
		return bizErr.Msg
	}
	return "internal server error"
}
