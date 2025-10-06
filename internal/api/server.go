package api

import (
	"github.com/gin-gonic/gin"
	"log"
	"path/filepath"
	"rip/internal/app/handler"
	"rip/internal/app/repository"
	"rip/internal/pkg/database"
)

func StartServer() error {
	log.Println("Starting server")

	db, err := database.ConnectDB()
	if err != nil {
		return err
	}

	repo := repository.NewRepository(db)

	h := handler.NewHandler(repo)

	r := gin.Default()

	templatesPath := filepath.Join("templates")
	r.LoadHTMLGlob(filepath.Join(templatesPath, "*.html"))

	r.Static("/resources", "./resources")
	r.Static("/templates", "./templates")
	r.StaticFile("/styles.css", filepath.Join(templatesPath, "styles.css"))

	r.GET("/", h.GetLayers)
	r.GET("/chrono_details/:id", h.GetLayerByID)
	r.GET("/chrono/:id", h.GetOrderFormByID)

	r.GET("/cart", h.GetCart)
	r.POST("/cart/add/:id", h.AddToCart)
	r.POST("/cart/delete/:id", h.DeleteRequest)
	r.POST("/cart/update/:id", h.UpdateRequest)

	return r.Run()
}
