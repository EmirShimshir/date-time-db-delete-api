package ports

import (
	"context"
	"time"
)

// CleanerRepository определяет интерфейс для доступа к данным
type CleanerRepository interface {
	// DeleteBatch удаляет пакет старых записей из указанной таблицы
	DeleteBatch(ctx context.Context, tableName string, beforeDate time.Time, batchSize int) (int, error)

	// TryAcquireLock пытается получить блокировку для таблицы
	TryAcquireLock(ctx context.Context, tableName string) (bool, func(), error)

	// ValidateTable проверяет существование таблицы и наличие индекса по дате
	ValidateTable(ctx context.Context, tableName string) error
}
