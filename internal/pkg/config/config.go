package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит настройки приложения
type Config struct {
	// Настройки HTTP-сервера
	ServerPort int

	// Настройки базы данных
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Настройки пула соединений
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Настройки очистки данных
	DefaultBatchSize int
	MaxRequestTime   time.Duration
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	// Загружаем .env файл, если он существует
	_ = godotenv.Load()

	config := &Config{
		// Значения по умолчанию
		ServerPort:        8080,
		DBMaxOpenConns:    10,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: 5 * time.Minute,
		DefaultBatchSize:  5000,
		MaxRequestTime:    30 * time.Minute,
	}

	// Сервер
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.ServerPort = p
		}
	}

	// База данных
	config.DBHost = getEnv("DB_HOST", "localhost")
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.DBPort = p
		} else {
			config.DBPort = 5432 // По умолчанию
		}
	} else {
		config.DBPort = 5432
	}
	config.DBUser = getEnv("DB_USER", "postgres")
	config.DBPassword = getEnv("DB_PASSWORD", "postgres")
	config.DBName = getEnv("DB_NAME", "postgres")
	config.DBSSLMode = getEnv("DB_SSL_MODE", "disable")

	// Пул соединений
	if val := os.Getenv("DB_MAX_OPEN_CONNS"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			config.DBMaxOpenConns = p
		}
	}
	if val := os.Getenv("DB_MAX_IDLE_CONNS"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			config.DBMaxIdleConns = p
		}
	}
	if val := os.Getenv("DB_CONN_MAX_LIFETIME"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.DBConnMaxLifetime = d
		}
	}

	// Настройки очистки
	if val := os.Getenv("DEFAULT_BATCH_SIZE"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			config.DefaultBatchSize = p
		}
	}
	if val := os.Getenv("MAX_REQUEST_TIME"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.MaxRequestTime = d
		}
	}

	return config, nil
}

// GetDBConnString возвращает строку подключения к PostgreSQL
func (c *Config) GetDBConnString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// Вспомогательная функция для получения переменной окружения с значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
