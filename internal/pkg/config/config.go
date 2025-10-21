package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

type Config struct {
	DB     DBConfig
	Redis  RedisConfig
	JWT    JWTConfig
	Server ServerConfig
	MinIO  MinIOConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type JWTConfig struct {
	Secret        string
	ExpiresIn     time.Duration
	SigningMethod jwt.SigningMethod
}

type ServerConfig struct {
	Port int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

func LoadConfig() (*Config, error) {
	jwtExpiresIn, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRES_IN: %w", err)
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "111517"),
			DBName:   getEnv("DB_NAME", "chrono"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", "redispassword"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpiresIn:     jwtExpiresIn,
			SigningMethod: jwt.SigningMethodHS256,
		},
		Server: ServerConfig{
			Port: serverPort,
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "admin123"),
			Bucket:    getEnv("MINIO_BUCKET", "chrono"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
