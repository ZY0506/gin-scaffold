package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	domain "github.com/ZY0506/gin-scaffold/internal/modules/user/domain"
)

type GormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db: db}
}

func (r *GormUserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByAccount(ctx context.Context, account string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ? OR email = ?", account, account).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r *GormUserRepo) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{})
	for key, val := range conditions {
		query = query.Where(key, val)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *GormUserRepo) UpdateLoginInfo(ctx context.Context, id uint, ip string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
	}).Error
}
