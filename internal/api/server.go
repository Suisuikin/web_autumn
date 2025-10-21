package api

import (
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"rip/internal/app/handler"
	"rip/internal/app/repository"
	"rip/internal/pkg/database"
)

func StartServer() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)

	layersHandler := handler.NewLayersHandler(repo, repo.Minio, repo.Bucket)
	requestsHandler := handler.NewRequestsHandler(repo)
	requestLayersHandler := handler.NewRequestLayersHandler(repo)
	usersHandler := handler.NewUsersHandler(repo)

	r := gin.Default()

	r.Static("/resources", "./resources")
	r.LoadHTMLGlob(filepath.Join("templates", "*.html"))

	apiGroup := r.Group("/api")

	layersHandler.RegisterRoutes(apiGroup)
	requestsHandler.RegisterRoutes(apiGroup)
	requestLayersHandler.RegisterRoutes(apiGroup)
	usersHandler.RegisterRoutes(apiGroup)

	log.Println("Server started on :8080")
	r.Run(":8080")
}
