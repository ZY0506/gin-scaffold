package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domain "github.com/ZY0506/gin-scaffold/internal/modules/admin/domain"
)

type GormAdminRepo struct {
	db *gorm.DB
}

func NewGormAdminRepo(db *gorm.DB) *GormAdminRepo {
	return &GormAdminRepo{db: db}
}

func (r *GormAdminRepo) Create(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *GormAdminRepo) FindByID(ctx context.Context, id uint) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.WithContext(ctx).First(&admin, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminRepo) FindByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminRepo) Update(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Model(&domain.Admin{}).Where("id = ?", admin.ID).Updates(admin).Error
}

func (r *GormAdminRepo) List(ctx context.Context, page, pageSize int) ([]domain.Admin, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var list []domain.Admin
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Admin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
