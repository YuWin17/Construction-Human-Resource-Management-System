// Package service 包含应用业务逻辑。
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials 不暴露账号是否存在，避免泄露认证信息。
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService 管理管理员初始化、登录和令牌校验。
type AuthService struct {
	admins    *repository.AdminRepository
	jwtSecret []byte
	jwtTTL    time.Duration
}

func NewAuthService(admins *repository.AdminRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		admins:    admins,
		jwtSecret: []byte(jwtSecret),
		jwtTTL:    jwtTTL,
	}
}

// EnsureInitialAdmin 仅在管理员表为空时创建首个账号。
func (s *AuthService) EnsureInitialAdmin(ctx context.Context, username, password string) error {
	hasAny, err := s.admins.HasAny(ctx)
	if err != nil {
		return fmt.Errorf("count administrators: %w", err)
	}
	if hasAny {
		return nil
	}

	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return errors.New("INITIAL_ADMIN_USERNAME and INITIAL_ADMIN_PASSWORD are required for an empty database")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash initial password: %w", err)
	}

	admin := &domain.Admin{Username: username, PasswordHash: string(hash)}
	if err := s.admins.Create(ctx, admin); err != nil {
		return fmt.Errorf("create initial administrator: %w", err)
	}
	return nil
}

// Login 校验凭据并签发短期访问令牌。
func (s *AuthService) Login(ctx context.Context, username, password string) (string, domain.Admin, error) {
	admin, err := s.admins.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return "", domain.Admin{}, ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   admin.ID,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", domain.Admin{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, admin, nil
}

// CurrentAdmin 将已签名令牌解析为当前管理员记录。
func (s *AuthService) CurrentAdmin(ctx context.Context, rawToken string) (domain.Admin, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return domain.Admin{}, ErrInvalidToken
	}

	admin, err := s.admins.FindByID(ctx, claims.Subject)
	if err != nil {
		return domain.Admin{}, ErrInvalidToken
	}
	return admin, nil
}
