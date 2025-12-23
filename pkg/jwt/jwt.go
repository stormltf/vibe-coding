package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenNotValidYet    = errors.New("token not active yet")
	ErrTokenMalformed      = errors.New("token is malformed")
	ErrTokenInvalid        = errors.New("token is invalid")
	ErrSecretNotConfigured = errors.New("jwt secret not configured")
	ErrRefreshTooEarly     = errors.New("token refresh not allowed: token still has sufficient validity")
)

// Config JWT 配置
type Config struct {
	Secret     string        // 密钥
	Issuer     string        // 签发者
	ExpireTime time.Duration // 过期时间
}

// DefaultConfig 默认配置
// 警告：仅用于开发环境，生产环境必须通过配置文件或环境变量设置安全的密钥
func DefaultConfig() *Config {
	return &Config{
		Secret:     "", // 空密钥，强制用户显式配置
		Issuer:     "test-tt",
		ExpireTime: 24 * time.Hour,
	}
}

// MinRefreshWindow Token 刷新的最小剩余有效期（只有剩余时间小于此值才允许刷新）
const MinRefreshWindow = 2 * time.Hour

// Claims 自定义声明
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT JWT 工具类
type JWT struct {
	config *Config
}

// New 创建 JWT 实例
func New(config *Config) *JWT {
	if config == nil {
		config = DefaultConfig()
	}
	return &JWT{config: config}
}

// ValidateConfig 验证 JWT 配置
func ValidateConfig(config *Config) error {
	if config == nil || config.Secret == "" {
		return ErrSecretNotConfigured
	}
	if len(config.Secret) < 32 {
		return errors.New("jwt secret must be at least 32 characters")
	}
	return nil
}

// GenerateToken 生成 token
func (j *JWT) GenerateToken(userID uint64, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新 token
// 只有当 token 剩余有效期小于 MinRefreshWindow 时才允许刷新
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查 token 剩余有效期
	if claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > MinRefreshWindow {
			return "", ErrRefreshTooEarly
		}
	}

	return j.GenerateToken(claims.UserID, claims.Username)
}

// ForceRefreshToken 强制刷新 token（不检查剩余有效期）
// 仅在特殊场景使用，如用户权限变更后需要立即刷新
func (j *JWT) ForceRefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return j.GenerateToken(claims.UserID, claims.Username)
}
