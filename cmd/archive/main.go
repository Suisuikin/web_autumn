package main

import (
	"math/rand"
	"time"

	_ "rip/docs" // Импорт сгенерированной документации
	"rip/internal/api"
)

// @title Chrono Research Service API
// @version 1.0
// @description API для системы хронологического анализа исторических слоев
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@chrono.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 127.0.0.1:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {ваш_JWT_токен}

func main() {
	rand.Seed(time.Now().UnixNano())
	api.StartServer()
}
