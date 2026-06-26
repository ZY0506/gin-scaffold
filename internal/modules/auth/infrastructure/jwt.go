package infrastructure

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ZY0506/gin-scaffold/config"
	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

type JWTService struct {
	cfg *config.JWTConfig
}

func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

// GeneratePair 生成双 Token，返回 accessToken, refreshToken, err
func (s *JWTService) GeneratePair(userID uint, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.generateToken(userID, role, "access", s.cfg.AccessExpire)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.generateToken(userID, role, "refresh", s.cfg.RefreshExpire)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken 验证 Token 并解析 Claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, bizErrors.New(bizErrors.ErrTokenInvalid, "无效的签名方法")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		if isExpiredJWTError(err) {
			return nil, bizErrors.New(bizErrors.ErrTokenExpired, "令牌已过期")
		}
		return nil, bizErrors.New(bizErrors.ErrTokenInvalid, "无效的令牌")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, bizErrors.New(bizErrors.ErrTokenInvalid, "无效的令牌")
	}

	return claims, nil
}

// GetJTI 从 token 中提取 JTI（JWT ID）
func (s *JWTService) GetJTI(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

// ParseToken 解析 Token，返回 userID, role, jti, tokenType, err
func (s *JWTService) ParseToken(tokenString string) (userID uint, role string, jti string, tokenType string, err error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, "", "", "", err
	}
	return claims.UserID, claims.Role, claims.ID, claims.TokenType, nil
}

// generateToken 生成单个 Token
func (s *JWTService) generateToken(userID uint, role, tokenType string, expire time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", bizErrors.Wrap(err, bizErrors.ErrTokenInvalid, "令牌签名失败")
	}

	return tokenString, nil
}

// isExpiredJWTError 判断 JWT 错误是否为 Token 过期
func isExpiredJWTError(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet)
}
