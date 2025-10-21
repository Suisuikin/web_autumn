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

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создание нового пользователя в системе
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param user body models.RegisterUserDTO true "Данные для регистрации"
// @Success 201 {object} map[string]interface{} "Пользователь создан"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 409 {object} map[string]string "Пользователь уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /users/register [post]
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

// Login godoc
// @Summary Аутентификация пользователя
// @Description Вход в систему с получением JWT токена
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param credentials body models.LoginDTO true "Учетные данные"
// @Success 200 {object} map[string]interface{} "Успешная аутентификация"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Failure 403 {object} map[string]string "Учетная запись деактивирована"
// @Router /users/login [post]
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

// Logout godoc
// @Summary Выход из системы
// @Description Добавление JWT токена в черный список
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Успешный выход"
// @Failure 400 {object} map[string]string "Токен не предоставлен"
// @Failure 401 {object} map[string]string "Невалидный токен"
// @Failure 500 {object} map[string]string "Ошибка выхода"
// @Router /users/logout [post]
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

// GetProfile godoc
// @Summary Получить профиль текущего пользователя
// @Description Возвращает информацию о текущем авторизованном пользователе
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Профиль пользователя"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /users/profile [get]
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

// UpdateProfile godoc
// @Summary Обновить профиль пользователя
// @Description Обновление данных профиля текущего пользователя
// @Tags Пользователи
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body models.UpdateUserDTO true "Данные для обновления"
// @Success 200 {object} map[string]string "Профиль обновлен"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Пользователь не авторизован"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка обновления"
// @Router /users/profile [put]
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
