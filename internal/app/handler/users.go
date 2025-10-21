package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"rip/internal/app/middleware"
	"rip/internal/app/models"
	"rip/internal/app/repository"
	"rip/internal/app/service"
	"rip/internal/pkg/redis"
)

type UsersHandler struct {
	Repository *repository.Repository
	JWTService *service.JWTService
	Redis      *redis.Client
}

func NewUsersHandler(r *repository.Repository, jwtService *service.JWTService, redisClient *redis.Client) *UsersHandler {
	return &UsersHandler{
		Repository: r,
		JWTService: jwtService,
		Redis:      redisClient,
	}
}

func (h *UsersHandler) RegisterRoutes(api *gin.RouterGroup, authMW *middleware.AuthMiddleware) {
	users := api.Group("/users")
	{
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)
		users.POST("/logout", authMW.AuthRequired(), h.Logout)
		users.GET("/profile", authMW.AuthRequired(), h.GetProfile)
		users.PUT("/profile", authMW.AuthRequired(), h.UpdateProfile)
	}
}

func (h *UsersHandler) Register(ctx *gin.Context) {
	var input models.RegisterUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	_, err := h.Repository.GetUserByUsername(input.Username)
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким именем уже существует"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хеширования пароля"})
		return
	}

	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		IsModerator:  false,
	}
	if input.Email != nil {
		user.Email = *input.Email
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"message":  "Пользователь успешно зарегистрирован",
	})
}

func (h *UsersHandler) Login(ctx *gin.Context) {
	var input models.LoginDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.Repository.GetUserByUsername(input.Username)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	if !user.IsActive {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Учетная запись деактивирована"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// Генерируем JWT токен
	token, err := h.JWTService.GenerateToken(user.ID, user.Username, user.IsModerator)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"is_moderator": user.IsModerator,
		},
	})
}

func (h *UsersHandler) Logout(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Токен не предоставлен"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := h.JWTService.ValidateToken(token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Невалидный токен"})
		return
	}

	expiresIn := time.Until(time.Unix(claims.ExpiresAt, 0))
	if expiresIn > 0 {
		if err := h.Redis.AddToBlacklist(ctx.Request.Context(), token, expiresIn); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка выхода"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Успешный выход"})
}

func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	claims, err := middleware.GetCurrentUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	user, err := h.Repository.GetUserByID(claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"is_moderator": user.IsModerator,
	})
}

func (h *UsersHandler) UpdateProfile(ctx *gin.Context) {
	claims, err := middleware.GetCurrentUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var input models.UpdateUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.Repository.GetUserByID(claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хеширования"})
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := h.Repository.UpdateUser(claims.UserID, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Профиль обновлен"})
}
