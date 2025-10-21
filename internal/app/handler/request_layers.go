package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"rip/internal/app/models"
	"rip/internal/app/repository"
)

type RequestLayersHandler struct {
	Repository *repository.Repository
}

func NewRequestLayersHandler(r *repository.Repository) *RequestLayersHandler {
	return &RequestLayersHandler{Repository: r}
}

func (h *RequestLayersHandler) RegisterRoutes(api *gin.RouterGroup) {
	requestLayers := api.Group("/request-layers")
	{
		requestLayers.DELETE("/:request_id/:layer_id", h.RemoveLayerFromRequest) // 15
		requestLayers.PUT("/:request_id/:layer_id", h.UpdateLayerComment)        // 16
	}
}

// 15. DELETE /api/request-layers/:request_id/:layer_id
func (h *RequestLayersHandler) RemoveLayerFromRequest(ctx *gin.Context) {
	requestID, err := strconv.Atoi(ctx.Param("request_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный request_id"})
		return
	}

	layerID, err := strconv.Atoi(ctx.Param("layer_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный layer_id"})
		return
	}

	if err := h.Repository.RemoveLayerFromRequest(uint(requestID), uint(layerID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// 16. PUT /api/request-layers/:request_id/:layer_id
func (h *RequestLayersHandler) UpdateLayerComment(ctx *gin.Context) {
	requestID, err := strconv.Atoi(ctx.Param("request_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный request_id"})
		return
	}

	layerID, err := strconv.Atoi(ctx.Param("layer_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный layer_id"})
		return
	}

	var input models.UpdateLayerCommentDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err := h.Repository.UpdateLayerComment(uint(requestID), uint(layerID), &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}
