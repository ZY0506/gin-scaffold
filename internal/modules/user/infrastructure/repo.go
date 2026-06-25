package infrastructure

import (
	"context"
	"errors"

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

func (r *GormUserRepo) findBy(ctx context.Context, field, value string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where(field+" = ?", value).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.findBy(ctx, "username", username)
}

func (r *GormUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findBy(ctx, "email", email)
}

func (r *GormUserRepo) FindByAccount(ctx context.Context, account string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ? OR email = ?", account, account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepo) Update(ctx context.Context, user *domain.User) error {
	// 使用 Updates 而非 Save，避免零值字段覆盖数据库已有值
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", user.ID).Updates(user).Error
}

func (r *GormUserRepo) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *GormUserRepo) List(ctx context.Context, page, pageSize int, conditions map[string][]interface{}) ([]domain.User, int64, error) {
	// 参数边界校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var users []domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{})
	for key, args := range conditions {
		query = query.Where(key, args...)
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
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
	}).Error
}
