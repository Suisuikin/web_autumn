package handler

import (
	"net/http"
	"rip/internal/app/models"
	"rip/internal/app/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/sirupsen/logrus"
)

type ChronoHandler struct {
	Repository *repository.Repository
}

func NewRequestsHandler(r *repository.Repository) *ChronoHandler {
	return &ChronoHandler{Repository: r}
}

func (h *ChronoHandler) RegisterRoutes(api *gin.RouterGroup) {
	requests := api.Group("/chrono")
	{
		requests.GET("/cart", h.GetCartBadge)
		requests.GET("", h.ListRequests)
		requests.POST("/draft/layers/:layer_id", h.AddLayerToDraft)

		requestByID := requests.Group("/:id")
		{
			requestByID.GET("", h.GetRequest)
			requestByID.PUT("", h.UpdateRequest)
			requestByID.DELETE("", h.DeleteRequest)
			requestByID.PUT("/form", h.FormRequest)
			requestByID.PUT("/resolve", h.ResolveRequest)

			layersInRequest := requestByID.Group("/layers/:layer_id")
			{
				layersInRequest.PUT("", h.UpdateRequestLayer)
				layersInRequest.DELETE("", h.RemoveRequestLayer)
			}
		}
	}
}

func (h *ChronoHandler) ListRequests(ctx *gin.Context) {
	requests, err := h.Repository.ListRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить заявки"})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

func (h *ChronoHandler) GetCartBadge(ctx *gin.Context) {
	req, err := h.Repository.GetCart()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения корзины"})
		return
	}
	ctx.JSON(http.StatusOK, req)
}

func (h *ChronoHandler) AddLayerToDraft(ctx *gin.Context) {
	layerID, _ := strconv.Atoi(ctx.Param("layer_id"))
	req, err := h.Repository.AddLayerToOpenRequest(uint(layerID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить слой"})
		return
	}
	ctx.JSON(http.StatusOK, req)
}

func (h *ChronoHandler) GetRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	req, err := h.Repository.GetRequest(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, req)
}

func (h *ChronoHandler) UpdateRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var input models.ResearchRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := h.Repository.UpdateRequest(uint(id), &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заявку"})
		return
	}
	ctx.JSON(http.StatusOK, input)
}

func (h *ChronoHandler) DeleteRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.DeleteRequest(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить заявку"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "удалена"})
}

func (h *ChronoHandler) FormRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.FormRequest(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сформировать заявку"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "сформирована"})
}

func (h *ChronoHandler) ResolveRequest(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repository.ResolveRequest(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассмотреть заявку"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "рассмотрена"})
}

func (h *ChronoHandler) UpdateRequestLayer(ctx *gin.Context) {
	reqID, _ := strconv.Atoi(ctx.Param("id"))
	layerID, _ := strconv.Atoi(ctx.Param("layer_id"))

	var input struct {
		Comment *string `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	err := h.Repository.UpdateRequestLayer(uint(reqID), uint(layerID), input.Comment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить слой"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "обновлено"})
}

func (h *ChronoHandler) RemoveRequestLayer(ctx *gin.Context) {
	reqID, _ := strconv.Atoi(ctx.Param("id"))
	layerID, _ := strconv.Atoi(ctx.Param("layer_id"))
	err := h.Repository.RemoveLayerFromRequest(uint(reqID), uint(layerID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить слой"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "удален"})
}
