package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggingMiddleware логирует информацию о HTTP-запросах
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Генерируем уникальный ID запроса
			requestID := uuid.New().String()

			// Добавляем ID запроса в заголовок ответа
			w.Header().Set("X-Request-ID", requestID)

			// Оборачиваем ResponseWriter для отслеживания статуса ответа
			rw := newResponseWriter(w)

			// Фиксируем время начала запроса
			start := time.Now()

			// Логируем информацию о запросе
			logger.Info("Request started",
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))

			// Вызываем следующий обработчик
			next.ServeHTTP(rw, r)

			// Логируем информацию о завершении запроса
			logger.Info("Request completed",
				zap.String("request_id", requestID),
				zap.Int("status", rw.status),
				zap.Int("bytes", rw.size),
				zap.Duration("duration", time.Since(start)))
		})
	}
}

// ResponseWriter расширяет стандартный ResponseWriter для отслеживания статуса и размера ответа
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
