package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"rip/internal/app/models"
	"rip/internal/app/repository"
)

type UsersHandler struct {
	Repository *repository.Repository
}

func NewUsersHandler(r *repository.Repository) *UsersHandler {
	return &UsersHandler{Repository: r}
}

func (h *UsersHandler) RegisterRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("/register", h.Register)    // 17
		users.POST("/login", h.Login)          // 18
		users.POST("/logout", h.Logout)        // 19
		users.GET("/profile", h.GetProfile)    // 20
		users.PUT("/profile", h.UpdateProfile) // 21
	}
}

// 17. POST /api/users/register
func (h *UsersHandler) Register(ctx *gin.Context) {
	var input models.RegisterUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
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
	})
}

// 18. POST /api/users/login
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// TODO: Создать сессию или JWT токен

	ctx.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"is_moderator": user.IsModerator,
	})
}

// 19. POST /api/users/logout
func (h *UsersHandler) Logout(ctx *gin.Context) {
	// TODO: Удалить сессию или инвалидировать JWT

	ctx.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// 20. GET /api/users/profile
func (h *UsersHandler) GetProfile(ctx *gin.Context) {
	// TODO: Получить userID из сессии/JWT
	userID := uint(1)

	user, err := h.Repository.GetUserByID(userID)
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

// 21. PUT /api/users/profile
func (h *UsersHandler) UpdateProfile(ctx *gin.Context) {
	// TODO: Получить userID из сессии/JWT
	userID := uint(1)

	var input models.UpdateUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.Repository.GetUserByID(userID)
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

	if err := h.Repository.UpdateUser(userID, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}
