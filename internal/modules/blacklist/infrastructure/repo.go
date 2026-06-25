package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	domain "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/domain"
)

type GormBlacklistRepo struct {
	db *gorm.DB
}

func NewGormBlacklistRepo(db *gorm.DB) *GormBlacklistRepo {
	return &GormBlacklistRepo{db: db}
}

func (r *GormBlacklistRepo) Create(ctx context.Context, bl *domain.Blacklist) error {
	return r.db.WithContext(ctx).Create(bl).Error
}

// activeScope 查询生效中的黑名单：IsActive=true 且（未过期或永久）
func (r *GormBlacklistRepo) activeScope(ctx context.Context) *gorm.DB {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Blacklist{}).
		Where("is_active = ?", true).
		Where("expired_at IS NULL OR expired_at > ?", now)
}

func (r *GormBlacklistRepo) FindActiveByUserID(ctx context.Context, userID uint) (*domain.Blacklist, error) {
	var bl domain.Blacklist
	err := r.activeScope(ctx).Where("user_id = ?", userID).First(&bl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bl, nil
}

func (r *GormBlacklistRepo) FindActiveByIP(ctx context.Context, ip string) (*domain.Blacklist, error) {
	var bl domain.Blacklist
	err := r.activeScope(ctx).Where("ip = ?", ip).First(&bl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bl, nil
}

func (r *GormBlacklistRepo) CheckLogin(ctx context.Context, userID uint, ip string) (*domain.Blacklist, error) {
	var bl domain.Blacklist
	err := r.activeScope(ctx).
		Where("user_id = ? OR ip = ?", userID, ip).
		Order("created_at DESC").
		First(&bl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bl, nil
}

func (r *GormBlacklistRepo) List(ctx context.Context, page, pageSize int) ([]domain.Blacklist, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var list []domain.Blacklist
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Blacklist{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *GormBlacklistRepo) Deactivate(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&domain.Blacklist{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrBlacklistNotFound
	}
	return nil
}
