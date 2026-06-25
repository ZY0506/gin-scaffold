package application

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
	userDomain "github.com/ZY0506/gin-scaffold/internal/modules/user/domain"
)

type UserService struct {
	userRepo userDomain.UserRepository
}

func NewUserService(userRepo userDomain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateByAdmin 管理员创建用户
func (s *UserService) CreateByAdmin(ctx context.Context, req *AdminCreateUserReq) (*UserItemResp, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, userDomain.ErrUsernameExist
	}
	if err != nil && !errors.IsCode(err, errors.ErrUserNotFound) {
		return nil, err
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingEmail != nil {
		return nil, userDomain.ErrEmailExist
	}
	if err != nil && !errors.IsCode(err, errors.ErrUserNotFound) {
		return nil, err
	}

	// 密码加密
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrPwdHashFailed, "密码加密失败")
	}

	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	user := &userDomain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPwd),
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Status:   status,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Wrap(err, errors.ErrDB, "创建用户失败")
	}

	return toUserItemResp(user), nil
}

// UpdateByAdmin 管理员更新用户信息
func (s *UserService) UpdateByAdmin(ctx context.Context, id uint, req *AdminUpdateUserReq) (*UserItemResp, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.Wrap(err, errors.ErrDB, "更新用户失败")
	}

	return toUserItemResp(user), nil
}

// GetByID 管理员获取用户详情
func (s *UserService) GetByID(ctx context.Context, id uint) (*UserItemResp, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserItemResp(user), nil
}

// List 管理员分页查询用户列表
func (s *UserService) List(ctx context.Context, req *UserListReq) ([]UserItemResp, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	conditions := make(map[string][]interface{})
	if req.Status != nil {
		conditions["status"] = []interface{}{*req.Status}
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		conditions["(username LIKE ? OR email LIKE ? OR nickname LIKE ?)"] =
			[]interface{}{kw, kw, kw}
	}

	users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize, conditions)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrDB, "查询用户列表失败")
	}

	resp := make([]UserItemResp, len(users))
	for i, u := range users {
		resp[i] = *toUserItemResp(&u)
	}

	return resp, total, nil
}

// ToggleStatus 启用/禁用用户
func (s *UserService) ToggleStatus(ctx context.Context, id uint, status int) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	user.Status = status
	return s.userRepo.Update(ctx, user)
}

func toUserItemResp(u *userDomain.User) *UserItemResp {
	resp := &UserItemResp{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Nickname:    u.Nickname,
		Avatar:      u.Avatar,
		Gender:      u.Gender,
		Status:      u.Status,
		LastLoginIP: u.LastLoginIP,
		CreatedAt:   u.CreatedAt.Format(time.DateTime),
	}
	if u.LastLoginAt != nil {
		resp.LastLoginAt = u.LastLoginAt.Format(time.DateTime)
	}
	return resp
}
