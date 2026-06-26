package application

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	adminDomain "github.com/ZY0506/gin-scaffold/internal/modules/admin/domain"
	"github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

type AdminService struct {
	repo adminDomain.AdminRepository
}

func NewAdminService(repo adminDomain.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// Login 管理员登录，返回 adminID, username, err
func (s *AdminService) Login(ctx context.Context, req *AdminLoginReq, ip string) (uint, string, error) {
	admin, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.IsCode(err, errors.ErrAdminNotFound) {
			return 0, "", errors.New(errors.ErrLoginFailed, "用户名或密码错误")
		}
		return 0, "", err
	}

	if admin.IsDisabled() {
		return 0, "", errors.New(errors.ErrUserDisabled, "管理员账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return 0, "", errors.New(errors.ErrLoginFailed, "用户名或密码错误")
	}

	return admin.ID, admin.Username, nil
}

// Create 创建管理员
func (s *AdminService) Create(ctx context.Context, req *AdminCreateReq) (*AdminItemResp, error) {
	existing, err := s.repo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, adminDomain.ErrAdminUsernameExist
	}
	if err != nil && !errors.IsCode(err, errors.ErrAdminNotFound) {
		return nil, err
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrPwdHashFailed, "密码加密失败")
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	admin := &adminDomain.Admin{
		Username: req.Username,
		Password: string(hashedPwd),
		Nickname: nickname,
		Status:   1,
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, errors.Wrap(err, errors.ErrDB, "创建管理员失败")
	}

	return toItemResp(admin), nil
}

// List 分页查询管理员列表
func (s *AdminService) List(ctx context.Context, page, pageSize int) ([]AdminItemResp, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrDB, "查询管理员列表失败")
	}

	resp := make([]AdminItemResp, len(list))
	for i, a := range list {
		resp[i] = *toItemResp(&a)
	}
	return resp, total, nil
}

// GetByID 获取管理员详情
func (s *AdminService) GetByID(ctx context.Context, id uint) (*AdminItemResp, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toItemResp(admin), nil
}

// Update 修改管理员
func (s *AdminService) Update(ctx context.Context, id uint, req *AdminUpdateReq) error {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Nickname != "" {
		admin.Nickname = req.Nickname
	}
	if req.Password != "" {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, errors.ErrPwdHashFailed, "密码加密失败")
		}
		admin.Password = string(hashedPwd)
	}
	if req.Status != nil {
		admin.Status = *req.Status
	}

	return s.repo.Update(ctx, admin)
}

func toItemResp(a *adminDomain.Admin) *AdminItemResp {
	resp := &AdminItemResp{
		ID:        a.ID,
		Username:  a.Username,
		Nickname:  a.Nickname,
		Avatar:    a.Avatar,
		Status:    a.Status,
		CreatedAt: a.CreatedAt.Format(time.DateTime),
	}
	if a.LastLoginAt != nil {
		resp.LastLoginAt = a.LastLoginAt.Format(time.DateTime)
	}
	return resp
}
