package domain

import (
	"context"

	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

// 管理员业务错误
var (
	ErrAdminNotFound      = errors.New(errors.ErrAdminNotFound, "管理员不存在")
	ErrAdminUsernameExist = errors.New(errors.ErrAdminUsernameExist, "管理员用户名已存在")
)

type AdminRepository interface {
	Create(ctx context.Context, admin *Admin) error
	FindByID(ctx context.Context, id uint) (*Admin, error)
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	Update(ctx context.Context, admin *Admin) error
	List(ctx context.Context, page, pageSize int) ([]Admin, int64, error)
}

// OperationLogRepository 操作日志仓库
type OperationLogRepository interface {
	Create(ctx context.Context, log *OperationLog) error
	List(ctx context.Context, page, pageSize int) ([]OperationLog, int64, error)
}
