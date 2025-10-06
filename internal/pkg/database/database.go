package database

import (
	"fmt"
	"rip/internal/app/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	dsn := "host=localhost user=admin password=111517 dbname=chrono port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Layer{},
		&models.ResearchRequest{},
		&models.RequestLayer{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка миграции базы данных: %w", err)
	}

	fmt.Println("Подключение к базе данных и миграция прошли успешно!")
	return db, nil
}
