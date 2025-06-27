package http

import (
	"context"
	"data-cleaner/internal/models/entities"
	"data-cleaner/internal/models/ports"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	cleanerUseCase ports.CleanerUseCase
	logger         *zap.Logger
}

// NewHandler создает новый обработчик HTTP-запросов
func NewHandler(uc ports.CleanerUseCase, logger *zap.Logger) *Handler {
	return &Handler{
		cleanerUseCase: uc,
		logger:         logger,
	}
}

// RegisterRoutes регистрирует пути API
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/cleanup", h.HandleCleanup).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/cleanup/async", h.HandleAsyncCleanup).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/cleanup/{taskID}", h.HandleGetCleanupStatus).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/health", h.HandleHealthCheck).Methods(http.MethodGet)
}

// HandleCleanup обрабатывает синхронный запрос на очистку данных
func (h *Handler) HandleCleanup(w http.ResponseWriter, r *http.Request) {
	var req entities.CleanupRequest

	// Декодируем JSON-запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Устанавливаем значения по умолчанию, если необходимо
	if req.BatchSize == 0 {
		req.BatchSize = 5000 // Значение по умолчанию
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Выполняем очистку
	result, err := h.cleanerUseCase.CleanTable(ctx, req)
	if err != nil {
		if _, ok := err.(entities.DomainError); ok {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			h.logger.Error("Cleanup error", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// HandleAsyncCleanup обрабатывает асинхронный запрос на очистку данных
func (h *Handler) HandleAsyncCleanup(w http.ResponseWriter, r *http.Request) {
	var req entities.CleanupRequest

	// Декодируем JSON-запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Устанавливаем значения по умолчанию, если необходимо
	if req.BatchSize == 0 {
		req.BatchSize = 5000 // Значение по умолчанию
	}

	// Запускаем асинхронную очистку
	taskID, err := h.cleanerUseCase.StartAsyncCleanup(r.Context(), req)
	if err != nil {
		if _, ok := err.(entities.DomainError); ok {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			h.logger.Error("Async cleanup error", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusAccepted, map[string]string{
		"task_id":    taskID,
		"status":     "pending",
		"status_url": "/api/v1/cleanup/" + taskID,
	})
}

// HandleGetCleanupStatus возвращает статус операции очистки
func (h *Handler) HandleGetCleanupStatus(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID задачи из URL
	vars := mux.Vars(r)
	taskID := vars["taskID"]

	// Получаем статус
	result, err := h.cleanerUseCase.GetCleanupStatus(r.Context(), taskID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// HandleHealthCheck проверяет работоспособность сервиса
func (h *Handler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Вспомогательные функции для ответов

func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Кодируем ответ в JSON
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("Error encoding response", zap.Error(err))
	}
}
