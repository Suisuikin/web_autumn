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
		layers.GET("", h.GetLayers)
		layers.GET("/:id", h.GetLayerByID)
		layers.POST("", h.CreateLayer)
		layers.PUT("/:id", h.UpdateLayer)
		layers.DELETE("/:id", h.DeleteLayer)
		layers.POST("/:id/image", h.UploadLayerImage)
	}
}

func (h *LayersHandler) GetLayers(ctx *gin.Context) {
	interval := ctx.Query("interval")

	layers, err := h.Repository.GetLayers(interval)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, layers)
}

func (h *LayersHandler) GetLayerByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID слоя"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil || layer == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Слой не найден"})
		return
	}

	ctx.JSON(http.StatusOK, layer)
}

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
		Status:      "active",
		ImageURL:    input.ImageURL,
	}

	if input.Status != "" {
		layer.Status = input.Status
	}

	if err := h.Repository.CreateLayer(&layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании слоя"})
		return
	}

	ctx.JSON(http.StatusCreated, layer)
}

func (h *LayersHandler) UpdateLayer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID слоя"})
		return
	}

	var input models.UpdateLayerDTO
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil || layer == nil {
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
	if input.Status != nil {
		layer.Status = *input.Status
	}
	if input.ImageURL != nil {
		layer.ImageURL = input.ImageURL
	}

	if err := h.Repository.UpdateLayer(uint(id), layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении слоя"})
		return
	}

	ctx.JSON(http.StatusOK, layer)
}

func (h *LayersHandler) DeleteLayer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID слоя"})
		return
	}

	if err := h.Repository.DeleteLayer(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении слоя"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "удален"})
}

func (h *LayersHandler) UploadLayerImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID слоя"})
		return
	}

	layer, err := h.Repository.GetLayerByID(uint(id))
	if err != nil || layer == nil {
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

	uploadInfo, err := h.Minio.PutObject(
		context.Background(),
		h.Bucket,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить файл в MinIO"})
		return
	}

	imageURL := fmt.Sprintf("http://%s/%s/%s", h.Minio.EndpointURL().Host, h.Bucket, objectName)
	layer.ImageURL = &imageURL

	if err := h.Repository.UpdateLayer(uint(id), layer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить слой"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"layer_id": id,
		"file":     uploadInfo.Key,
		"url":      imageURL,
	})
}
