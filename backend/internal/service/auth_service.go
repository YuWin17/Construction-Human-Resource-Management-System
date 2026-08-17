// Package service contains application business logic.
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
	// ErrInvalidCredentials intentionally does not reveal whether an account exists.
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService manages administrator initialization, login, and token validation.
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

// EnsureInitialAdmin creates the first account only when the table is empty.
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

// Login verifies credentials and issues a short-lived signed access token.
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

// CurrentAdmin resolves a signed token to its current administrator record.
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
