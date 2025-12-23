package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/test-tt/config"
	"github.com/test-tt/internal/dao"
	"github.com/test-tt/internal/model"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/jwt"
)

const (
	tokenBlacklistKey = "token:blacklist:%s"
	minPasswordLength = 6
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrEmailExists      = errors.New("email already exists")
	ErrPasswordTooShort = errors.New("password too short")
	ErrTokenBlacklisted = errors.New("token is blacklisted")
)

type AuthService struct {
	userDAO *dao.UserDAO
	jwt     *jwt.JWT
}

func NewAuthService() *AuthService {
	var jwtConfig *jwt.Config
	if config.Cfg != nil && config.Cfg.JWT != nil {
		jwtConfig = &jwt.Config{
			Secret:     config.Cfg.JWT.Secret,
			Issuer:     config.Cfg.JWT.Issuer,
			ExpireTime: config.Cfg.JWT.ExpireTime,
		}
	} else {
		jwtConfig = jwt.DefaultConfig()
		jwtConfig.Secret = "dev-secret-key-at-least-32-chars!"
	}
	return &AuthService{
		userDAO: dao.NewUserDAO(),
		jwt:     jwt.New(jwtConfig),
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, string, error) {
	if len(password) < minPasswordLength {
		return nil, "", ErrPasswordTooShort
	}

	// Check if email already exists
	exists, err := s.userDAO.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userDAO.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := s.jwt.GenerateToken(user.ID, user.Name)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login authenticates user and returns token
func (s *AuthService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	user, err := s.userDAO.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	// Generate token
	token, err := s.jwt.GenerateToken(user.ID, user.Name)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Logout invalidates a token by adding it to blacklist
func (s *AuthService) Logout(ctx context.Context, token string) error {
	claims, err := s.jwt.ParseToken(token)
	if err != nil {
		return err
	}

	if cache.RDB == nil {
		return nil
	}

	// Calculate remaining TTL
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 {
		return nil
	}

	key := fmt.Sprintf(tokenBlacklistKey, token)
	return cache.Set(ctx, key, "1", remaining)
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) bool {
	if cache.RDB == nil {
		return false
	}
	key := fmt.Sprintf(tokenBlacklistKey, token)
	val, err := cache.Get(ctx, key)
	return err == nil && val != ""
}

// UpdateProfile updates user profile information
func (s *AuthService) UpdateProfile(ctx context.Context, userID uint64, name string, age int, email string) (*model.User, error) {
	user, err := s.userDAO.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if new email is already used by another user
	if email != "" && email != user.Email {
		existingUser, err := s.userDAO.GetByEmail(ctx, email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, ErrEmailExists
		}
	}

	// Update fields
	fields := make(map[string]interface{})
	if name != "" {
		fields["name"] = name
	}
	if age > 0 {
		fields["age"] = age
	}
	if email != "" {
		fields["email"] = email
	}

	if len(fields) > 0 {
		if err := s.userDAO.UpdateFields(ctx, userID, fields); err != nil {
			return nil, err
		}
	}

	// Fetch updated user
	return s.userDAO.GetByID(ctx, userID)
}

// ChangePassword changes user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	if len(newPassword) < minPasswordLength {
		return ErrPasswordTooShort
	}

	user, err := s.userDAO.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userDAO.UpdateFields(ctx, userID, map[string]interface{}{
		"password": string(hashedPassword),
	})
}

// DeleteAccount deletes user account
func (s *AuthService) DeleteAccount(ctx context.Context, userID uint64, password string) error {
	user, err := s.userDAO.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify password before deletion
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return ErrInvalidPassword
	}

	return s.userDAO.Delete(ctx, userID)
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*model.User, error) {
	user, err := s.userDAO.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
