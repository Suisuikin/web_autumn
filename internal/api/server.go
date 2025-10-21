package api

import (
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"rip/internal/app/handler"
	"rip/internal/app/middleware"
	"rip/internal/app/repository"
	"rip/internal/app/service"
	"rip/internal/pkg/config"
	"rip/internal/pkg/database"
	"rip/internal/pkg/redis"
)

func StartServer() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключаем базу данных
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Подключаем Redis
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Создаем сервисы
	jwtService := service.NewJWTService(cfg.JWT)

	// Создаем репозиторий (только с db)
	repo := repository.NewRepository(db)

	// Создаем middleware
	authMW := middleware.NewAuthMiddleware(jwtService, redisClient)

	// Создаем обработчики
	layersHandler := handler.NewLayersHandler(repo, repo.Minio, repo.Bucket)
	requestsHandler := handler.NewRequestsHandler(repo)
	requestLayersHandler := handler.NewRequestLayersHandler(repo)
	usersHandler := handler.NewUsersHandler(repo, jwtService, redisClient)

	// Настраиваем Gin
	r := gin.Default()

	// Статические файлы и шаблоны
	r.Static("/resources", "./resources")
	r.LoadHTMLGlob(filepath.Join("templates", "*.html"))

	// API группа
	apiGroup := r.Group("/api")

	// Регистрируем роуты
	usersHandler.RegisterRoutes(apiGroup, authMW)

	// Layers - публичное чтение, изменение только для модераторов
	layersGroup := apiGroup.Group("/layers")
	{
		layersGroup.GET("", layersHandler.GetLayers)                                               // Публично
		layersGroup.GET("/:id", layersHandler.GetLayerByID)                                        // Публично
		layersGroup.POST("", authMW.ModeratorOnly(), layersHandler.CreateLayer)                    // Модератор
		layersGroup.PUT("/:id", authMW.ModeratorOnly(), layersHandler.UpdateLayer)                 // Модератор
		layersGroup.DELETE("/:id", authMW.ModeratorOnly(), layersHandler.DeleteLayer)              // Модератор
		layersGroup.POST("/:id/image", authMW.ModeratorOnly(), layersHandler.UploadLayerImage)     // Модератор
		layersGroup.POST("/:id/add-to-request", authMW.AuthRequired(), layersHandler.AddToRequest) // Авторизованный
	}

	// Requests - требуется авторизация
	requestsGroup := apiGroup.Group("/chrono")
	{
		requestsGroup.GET("/cart-icon", authMW.OptionalAuth(), requestsHandler.GetCartIcon)         // Опционально
		requestsGroup.GET("", authMW.AuthRequired(), requestsHandler.GetRequests)                   // Авторизованный
		requestsGroup.GET("/:id", authMW.AuthRequired(), requestsHandler.GetRequestByID)            // Авторизованный
		requestsGroup.PUT("/:id", authMW.AuthRequired(), requestsHandler.UpdateRequest)             // Авторизованный
		requestsGroup.PUT("/:id/form", authMW.AuthRequired(), requestsHandler.FormRequest)          // Авторизованный
		requestsGroup.PUT("/:id/complete", authMW.ModeratorOnly(), requestsHandler.CompleteRequest) // Модератор
		requestsGroup.DELETE("/:id", authMW.AuthRequired(), requestsHandler.DeleteRequest)          // Авторизованный
	}

	// Request-Layers
	requestLayersHandler.RegisterRoutes(apiGroup)

	log.Println("Server started on :8080")
	r.Run(":8080")
}
