package application

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ZY0506/gin-scaffold/config"
	authDomain "github.com/ZY0506/gin-scaffold/internal/modules/auth/domain"
	blacklistDomain "github.com/ZY0506/gin-scaffold/internal/modules/blacklist/domain"
	userDomain "github.com/ZY0506/gin-scaffold/internal/modules/user/domain"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

const (
	codeTTL    = 300               // 验证码有效期（秒）
	codeChars  = "0123456789"
	codeLength = 6
	roleUser   = "user" // 默认用户角色
)

// JWTService JWT 令牌服务接口（应用层定义，由基础设施层实现）
type JWTService interface {
	// GeneratePair 生成双 Token，返回 accessToken, refreshToken, err
	GeneratePair(userID uint, role string) (accessToken, refreshToken string, err error)
	// ParseToken 解析 Token，返回 userID, role, jti, tokenType, err
	ParseToken(tokenString string) (userID uint, role string, jti string, tokenType string, err error)
}

// TokenBlacklist Token 黑名单接口（应用层定义，由基础设施层实现）
type TokenBlacklist interface {
	// Add 将 JTI 加入黑名单，TTL 为 Token 剩余有效期
	Add(ctx context.Context, jti string, ttl time.Duration) error
}

// AuthService 认证服务，编排所有认证相关的业务用例
type AuthService struct {
	cfg            *config.Config
	userRepo       userDomain.UserRepository
	blRepo         blacklistDomain.BlacklistRepository
	jwtSvc         JWTService
	tokenBlacklist TokenBlacklist
	codeStore      authDomain.CodeStore
	emailSender    authDomain.EmailSender
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	cfg *config.Config,
	userRepo userDomain.UserRepository,
	blRepo blacklistDomain.BlacklistRepository,
	jwtSvc JWTService,
	tokenBlacklist TokenBlacklist,
	codeStore authDomain.CodeStore,
	emailSender authDomain.EmailSender,
) *AuthService {
	return &AuthService{
		cfg:            cfg,
		userRepo:       userRepo,
		blRepo:         blRepo,
		jwtSvc:         jwtSvc,
		tokenBlacklist: tokenBlacklist,
		codeStore:      codeStore,
		emailSender:    emailSender,
	}
}

// SendCode 发送邮箱验证码
func (s *AuthService) SendCode(ctx context.Context, email string) error {
	code := generateCode()

	if err := s.codeStore.Set(ctx, email, code, codeTTL); err != nil {
		return err
	}

	subject := fmt.Sprintf("%s - 验证码", s.cfg.App.Name)
	body := fmt.Sprintf("您的验证码是：%s，有效期 %d 秒。请勿泄露给他人。", code, codeTTL)

	return s.emailSender.Send(ctx, email, subject, body)
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *RegisterReq) error {
	// 1. 验证验证码
	if err := s.verifyCode(ctx, req.Email, req.Code); err != nil {
		return err
	}

	// 2. 检查用户名唯一性
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil && !bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
		return err
	}
	if existing != nil {
		return userDomain.ErrUsernameExist
	}

	// 3. 检查邮箱唯一性
	existing, err = s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
		return err
	}
	if existing != nil {
		return userDomain.ErrEmailExist
	}

	// 4. 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return bizErrors.Wrap(err, bizErrors.ErrPwdHashFailed, "密码加密失败")
	}

	// 5. 创建用户
	user := &userDomain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Username,
		Status:   userDomain.StatusActive,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// 6. 验证成功后清除验证码
	_ = s.codeStore.Del(ctx, req.Email)

	return nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *LoginReq, ip string) (*AuthTokenResp, error) {
	// 1. 查找用户（通过用户名或邮箱）
	user, err := s.userRepo.FindByAccount(ctx, req.Account)
	if err != nil {
		if bizErrors.IsCode(err, bizErrors.ErrUserNotFound) {
			return nil, bizErrors.New(bizErrors.ErrLoginFailed, "用户名或密码错误")
		}
		return nil, err
	}

	// 2. 检查账号状态
	if user.IsDisabled() {
		return nil, userDomain.ErrUserDisabled
	}

	// 3. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, bizErrors.New(bizErrors.ErrLoginFailed, "用户名或密码错误")
	}

	// 4. 检查黑名单（风控）
	bl, err := s.blRepo.CheckLogin(ctx, user.ID, ip)
	if err != nil && !bizErrors.IsCode(err, bizErrors.ErrBlacklisted) {
		return nil, err
	}
	if bl != nil && !bl.IsExpired() {
		return nil, bizErrors.New(bizErrors.ErrBlacklisted, "账号已被限制登录")
	}

	// 5. 生成双 Token（新注册用户使用默认角色）
	role := roleUser
	accessToken, refreshToken, err := s.jwtSvc.GeneratePair(user.ID, role)
	if err != nil {
		return nil, err
	}

	// 6. 更新登录信息（非关键步骤，失败不阻塞登录）
	_ = s.userRepo.UpdateLoginInfo(ctx, user.ID, ip)

	return &AuthTokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWT.AccessExpire.Seconds()),
	}, nil
}

