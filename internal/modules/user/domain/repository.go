package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByAccount(ctx context.Context, account string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]User, int64, error)
	UpdateLoginInfo(ctx context.Context, id uint, ip string) error
}
