package domain

import (
	"context"

	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

// 用户模块业务错误
var (
	ErrUserNotFound  = errors.New(errors.ErrUserNotFound, "用户不存在")
	ErrUsernameExist = errors.New(errors.ErrUsernameExist, "用户名已存在")
	ErrEmailExist    = errors.New(errors.ErrEmailExist, "邮箱已被注册")
	ErrUserDisabled  = errors.New(errors.ErrUserDisabled, "账号已被禁用")
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByAccount(ctx context.Context, account string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	// List 分页查询，conditions 为 GORM Where 条件的 key-value 映射
	List(ctx context.Context, page, pageSize int, conditions map[string]any) ([]User, int64, error)
	UpdateLoginInfo(ctx context.Context, id uint, ip string) error
}
