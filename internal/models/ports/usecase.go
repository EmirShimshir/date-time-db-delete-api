package ports

import (
	"context"
	"data-cleaner/internal/models/entities"
)

// CleanerUseCase определяет бизнес-логику очистки данных
type CleanerUseCase interface {
	// CleanTable удаляет старые данные из указанной таблицы
	CleanTable(ctx context.Context, req entities.CleanupRequest) (*entities.CleanupResult, error)

	// StartAsyncCleanup запускает асинхронную очистку и возвращает идентификатор задачи
	StartAsyncCleanup(ctx context.Context, req entities.CleanupRequest) (string, error)

	// GetCleanupStatus возвращает статус операции очистки по идентификатору
	GetCleanupStatus(ctx context.Context, taskID string) (*entities.CleanupResult, error)
}
