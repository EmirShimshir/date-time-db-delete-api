package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server представляет HTTP-сервер
type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

// NewServer создает новый HTTP-сервер
func NewServer(handler *Handler, logger *zap.Logger, port int) *Server {
	// Создаем маршрутизатор
	router := mux.NewRouter()

	// Регистрируем middleware
	router.Use(LoggingMiddleware(logger))

	// Регистрируем маршруты
	handler.RegisterRoutes(router)

	// Создаем HTTP-сервер
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start запускает HTTP-сервер
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("address", s.httpServer.Addr))

	// Запускаем сервер
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop останавливает HTTP-сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")

	// Останавливаем сервер
	return s.httpServer.Shutdown(ctx)
}
