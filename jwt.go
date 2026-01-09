package goauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWT 错误
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidClaims    = errors.New("invalid claims")
	ErrTokenRevoked     = errors.New("token revoked")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// TokenType Token 类型
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// JWTConfig JWT 配置
type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key" json:"secret_key"`         // JWT 密钥
	Issuer        string        `yaml:"issuer" json:"issuer"`                 // 签发者
	AccessExpiry  time.Duration `yaml:"access_expiry" json:"access_expiry"`   // Access Token 有效期
	RefreshExpiry time.Duration `yaml:"refresh_expiry" json:"refresh_expiry"` // Refresh Token 有效期
}

// DefaultJWTConfig 默认 JWT 配置
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:     "",
		Issuer:        "go-auth",
		AccessExpiry:  24 * time.Hour,     // 24小时
		RefreshExpiry: 7 * 24 * time.Hour, // 7天
	}
}

// Claims JWT Claims
type Claims struct {
	UserID   uint64    `json:"user_id"`            // 用户ID
	Username string    `json:"username"`           // 用户名
	Role     string    `json:"role,omitempty"`     // 角色（如 admin, user）
	AppID    string    `json:"app_id,omitempty"`   // 应用ID
	Type     TokenType `json:"type"`               // Token 类型
	Custom   map[string]interface{} `json:"custom,omitempty"` // 自定义字段
	jwt.RegisteredClaims
}

// JWTService JWT 服务
type JWTService struct {
	config     JWTConfig
	blacklist  TokenBlacklist
}

// NewJWTService 创建 JWT 服务
func NewJWTService(config JWTConfig) *JWTService {
	if config.SecretKey == "" {
		panic("JWT secret key is required")
	}
	if config.Issuer == "" {
		config.Issuer = "go-auth"
	}
	if config.AccessExpiry == 0 {
		config.AccessExpiry = 24 * time.Hour
	}
	if config.RefreshExpiry == 0 {
		config.RefreshExpiry = 7 * 24 * time.Hour
	}

	return &JWTService{
		config:    config,
		blacklist: NewInMemoryBlacklist(),
	}
}

// WithBlacklist 设置 Token 黑名单
func (s *JWTService) WithBlacklist(blacklist TokenBlacklist) *JWTService {
	s.blacklist = blacklist
	return s
}

// GenerateAccessToken 生成 Access Token
func (s *JWTService) GenerateAccessToken(userID uint64, username, role, appID string, custom map[string]interface{}) (string, error) {
	return s.generateToken(userID, username, role, appID, TokenTypeAccess, s.config.AccessExpiry, custom)
}

// GenerateRefreshToken 生成 Refresh Token
func (s *JWTService) GenerateRefreshToken(userID uint64, username string) (string, error) {
	return s.generateToken(userID, username, "", "", TokenTypeRefresh, s.config.RefreshExpiry, nil)
}

// generateToken 生成 Token
func (s *JWTService) generateToken(userID uint64, username, role, appID string, tokenType TokenType, expiry time.Duration, custom map[string]interface{}) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		AppID:    appID,
		Type:     tokenType,
		Custom:   custom,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// ParseToken 解析并验证 Token
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	// 检查黑名单
	if s.blacklist != nil && s.blacklist.IsBlacklisted(HashToken(tokenString)) {
		return nil, ErrTokenRevoked
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshAccessToken 刷新 Access Token
func (s *JWTService) RefreshAccessToken(refreshToken string, role, appID string, custom map[string]interface{}) (string, error) {
	// 解析 Refresh Token
	claims, err := s.ParseToken(refreshToken)
	if err != nil {
		return "", err
	}

	// 验证是 Refresh Token
	if claims.Type != TokenTypeRefresh {
		return "", ErrInvalidTokenType
	}

	// 生成新的 Access Token
	return s.GenerateAccessToken(claims.UserID, claims.Username, role, appID, custom)
}

// RevokeToken 撤销 Token（加入黑名单）
func (s *JWTService) RevokeToken(tokenString string) error {
	if s.blacklist == nil {
		return errors.New("token blacklist not configured")
	}

	// 解析 Token 获取过期时间
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		// 即使 Token 无效也加入黑名单
		claims = &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpiry)),
			},
		}
	}

	return s.blacklist.Add(HashToken(tokenString), claims.ExpiresAt.Time)
}

// ValidateAccessToken 验证 Access Token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeAccess {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken 验证 Refresh Token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// TokenPair Token 对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresIn    int64     `json:"expires_in"` // 秒
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair 生成 Token 对
func (s *JWTService) GenerateTokenPair(userID uint64, username, role, appID string, custom map[string]interface{}) (*TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(userID, username, role, appID, custom)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(userID, username)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.config.AccessExpiry)
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ExpiresIn:    int64(s.config.AccessExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GetConfig 获取配置
func (s *JWTService) GetConfig() JWTConfig {
	return s.config
}
