package services

import (
	"errors"

	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/constants"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/jwt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrUserExists          = errors.New("user with this email or username already exists")
	ErrSessionInvalid      = errors.New("session is invalid or expired")
	ErrInvalidToken        = errors.New("invalid token format")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
)

type AuthService struct {
	config *config.Config
}

func NewAuthService(config *config.Config) *AuthService {
	return &AuthService{
		config: config,
	}
}

func (s *AuthService) ValidateAccessToken(accessToken string) (*jwt.CustomClaims, error) {
	claims, err := jwt.ValidateToken(accessToken, s.config.Internal.Jwt.Secret)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != constants.AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
