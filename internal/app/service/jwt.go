package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"rip/internal/app/models"
	"rip/internal/pkg/config"
)

type JWTService struct {
	cfg config.JWTConfig
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateToken(userID uint, username string, isModerator bool) (string, error) {
	claims := &models.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.cfg.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "chrono-service",
		},
		UserID:      userID,
		Username:    username,
		IsModerator: isModerator,
	}

	token := jwt.NewWithClaims(s.cfg.SigningMethod, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *JWTService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
