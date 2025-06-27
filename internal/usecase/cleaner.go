package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"data-cleaner/internal/models/entities"
	"data-cleaner/internal/models/ports"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type cleanerUseCase struct {
	repo            ports.CleanerRepository
	logger          *zap.Logger
	activeTasksLock sync.RWMutex
	activeTasks     map[string]*entities.CleanupResult
}

// NewCleanerUseCase создает новый экземпляр сервиса очистки данных
func NewCleanerUseCase(repo ports.CleanerRepository, logger *zap.Logger) ports.CleanerUseCase {
	return &cleanerUseCase{
		repo:        repo,
		logger:      logger,
		activeTasks: make(map[string]*entities.CleanupResult),
	}
}

// CleanTable удаляет старые данные из указанной таблицы
func (uc *cleanerUseCase) CleanTable(ctx context.Context, req entities.CleanupRequest) (*entities.CleanupResult, error) {
	// Валидируем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем существование таблицы и индекса
	if err := uc.repo.ValidateTable(ctx, req.TableName); err != nil {
		return nil, fmt.Errorf("table validation failed: %w", err)
	}

	// Пытаемся получить блокировку для таблицы
	acquired, unlock, err := uc.repo.TryAcquireLock(ctx, req.TableName)
	if err != nil {
		return nil, fmt.Errorf("lock acquisition failed: %w", err)
	}

	if !acquired {
		return nil, fmt.Errorf("another process is already cleaning table %s", req.TableName)
	}
	defer unlock()

	// Логируем начало операции
	uc.logger.Info("Starting data cleanup",
		zap.String("table", req.TableName),
		zap.Time("before_date", req.BeforeDate),
		zap.Int("batch_size", req.BatchSize))

	startTime := time.Now()
	result := &entities.CleanupResult{
		TableName:   req.TableName,
		Status:      "in_progress",
		RowsDeleted: 0,
	}

	// Удаляем данные небольшими порциями
	totalDeleted := 0
	for {
		// Устанавливаем таймаут для каждой итерации
		iterCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Удаляем пакет данных
		deleted, err := uc.repo.DeleteBatch(iterCtx, req.TableName, req.BeforeDate, req.BatchSize)
		if err != nil {
			uc.logger.Error("Error deleting batch",
				zap.String("table", req.TableName),
				zap.Error(err))

			result.Status = "failed"
			result.ErrorMessage = err.Error()
			result.ElapsedTime = time.Since(startTime)
			return result, fmt.Errorf("batch deletion failed: %w", err)
		}

		totalDeleted += deleted
		uc.logger.Info("Batch deleted",
			zap.String("table", req.TableName),
			zap.Int("deleted_count", deleted),
			zap.Int("total_deleted", totalDeleted))

		// Если удалили меньше, чем размер пакета, значит данных больше нет
		if deleted < req.BatchSize {
			break
		}

		// Небольшая пауза между пакетами, чтобы снизить нагрузку
		select {
		case <-time.After(100 * time.Millisecond):
			// Продолжаем выполнение
		case <-ctx.Done():
			// Контекст был отменен
			result.Status = "canceled"
			result.RowsDeleted = totalDeleted
			result.ElapsedTime = time.Since(startTime)
			return result, ctx.Err()
		}
	}

	elapsedTime := time.Since(startTime)
	uc.logger.Info("Cleanup completed",
		zap.String("table", req.TableName),
		zap.Int("total_deleted", totalDeleted),
		zap.Duration("duration", elapsedTime))

	result.Status = "completed"
	result.RowsDeleted = totalDeleted
	result.ElapsedTime = elapsedTime

	return result, nil
}

// StartAsyncCleanup запускает асинхронную очистку и возвращает идентификатор задачи
func (uc *cleanerUseCase) StartAsyncCleanup(ctx context.Context, req entities.CleanupRequest) (string, error) {
	// Валидируем запрос
	if err := req.Validate(); err != nil {
		return "", err
	}

	// Генерируем уникальный ID для задачи
	taskID := uuid.New().String()

	// Создаем начальный результат
	result := &entities.CleanupResult{
		TableName: req.TableName,
		Status:    "pending",
	}

	// Сохраняем задачу в списке активных
	uc.activeTasksLock.Lock()
	uc.activeTasks[taskID] = result
	uc.activeTasksLock.Unlock()

	// Создаем новый контекст с таймаутом для асинхронной операции
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)

	// Запускаем очистку в отдельной горутине
	go func() {
		defer cancel()

		// Обновляем статус
		result.Status = "in_progress"

		// Выполняем очистку
		cleanResult, err := uc.CleanTable(cleanupCtx, req)

		// Обновляем результат
		uc.activeTasksLock.Lock()
		if err != nil && cleanResult == nil {
			// Если произошла ошибка и результат не был возвращен
			result.Status = "failed"
			result.ErrorMessage = err.Error()
		} else {
			// Копируем данные из результата
			*result = *cleanResult
		}
		uc.activeTasksLock.Unlock()

		// Очищаем информацию о задаче через некоторое время
		time.AfterFunc(1*time.Hour, func() {
			uc.activeTasksLock.Lock()
			delete(uc.activeTasks, taskID)
			uc.activeTasksLock.Unlock()
		})
	}()

	return taskID, nil
}

// GetCleanupStatus возвращает статус операции очистки по идентификатору
func (uc *cleanerUseCase) GetCleanupStatus(ctx context.Context, taskID string) (*entities.CleanupResult, error) {
	uc.activeTasksLock.RLock()
	defer uc.activeTasksLock.RUnlock()

	result, exists := uc.activeTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	// Возвращаем копию результата
	resultCopy := *result
	return &resultCopy, nil
}
