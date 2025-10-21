package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"rip/internal/app/models"
	"rip/internal/app/repository"
)

type LayersHandler struct {
	Repository *repository.Repository
	Minio      *minio.Client
	Bucket     string
}

func NewLayersHandler(r *repository.Repository, minioClient *minio.Client, bucket string) *LayersHandler {
	return &LayersHandler{
		Repository: r,
		Minio:      minioClient,
		Bucket:     bucket,
	}
}

func (h *LayersHandler) RegisterRoutes(api *gin.RouterGroup) {
	layers := api.Group("/layers")
	{
		layers.GET("", h.GetLayers)                        // 1
		layers.GET("/:id", h.GetLayerByID)                 // 2
		layers.POST("", h.CreateLayer)                     // 3
		layers.PUT("/:id", h.UpdateLayer)                  // 4
		layers.DELETE("/:id", h.DeleteLayer)               // 5
		layers.POST("/:id/image", h.UploadLayerImage)      // 6
		layers.POST("/:id/add-to-request", h.AddToRequest) // 7
	}
}

// 1. GET /api/layers - список с фильтрацией
func (h *LayersHandler) GetLayers(ctx *gin.Context) {
	query := ctx.Query("query")

	layers, err := h.Repository.GetLayers(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, layers)
}

// 2. GET /api/layers/:id - один слой
func (h *LayersHandler) GetLayerByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Слой не найден"})
		return
	}

	ctx.JSON(http.StatusOK, layer)
}

// 3. POST /api/layers - создание
func (h *LayersHandler) CreateLayer(ctx *gin.Context) {
	var input models.CreateLayerDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	layer := models.Layer{
		Name:        input.Name,
		Description: input.Description,
		FromYear:    input.FromYear,
		ToYear:      input.ToYear,
		Words:       input.Words,
		ImageURL:    input.ImageURL,
	}

	if err := h.Repository.CreateLayer(&layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания"})
		return
	}

	ctx.JSON(http.StatusCreated, layer)
}

// 4. PUT /api/layers/:id - изменение
func (h *LayersHandler) UpdateLayer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var input models.UpdateLayerDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Слой не найден"})
		return
	}

	if input.Name != nil {
		layer.Name = *input.Name
	}
	if input.Description != nil {
		layer.Description = *input.Description
	}
	if input.FromYear != nil {
		layer.FromYear = *input.FromYear
	}
	if input.ToYear != nil {
		layer.ToYear = *input.ToYear
	}
	if input.Words != nil {
		layer.Words = *input.Words
	}
	if input.ImageURL != nil {
		layer.ImageURL = input.ImageURL
	}

	if err := h.Repository.UpdateLayer(uint(id), layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}

	ctx.JSON(http.StatusOK, layer)
}

// 5. DELETE /api/layers/:id - удаление
func (h *LayersHandler) DeleteLayer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repository.DeleteLayer(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// 6. POST /api/layers/:id/image - загрузка картинки
func (h *LayersHandler) UploadLayerImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Слой не найден"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка открытия файла"})
		return
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%d_%d%s", id, time.Now().UnixNano(), ext)

	_, err = h.Minio.PutObject(
		context.Background(),
		h.Bucket,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить файл"})
		return
	}

	imageURL := fmt.Sprintf("http://%s/%s/%s", h.Minio.EndpointURL().Host, h.Bucket, objectName)
	layer.ImageURL = &imageURL

	if err := h.Repository.UpdateLayer(uint(id), layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"layer_id": id,
		"url":      imageURL,
	})
}

// 7. POST /api/layers/:id/add-to-request - добавление в заявку
func (h *LayersHandler) AddToRequest(ctx *gin.Context) {
	layerID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// TODO: Получить userID из сессии/JWT
	userID := uint(1)

	if err := h.Repository.AddLayerToRequest(userID, uint(layerID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "added"})
}
