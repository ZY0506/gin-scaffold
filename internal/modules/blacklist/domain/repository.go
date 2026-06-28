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

// BlacklistRepository 黑名单仓库接口
type BlacklistRepository interface {
	Create(ctx context.Context, bl *Blacklist) error
	FindByID(ctx context.Context, id uint) (*Blacklist, error)
	Update(ctx context.Context, bl *Blacklist) error
	FindActiveByUserID(ctx context.Context, userID uint) (*Blacklist, error)
	FindActiveByIP(ctx context.Context, ip string) (*Blacklist, error)
	CheckLogin(ctx context.Context, userID uint, ip string) (*Blacklist, error)
	List(ctx context.Context, page, pageSize int) ([]Blacklist, int64, error)
	Deactivate(ctx context.Context, id uint) error
}
