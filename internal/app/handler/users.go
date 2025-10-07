package handler

import (
	"net/http"
	"rip/internal/app/models"
	"rip/internal/app/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
		users.POST("", h.Register)
		users.GET("/me", h.GetMyUserData)
		users.PUT("/me", h.UpdateMyUserData)
	}
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
	}
}

func (h *UsersHandler) Register(ctx *gin.Context) {
	var input models.RegisterUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось зарегистрировать пользователя"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username, "email": user.Email})
}

func (h *UsersHandler) GetMyUserData(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (h *UsersHandler) UpdateMyUserData(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var input models.UpdateUserDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err := h.Repository.UpdateUser(userID, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить пользователя"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "обновлено"})
}

func (h *UsersHandler) Login(ctx *gin.Context) {
	var input models.LoginDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.Repository.GetUserByUsername(input.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user_id": user.ID})
}

func (h *UsersHandler) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "вышли"})
}
