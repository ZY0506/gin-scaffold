package interfaces

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ZY0506/gin-scaffold/internal/middleware"
	"github.com/ZY0506/gin-scaffold/internal/modules/user/application"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	"github.com/ZY0506/gin-scaffold/internal/pkg/response"
)

type UserHandler struct {
	svc         *application.UserService
	saveDir     string
	maxSize     int64
	allowedExts map[string]bool
}

func NewUserHandler(svc *application.UserService, saveDir string, maxSize int64, allowedExts []string) *UserHandler {
	exts := make(map[string]bool, len(allowedExts))
	for _, ext := range allowedExts {
		exts[strings.ToLower(ext)] = true
	}
	return &UserHandler{
		svc:         svc,
		saveDir:     saveDir,
		maxSize:     maxSize,
		allowedExts: exts,
	}
}

// ChangePassword 修改密码
// @Summary      修改密码
// @Description  验证旧密码后修改为新密码
// @Tags         用户模块
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req body application.ChangePasswordReq true "密码信息"
// @Success      200 {object} response.Response
// @Router       /user/change-password [post]
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
// @Summary      修改个人信息
// @Description  修改昵称、头像、性别、生日等信息
// @Tags         用户模块
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req body application.UpdateProfileReq true "个人信息"
// @Success      200 {object} response.Response
// @Router       /user/profile [put]
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
// @Summary      获取个人信息
// @Description  获取当前登录用户的详细信息
// @Tags         用户模块
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response{data=application.UserInfoResp}
// @Router       /user/profile [get]
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
// @Summary      注销账户
// @Description  永久注销当前登录用户的账号
// @Tags         用户模块
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} response.Response
// @Router       /user/account [delete]
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

// UploadAvatar 上传头像
// @Summary      上传头像
// @Description  上传用户头像，支持 jpg/png/gif/webp，最大 5MB
// @Tags         用户模块
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "头像文件"
// @Success      200 {object} response.Response{data=object{avatar=string}}
// @Router       /user/avatar [post]
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, "请选择要上传的文件")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !h.allowedExts[ext] {
		response.Fail(c, bizErrors.ErrFileTypeNotAllowed, fmt.Sprintf("不支持的文件类型: %s，允许: jpg/jpeg/png/gif/webp", ext))
		return
	}

	if header.Size > h.maxSize {
		response.Fail(c, bizErrors.ErrFileTooLarge, fmt.Sprintf("文件大小超过限制 (%d MB)", h.maxSize/(1024*1024)))
		return
	}

	if err := os.MkdirAll(h.saveDir, 0755); err != nil {
		response.Fail(c, bizErrors.ErrUploadFailed, "创建存储目录失败")
		return
	}

	ts := time.Now().UnixMilli()
	randSuffix := fmt.Sprintf("%04d", rand.Intn(10000))
	filename := fmt.Sprintf("%d_%d_%s%s", userID, ts, randSuffix, ext)
	savePath := filepath.Join(h.saveDir, filename)

	out, err := os.Create(savePath)
	if err != nil {
		response.Fail(c, bizErrors.ErrUploadFailed, "文件保存失败")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(savePath)
		response.Fail(c, bizErrors.ErrUploadFailed, "文件保存失败")
		return
	}

	avatarURL := "/uploads/avatars/" + filename

	if err := h.deleteOldAvatar(c.Request.Context(), userID, avatarURL); err != nil {
	}

	if err := h.svc.UpdateAvatar(c.Request.Context(), userID, avatarURL); err != nil {
		os.Remove(savePath)
		response.Fail(c, bizErrors.ErrUploadFailed, "头像更新失败")
		return
	}

	response.Success(c, gin.H{"avatar": avatarURL})
}

func (h *UserHandler) deleteOldAvatar(ctx context.Context, userID uint, newAvatarURL string) error {
	user, err := h.svc.GetUserInfo(ctx, userID)
	if err != nil {
		return err
	}

	oldURL := user.Avatar
	if oldURL == "" || oldURL == newAvatarURL {
		return nil
	}

	if strings.HasPrefix(oldURL, "/uploads/avatars/") {
		filename := strings.TrimPrefix(oldURL, "/uploads/avatars/")
		oldPath := filepath.Join(h.saveDir, filename)
		os.Remove(oldPath)
	}
	return nil
}

// List 用户列表（管理端）
// @Summary      用户列表
// @Description  管理端分页查询用户列表
// @Tags         管理端-用户
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页条数" default(20)
// @Param        keyword query string false "搜索关键词"
// @Param        status query int false "状态" Enums(0, 1)
// @Success      200 {object} response.PageData{list=[]application.UserItemResp}
// @Router       /admin/users [get]
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

// GetByID 用户详情（管理端）
// @Summary      用户详情
// @Description  管理端获取指定用户详细信息
// @Tags         管理端-用户
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "用户ID"
// @Success      200 {object} response.Response{data=application.UserItemResp}
// @Router       /admin/users/{id} [get]
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

// Create 创建用户（管理端）
// @Summary      创建用户
// @Description  管理端创建新用户
// @Tags         管理端-用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        req body application.AdminCreateUserReq true "用户信息"
// @Success      200 {object} response.Response{data=application.UserItemResp}
// @Router       /admin/users [post]
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

// Update 更新用户（管理端）
// @Summary      更新用户
// @Description  管理端更新用户信息
// @Tags         管理端-用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "用户ID"
// @Param        req body application.AdminUpdateUserReq true "更新信息"
// @Success      200 {object} response.Response{data=application.UserItemResp}
// @Router       /admin/users/{id} [put]
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

// ToggleStatus 启用/禁用用户（管理端）
// @Summary      启用/禁用用户
// @Description  管理端启用或禁用指定用户
// @Tags         管理端-用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "用户ID"
// @Param        req body object{status=int} true "状态: 1启用 0禁用"
// @Success      200 {object} response.Response
// @Router       /admin/users/{id}/status [patch]
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
