package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
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

func (h *Handler) GetOrders(ctx *gin.Context) {
	searchQuery := ctx.Query("query")

	var orders []repository.Order
	var err error

	if searchQuery != "" {
		orders, err = h.Repository.SearchOrders(searchQuery)
	} else {
		orders, err = h.Repository.GetOrders()
	}

	if err != nil {
		logrus.Error(err)
	}

	cartCount := 2 //len(getCartFromSession(ctx))

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"orders":    orders,
		"query":     searchQuery,
		"cartCount": cartCount,
	})
}

func (h *Handler) GetChronoServiceByID(ctx *gin.Context) {
	id := ctx.Param("id")

	order, err := h.Repository.GetOrderByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Услуга не найдена",
		})
		return
	}

	cartCount := 2 //len(getCartFromSession(ctx))

	ctx.HTML(http.StatusOK, "chrono_service.html", gin.H{
		"chrono":    order,
		"cartCount": cartCount,
	})
}

func (h *Handler) GetOrderForm(ctx *gin.Context) {
	cartCount := 2 //len(getCartFromSession(ctx))

	allServices, err := h.Repository.GetOrders()
	if err != nil {
		logrus.Error("Ошибка получения случайных карточек: ", err)
		allServices = []repository.Order{}
	}

	rand.Shuffle(len(allServices), func(i, j int) {
		allServices[i], allServices[j] = allServices[j], allServices[i]
	})

	numRandomServices := 2
	if len(allServices) < numRandomServices {
		numRandomServices = len(allServices)
	}
	randomServices := allServices[:numRandomServices]

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"time":           time.Now().Format("15:04:05"),
		"cartCount":      cartCount,
		"RandomServices": randomServices,
	})
}
