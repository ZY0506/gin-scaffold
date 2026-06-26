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

// UploadAvatar 上传头像
// POST /api/v1/user/avatar (multipart/form-data, field: "file")
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, bizErrors.ErrUnauthorized, "未获取到用户信息")
		return
	}

	// 1. 获取上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, bizErrors.ErrBadRequest, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 2. 校验文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !h.allowedExts[ext] {
		response.Fail(c, bizErrors.ErrFileTypeNotAllowed, fmt.Sprintf("不支持的文件类型: %s，允许: jpg/jpeg/png/gif/webp", ext))
		return
	}

	// 3. 校验文件大小
	if header.Size > h.maxSize {
		response.Fail(c, bizErrors.ErrFileTooLarge, fmt.Sprintf("文件大小超过限制 (%d MB)", h.maxSize/(1024*1024)))
		return
	}

	// 4. 确保存储目录存在
	if err := os.MkdirAll(h.saveDir, 0755); err != nil {
		response.Fail(c, bizErrors.ErrUploadFailed, "创建存储目录失败")
		return
	}

	// 5. 生成唯一文件名: {userID}_{时间戳}_{4位随机}.{ext}
	ts := time.Now().UnixMilli()
	randSuffix := fmt.Sprintf("%04d", rand.Intn(10000))
	filename := fmt.Sprintf("%d_%d_%s%s", userID, ts, randSuffix, ext)
	savePath := filepath.Join(h.saveDir, filename)

	// 6. 保存文件
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

	// 7. 生成头像访问 URL
	avatarURL := "/uploads/avatars/" + filename

	// 8. 删除旧头像文件
	if err := h.deleteOldAvatar(c.Request.Context(), userID, avatarURL); err != nil {
		// 删除旧文件失败不影响主流程，只记录日志
	}

	// 9. 更新数据库中的头像字段
	if err := h.svc.UpdateAvatar(c.Request.Context(), userID, avatarURL); err != nil {
		// 数据库更新失败，删除已保存的文件
		os.Remove(savePath)
		response.Fail(c, bizErrors.ErrUploadFailed, "头像更新失败")
		return
	}

	response.Success(c, gin.H{"avatar": avatarURL})
}

// deleteOldAvatar 删除用户旧的头像文件
func (h *UserHandler) deleteOldAvatar(ctx context.Context, userID uint, newAvatarURL string) error {
	user, err := h.svc.GetUserInfo(ctx, userID)
	if err != nil {
		return err
	}

	oldURL := user.Avatar
	if oldURL == "" || oldURL == newAvatarURL {
		return nil
	}

	// 只清理本地上传的头像文件
	if strings.HasPrefix(oldURL, "/uploads/avatars/") {
		filename := strings.TrimPrefix(oldURL, "/uploads/avatars/")
		oldPath := filepath.Join(h.saveDir, filename)
		os.Remove(oldPath) // 删除失败不影响主流程
	}
	return nil
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
