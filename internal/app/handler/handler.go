package handler

import (
	_ "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	_ "gorm.io/gorm"
	"net/http"
	"rip/internal/app/models"
	"rip/internal/app/repository"
	"time"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetLayers(ctx *gin.Context) {
	searchQuery := ctx.Query("interval")

	var layers []models.Layer
	var err error

	if searchQuery != "" {
		layers, err = h.Repository.SearchLayers(searchQuery)
	} else {
		layers, err = h.Repository.GetLayers()
	}

	if err != nil {
		logrus.Error(err)
		layers = []models.Layer{}
	}

	userID := uint(1) // временно фиксировано
	req, err := h.Repository.GetOpenRequest(userID)
	if err != nil {
		logrus.Error("Ошибка получения открытой заявки: ", err)
	}

	cartCount := 0
	if req != nil {
		cartCount = len(req.Layers)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"layers":    layers,
		"query":     searchQuery,
		"cartCount": cartCount,
	})
}

func (h *Handler) GetLayerByID(ctx *gin.Context) {
	idParam := ctx.Param("id")

	var id uint
	_, err := fmt.Sscanf(idParam, "%d", &id)
	if err != nil {
		logrus.Error("неверный ID: ", err)
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Неверный ID",
		})
		return
	}

	layer, err := h.Repository.GetLayerByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Слой не найден",
		})
		return
	}

	cartCount := 2

	ctx.HTML(http.StatusOK, "chrono_service.html", gin.H{
		"layer":     layer,
		"cartCount": cartCount,
	})
}

func (h *Handler) GetOrderForm(ctx *gin.Context) {
	userID := uint(1)
	req, err := h.Repository.GetOpenRequest(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка получения заявки")
		return
	}

	var layers []models.Layer
	var requestID uint
	if req != nil {
		layers = req.Layers
		requestID = req.ID
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"time":         time.Now().Format("15:04:05"),
		"cartCount":    len(layers),
		"RandomLayers": layers,
		"RequestID":    requestID,
	})
}

func (h *Handler) AddToCart(ctx *gin.Context) {
	var layerID uint
	_, err := fmt.Sscanf(ctx.Param("id"), "%d", &layerID)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный ID слоя")
		return
	}

	userID := uint(1)

	req, err := h.Repository.GetOpenRequest(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка получения заявки")
		return
	}

	if req == nil {
		req, err = h.Repository.CreateNewRequest(userID)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Ошибка создания заявки")
			return
		}
	}

	if err := h.Repository.AddLayerToRequest(req.ID, layerID); err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка добавления слоя")
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) GetCart(ctx *gin.Context) {
	userID := uint(1) // временно

	req, err := h.Repository.GetOpenRequest(userID)
	if err != nil || req == nil {
		req, err = h.Repository.CreateNewRequest(userID)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Ошибка при создании заявки")
			return
		}
	}

	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/chrono/%d", req.ID))
}

func (h *Handler) DeleteRequest(ctx *gin.Context) {
	var requestID uint
	_, err := fmt.Sscanf(ctx.Param("id"), "%d", &requestID)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	if err := h.Repository.CloseRequest(requestID); err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка при закрытии заявки")
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) UpdateRequest(ctx *gin.Context) {
	var requestID uint
	_, err := fmt.Sscanf(ctx.Param("id"), "%d", &requestID)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	if err := ctx.Request.ParseForm(); err != nil {
		ctx.String(http.StatusBadRequest, "Ошибка чтения формы")
		return
	}

	notes := ctx.Request.FormValue("notes")
	if err := h.Repository.UpdateRequestNotes(requestID, notes); err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка при сохранении слов")
		return
	}

	comments := make(map[uint]string)
	for key, values := range ctx.Request.PostForm {
		var layerID uint
		if _, err := fmt.Sscanf(key, "comment_%d", &layerID); err == nil && len(values) > 0 {
			comments[layerID] = values[0]
		}
	}

	if err := h.Repository.UpdateLayerComments(requestID, comments); err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка при сохранении комментариев")
		return
	}

	if err := h.Repository.UpdateRequestStatus(requestID, "рассматривается"); err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка при обновлении статуса заявки")
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) GetOrderFormByID(ctx *gin.Context) {
	var requestID uint
	_, err := fmt.Sscanf(ctx.Param("id"), "%d", &requestID)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный ID заявки")
		return
	}

	req, err := h.Repository.GetRequestByID(requestID)
	if err != nil || req == nil {
		ctx.String(http.StatusNotFound, "Заявка не найдена")
		return
	}

	if req.Status == "удалена" {
		ctx.String(http.StatusNotFound, "Заявка не найдена")
		return
	}

	comments, err := h.Repository.GetLayerComments(req.ID)
	if err != nil {
		comments = make(map[uint]string)
	}

	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"time":          time.Now().Format("15:04:05"),
		"cartCount":     len(req.Layers),
		"RandomLayers":  req.Layers,
		"Comments":      comments,
		"RequestNotes":  notes,
		"RequestID":     req.ID,
		"RequestStatus": req.Status,
	})
}
