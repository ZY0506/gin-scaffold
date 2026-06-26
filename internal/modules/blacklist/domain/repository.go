package domain

import (
	"context"

	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

// 黑名单模块业务错误
var (
	ErrBlacklistNotFound  = errors.New(errors.ErrBlacklisted, "黑名单记录不存在")
	ErrAlreadyBlacklisted = errors.New(errors.ErrBlacklisted, "该用户或IP已被加入黑名单")
)

type BlacklistRepository interface {
	Create(ctx context.Context, bl *Blacklist) error
	FindByID(ctx context.Context, id uint) (*Blacklist, error)
	// Update 更新黑名单记录（原因、过期时间等）
	Update(ctx context.Context, bl *Blacklist) error
	// FindActiveByUserID 查询用户是否有生效的黑名单记录
	FindActiveByUserID(ctx context.Context, userID uint) (*Blacklist, error)
	// FindActiveByIP 查询IP是否有生效的黑名单记录
	FindActiveByIP(ctx context.Context, ip string) (*Blacklist, error)
	// CheckLogin 检查登录用户和IP是否被封禁，返回最先命中的记录
	CheckLogin(ctx context.Context, userID uint, ip string) (*Blacklist, error)
	// List 分页查询黑名单记录
	List(ctx context.Context, page, pageSize int) ([]Blacklist, int64, error)
	// Deactivate 解封（软解除）
	Deactivate(ctx context.Context, id uint) error
}
