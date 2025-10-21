package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"rip/internal/app/models"
	"rip/internal/app/service"
	"rip/internal/pkg/redis"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
)

type AuthMiddleware struct {
	jwtService *service.JWTService
	redis      *redis.Client
}

func NewAuthMiddleware(jwtService *service.JWTService, redis *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		redis:      redis,
	}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := m.validateRequest(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		ctx.Set(UserContextKey, claims)
		ctx.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := m.validateRequest(ctx)
		if claims != nil {
			ctx.Set(UserContextKey, claims)
		}
		ctx.Next()
	}
}

func (m *AuthMiddleware) ModeratorOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := m.validateRequest(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		if !claims.IsModerator {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: moderator access required"})
			ctx.Abort()
			return
		}

		ctx.Set(UserContextKey, claims)
		ctx.Next()
	}
}

func (m *AuthMiddleware) validateRequest(ctx *gin.Context) (*models.JWTClaims, error) {
	authHeader := ctx.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return nil, errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return nil, errors.New("invalid authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

	blacklisted, err := m.redis.IsBlacklisted(ctx.Request.Context(), tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	claims, err := m.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

func GetCurrentUser(ctx *gin.Context) (*models.JWTClaims, error) {
	value, exists := ctx.Get(UserContextKey)
	if !exists {
		return nil, errors.New("user not found in context")
	}

	claims, ok := value.(*models.JWTClaims)
	if !ok {
		return nil, errors.New("invalid user data in context")
	}

	return claims, nil
}

func GetUserID(ctx *gin.Context) (uint, error) {
	claims, err := GetCurrentUser(ctx)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func GetIsModerator(ctx *gin.Context) (bool, error) {
	claims, err := GetCurrentUser(ctx)
	if err != nil {
		return false, err
	}
	return claims.IsModerator, nil
}
