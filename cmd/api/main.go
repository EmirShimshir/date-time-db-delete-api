package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"data-cleaner/internal/delivery/http"
	"data-cleaner/internal/pkg/config"
	"data-cleaner/internal/pkg/logger"
	"data-cleaner/internal/pkg/postgres"
	repo "data-cleaner/internal/repository/postgres"
	"data-cleaner/internal/usecase"
)

func main() {
	// Создаем контекст приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Инициализируем логгер
	isDevelopment := os.Getenv("APP_ENV") != "production"
	l, err := logger.NewLogger(isDevelopment)
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	log := l.Named("main")

	// Подключаемся к базе данных
	db, err := postgres.NewPostgresDB(ctx, cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer postgres.CloseDB(db, log)

	// Инициализируем слои приложения
	cleanerRepo := repo.NewPostgresRepository(db, log.Named("repository"))
	cleanerUseCase := usecase.NewCleanerUseCase(cleanerRepo, log.Named("usecase"))
	handler := http.NewHandler(cleanerUseCase, log.Named("handler"))

	// Создаем и запускаем HTTP-сервер
	server := http.NewServer(handler, log.Named("server"), cfg.ServerPort)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	log.Info("Application started")

	// Обрабатываем сигналы остановки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал остановки
	<-quit
	log.Info("Shutting down application...")

	// Даем серверу 30 секунд на завершение текущих запросов
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем сервер
	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Application stopped")
}
