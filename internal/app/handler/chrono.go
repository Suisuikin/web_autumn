package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rip/internal/app/middleware"
	"rip/internal/app/models"
	"rip/internal/app/repository"
)

type RequestsHandler struct {
	Repository *repository.Repository
}

func NewRequestsHandler(r *repository.Repository) *RequestsHandler {
	return &RequestsHandler{Repository: r}
}

func (h *RequestsHandler) RegisterRoutes(api *gin.RouterGroup) {
	requests := api.Group("/chrono")
	{
		requests.GET("/cart-icon", h.GetCartIcon)        // 8
		requests.GET("", h.GetRequests)                  // 9
		requests.GET("/:id", h.GetRequestByID)           // 10
		requests.PUT("/:id", h.UpdateRequest)            // 11
		requests.PUT("/:id/form", h.FormRequest)         // 12
		requests.PUT("/:id/complete", h.CompleteRequest) // 13
		requests.DELETE("/:id", h.DeleteRequest)         // 14
	}
}

// 8. GET /api/chrono/cart-icon - иконка корзины
func (h *RequestsHandler) GetCartIcon(ctx *gin.Context) {
	// Пытаемся получить userID, если пользователь авторизован
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		// Если не авторизован, возвращаем пустую корзину
		ctx.JSON(http.StatusOK, &models.CartIconDTO{RequestID: nil, Count: 0})
		return
	}

	icon, err := h.Repository.GetCartIcon(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, icon)
}

// 9. GET /api/chrono - список заявок
func (h *RequestsHandler) GetRequests(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isModerator, err := middleware.GetIsModerator(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status := ctx.Query("status")
	dateFrom := ctx.Query("date_from")
	dateTo := ctx.Query("date_to")

	requests, err := h.Repository.GetRequests(userID, isModerator, status, dateFrom, dateTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

// 10. GET /api/chrono/:id - одна заявка
func (h *RequestsHandler) GetRequestByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isModerator, err := middleware.GetIsModerator(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	request, err := h.Repository.GetRequestByID(uint(id), userID, isModerator)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, request)
}

// 11. PUT /api/chrono/:id - изменение полей
func (h *RequestsHandler) UpdateRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Получаем userID из контекста
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input models.UpdateRequestDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Передаём все 3 аргумента: id, userID, dto
	if err := h.Repository.UpdateRequest(uint(id), userID, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// 12. PUT /api/chrono/:id/form - сформировать
func (h *RequestsHandler) FormRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Repository.FormRequest(uint(id), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка формирования"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "formed"})
}

// 13. PUT /api/chrono/:id/complete - завершить
func (h *RequestsHandler) CompleteRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	moderatorID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Repository.CompleteRequest(uint(id), moderatorID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "completed"})
}

// 14. DELETE /api/chrono/:id - удалить
func (h *RequestsHandler) DeleteRequest(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.Repository.DeleteRequest(uint(id), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
