package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"path/filepath"
	"rip/internal/app/handler"
	"rip/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	templatesPath := filepath.Join("templates")
	r.LoadHTMLGlob(filepath.Join(templatesPath, "*.html"))

	r.Static("/resources", "./resources")
	r.Static("/templates", "./templates")

	r.StaticFile("/styles.css", filepath.Join(templatesPath, "styles.css"))

	r.GET("/", handler.GetOrders)
	r.GET("/chrono_service/:id", handler.GetChronoServiceByID)
	r.GET("/chrono_calc", handler.GetOrderForm)

	r.Run()
	log.Println("Server down")
}
