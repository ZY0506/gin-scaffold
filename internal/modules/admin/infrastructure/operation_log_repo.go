package infrastructure

import (
	"context"

	"gorm.io/gorm"

	domain "github.com/ZY0506/gin-scaffold/internal/modules/admin/domain"
)

type GormOperationLogRepo struct {
	db *gorm.DB
}

func NewGormOperationLogRepo(db *gorm.DB) *GormOperationLogRepo {
	return &GormOperationLogRepo{db: db}
}

func (r *GormOperationLogRepo) Create(ctx context.Context, log *domain.OperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormOperationLogRepo) List(ctx context.Context, page, pageSize int) ([]domain.OperationLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var list []domain.OperationLog
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.OperationLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
