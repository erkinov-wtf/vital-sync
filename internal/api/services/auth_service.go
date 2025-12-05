package services

import (
	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/constants"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/jwt"
	"gorm.io/gorm"
)

type AuthService struct {
	config *config.Config
	db     *gorm.DB
}

func NewAuthService(config *config.Config, db *gorm.DB) *AuthService {
	return &AuthService{
		config: config,
		db:     db,
	}
}

func (s *AuthService) ValidateAccessToken(accessToken string) (*jwt.CustomClaims, error) {
	claims, err := jwt.ValidateToken(accessToken, s.config.Internal.Jwt.Secret)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != constants.AccessToken {
		return nil, errs.ErrInvalidToken
	}

	return claims, nil
}
