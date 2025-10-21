package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"rip/internal/app/middleware"
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

// GetLayers godoc
// @Summary Получить список слоев
// @Description Возвращает список всех активных исторических слоев с возможностью фильтрации
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Param query query string false "Поиск по названию"
// @Success 200 {array} models.Layer "Список слоев"
// @Failure 500 {object} map[string]string "Внутренняя ошибка"
// @Router /layers [get]
func (h *LayersHandler) GetLayers(ctx *gin.Context) {
	query := ctx.Query("query")

	layers, err := h.Repository.GetLayers(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, layers)
}

// GetLayerByID godoc
// @Summary Получить слой по ID
// @Description Возвращает подробную информацию об одном историческом слое
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Param id path int true "ID слоя"
// @Success 200 {object} models.Layer "Данные слоя"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "Слой не найден"
// @Router /layers/{id} [get]
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

// CreateLayer godoc
// @Summary Создать новый слой
// @Description Создание нового исторического слоя (только для модераторов)
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param layer body models.CreateLayerDTO true "Данные слоя"
// @Success 201 {object} models.Layer "Слой создан"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 500 {object} map[string]string "Ошибка создания"
// @Router /layers [post]
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

// UpdateLayer godoc
// @Summary Обновить слой
// @Description Обновление данных исторического слоя (только для модераторов)
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID слоя"
// @Param layer body models.UpdateLayerDTO true "Данные для обновления"
// @Success 200 {object} models.Layer "Слой обновлен"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Слой не найден"
// @Failure 500 {object} map[string]string "Ошибка обновления"
// @Router /layers/{id} [put]
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

// DeleteLayer godoc
// @Summary Удалить слой
// @Description Логическое удаление исторического слоя (только для модераторов)
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID слоя"
// @Success 200 {object} map[string]string "Слой удален"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 500 {object} map[string]string "Ошибка удаления"
// @Router /layers/{id} [delete]
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

// UploadLayerImage godoc
// @Summary Загрузить изображение для слоя
// @Description Загрузка изображения в MinIO и привязка к слою (только для модераторов)
// @Tags Слои (услуги)
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID слоя"
// @Param image formData file true "Файл изображения"
// @Success 200 {object} map[string]interface{} "Изображение загружено"
// @Failure 400 {object} map[string]string "Неверный ID или файл не найден"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Слой не найден"
// @Failure 500 {object} map[string]string "Ошибка загрузки"
// @Router /layers/{id}/image [post]
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

// AddToRequest godoc
// @Summary Добавить слой в заявку
// @Description Добавление исторического слоя в текущую заявку пользователя (корзину)
// @Tags Слои (услуги)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID слоя"
// @Success 200 {object} map[string]string "Слой добавлен в заявку"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Требуется авторизация"
// @Failure 500 {object} map[string]string "Ошибка добавления"
// @Router /layers/{id}/add-to-request [post]
func (h *LayersHandler) AddToRequest(ctx *gin.Context) {
	layerID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	if err := h.Repository.AddLayerToRequest(userID, uint(layerID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "added"})
}
