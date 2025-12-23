package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rip/internal/app/middleware"
	"rip/internal/app/models"
	"rip/internal/app/repository"
	"errors"
	"gorm.io/gorm"
	"log"
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
        requests.POST("/async-result", h.UpdateAsyncResult) // async
    }
}

// Запуск асинхронного расчета (вызывается из CompleteRequest)
// GetChronoForAsync godoc
// @Summary Получить данные Chrono для асинхронного расчета
// @Description Передает полные данные заявки асинхронному сервису
// @Tags Заявки
// @Accept json
// @Produce json
// @Param request body models.AsyncChronoRequest true "ID заявки"
// @Success 200 {object} models.AsyncChronoData "Данные для расчета"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 403 {object} map[string]string "Неверный токен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Router /chrono/calculate-async [post]
func (h *RequestsHandler) StartAsyncCalculation(ctx *gin.Context) {
    var input struct {
        ResearchRequestID uint   `json:"research_request_id" binding:"required"`
        AuthToken         string `json:"auth_token"`
    }

    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
        return
    }

    const AUTH_TOKEN = "111517"
    if input.AuthToken != AUTH_TOKEN {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid auth token"})
        return
    }

    chronoData, err := h.Repository.GetChronoDataForAsync(input.ResearchRequestID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, chronoData)
}

func (h *RequestsHandler) UpdateAsyncResult(ctx *gin.Context) {
    log.Printf("🎯 POST /api/chrono/async-result ПОЛУЧЕН! Method=%s", ctx.Request.Method)

    var input models.AsyncResultDTO
    if err := ctx.ShouldBindJSON(&input); err != nil {
        log.Printf("❌ BindJSON ОШИБКА: %v", err)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
        return
    }

    log.Printf("✅ JSON распарсен: ID=%d, token=%s", input.ResearchRequestID, input.AuthToken)

    const AUTH_TOKEN = "111517"
    if input.AuthToken != AUTH_TOKEN {
        log.Printf("❌ Неверный токен: %s", input.AuthToken)
        ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
        return
    }

    err := h.Repository.UpdateAsyncResult(input.ResearchRequestID, &input)
    if err != nil {
        log.Printf("❌ DB ошибка: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    log.Printf("🎉 ✅ Результат сохранен ID=%d", input.ResearchRequestID)
    ctx.JSON(http.StatusOK, gin.H{"status": "updated", "request_id": input.ResearchRequestID})
}


// GetCartIcon godoc
// @Summary Получить иконку корзины
// @Description Возвращает информацию о текущей заявке пользователя (корзине)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.CartIconDTO "Информация о корзине"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /chrono/cart-icon [get]
func (h *RequestsHandler) GetCartIcon(ctx *gin.Context) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
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

// GetRequests godoc
// @Summary Получить список заявок
// @Description Возвращает список заявок текущего пользователя (или всех для модератора)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Фильтр по статусу"
// @Param date_from query string false "Дата от (YYYY-MM-DD)"
// @Param date_to query string false "Дата до (YYYY-MM-DD)"
// @Success 200 {array} models.ResearchRequest "Список заявок"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /chrono [get]
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

// GetRequestByID godoc
// @Summary Получить заявку по ID
// @Description Возвращает подробную информацию о заявке со всеми слоями
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} models.ResearchRequest "Данные заявки"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Router /chrono/{id} [get]
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

// UpdateRequest godoc
// @Summary Обновить заявку
// @Description Обновление полей заявки (только в статусе draft)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param request body models.UpdateRequestDTO true "Данные для обновления"
// @Success 200 {object} map[string]string "Заявка обновлена"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Ошибка обновления"
// @Router /chrono/{id} [put]
func (h *RequestsHandler) UpdateRequest(ctx *gin.Context) {
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

	var input models.UpdateRequestDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err := h.Repository.UpdateRequest(uint(id), userID, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// FormRequest godoc
// @Summary Сформировать заявку
// @Description Перевод заявки из статуса draft в formed
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка сформирована"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Ошибка формирования"
// @Router /chrono/{id}/form [put]
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

// CompleteRequest godoc
// @Summary Завершить заявку
// @Description Завершение заявки с расчетом совпадений (только для модераторов)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка завершена"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 500 {object} map[string]string "Ошибка завершения"
// @Router /chrono/{id}/complete [put]
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

    isModerator, err := middleware.GetIsModerator(ctx)
    if err != nil || !isModerator {
        ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: только модераторы"})
        return
    }

    if err := h.Repository.CompleteRequest(uint(id), moderatorID); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    go func(requestID uint) {
        err := h.Repository.SendToAsyncService(requestID)
        if err != nil {
            log.Printf("❌ ERROR: Failed to trigger async calculation for ReqID=%d: %v", requestID, err)
        } else {
            log.Printf("🚀 SUCCESS: Triggered async calculation for ReqID=%d", requestID)
        }
    }(uint(id))

    ctx.JSON(http.StatusOK, gin.H{
        "status": "completed",
        "message": "Заявка переведена в статус completed. Запущен научный анализ.",
        "async_started": true,
    })
}


// DeleteRequest godoc
// @Summary Удалить заявку
// @Description Логическое удаление заявки (только в статусе draft)
// @Tags Заявки
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка удалена"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Ошибка удаления"
// @Router /chrono/{id} [delete]
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