// RefreshToken 刷新双 Token
func (s *AuthService) RefreshToken(ctx context.Context, req *RefreshTokenReq) (*AuthTokenResp, error) {
	// 1. 解析 refresh token
	userID, role, jti, tokenType, err := s.jwtSvc.ParseToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 验证必须是 refresh token
	if tokenType != "refresh" {
		return nil, bizErrors.New(bizErrors.ErrTokenInvalid, "无效的刷新令牌")
	}

	// 3. 将旧的 refresh token 加入黑名单（防重放）
	if err := s.tokenBlacklist.Add(ctx, jti, s.cfg.JWT.RefreshExpire); err != nil {
		return nil, err
	}

	// 4. 生成新的 Token 对
	accessToken, newRefreshToken, err := s.jwtSvc.GeneratePair(userID, role)
	if err != nil {
		return nil, err
	}

	return &AuthTokenResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWT.AccessExpire.Seconds()),
	}, nil
}

// Logout 登出（将当前 Token 加入黑名单）
func (s *AuthService) Logout(ctx context.Context, jti string) error {
	return s.tokenBlacklist.Add(ctx, jti, s.cfg.JWT.AccessExpire)
}

// ChangePassword 修改密码（需验证旧密码）
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordReq) error {
	// 1. 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return bizErrors.New(bizErrors.ErrPwdMismatch, "原密码错误")
	}

	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return bizErrors.Wrap(err, bizErrors.ErrPwdHashFailed, "密码加密失败")
	}

	// 4. 更新密码
	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// ResetPassword 重置密码（通过邮箱验证码）
func (s *AuthService) ResetPassword(ctx context.Context, req *ResetPasswordReq) error {
	// 1. 验证验证码
	if err := s.verifyCode(ctx, req.Email, req.Code); err != nil {
		return err
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	// 3. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return bizErrors.Wrap(err, bizErrors.ErrPwdHashFailed, "密码加密失败")
	}

	// 4. 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 5. 验证成功后清除验证码
	_ = s.codeStore.Del(ctx, req.Email)

	return nil
}

// UpdateProfile 修改个人信息
func (s *AuthService) UpdateProfile(ctx context.Context, userID uint, req *UpdateProfileReq) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
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
	if req.Birthday != "" {
		t, parseErr := time.Parse("2006-01-02", req.Birthday)
		if parseErr != nil {
			return bizErrors.New(bizErrors.ErrBadRequest, "生日格式错误，正确格式: 2006-01-02")
		}
		user.Birthday = &t
	}

	return s.userRepo.Update(ctx, user)
}

// GetUserInfo 获取当前登录用户信息
func (s *AuthService) GetUserInfo(ctx context.Context, userID uint) (*UserInfoResp, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &UserInfoResp{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if user.Birthday != nil {
		resp.Birthday = user.Birthday.Format("2006-01-02")
	}
	if user.LastLoginAt != nil {
		resp.LastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}
	resp.LastLoginIP = user.LastLoginIP

	return resp, nil
}

// DeleteAccount 注销账户
func (s *AuthService) DeleteAccount(ctx context.Context, userID uint) error {
	return s.userRepo.Delete(ctx, userID)
}

// verifyCode 验证邮箱验证码
func (s *AuthService) verifyCode(ctx context.Context, email, code string) error {
	storedCode, err := s.codeStore.Get(ctx, email)
	if err != nil {
		return err
	}
	if storedCode == "" {
		return bizErrors.New(bizErrors.ErrCodeExpired, "验证码已过期或不存在")
	}
	if storedCode != code {
		return bizErrors.New(bizErrors.ErrCodeInvalid, "验证码错误")
	}
	return nil
}

// generateCode 生成 N 位数字验证码
func generateCode() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, codeLength)
	for i := range code {
		code[i] = codeChars[rng.Intn(len(codeChars))]
	}
	return string(code)
}
