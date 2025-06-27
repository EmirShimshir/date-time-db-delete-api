package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"data-cleaner/internal/models/ports"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type postgresRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPostgresRepository создает новый экземпляр PostgreSQL репозитория
func NewPostgresRepository(db *sqlx.DB, logger *zap.Logger) ports.CleanerRepository {
	return &postgresRepository{
		db:     db,
		logger: logger,
	}
}

// DeleteBatch реализует удаление данных небольшими порциями
func (r *postgresRepository) DeleteBatch(ctx context.Context, tableName string, beforeDate time.Time, batchSize int) (int, error) {
	// Санитизация имени таблицы
	if !r.isValidTableName(tableName) {
		return 0, fmt.Errorf("invalid table name: %s", tableName)
	}

	// Использование CTE для эффективного удаления с минимальной блокировкой
	query := fmt.Sprintf(`
		WITH rows_to_delete AS (
			SELECT id FROM %s
			WHERE created_at < $1
			ORDER BY created_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM %s
		WHERE id IN (SELECT id FROM rows_to_delete)
		RETURNING id;
	`, tableName, tableName)

	// Начинаем транзакцию с уровнем изоляции READ COMMITTED
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Выполняем запрос
	rows, err := tx.QueryxContext(ctx, query, beforeDate, batchSize)
	if err != nil {
		return 0, fmt.Errorf("execute delete query: %w", err)
	}
	defer rows.Close()

	// Считаем количество удаленных строк
	var count int
	for rows.Next() {
		count++
	}

	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("process result rows: %w", err)
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return count, nil
}

// TryAcquireLock пытается получить advisory lock для таблицы
func (r *postgresRepository) TryAcquireLock(ctx context.Context, tableName string) (bool, func(), error) {
	// Генерируем уникальный ID для блокировки на основе имени таблицы
	lockID := r.generateLockID(tableName)

	var acquired bool
	err := r.db.GetContext(ctx, &acquired, "SELECT pg_try_advisory_lock($1)", lockID)
	if err != nil {
		return false, nil, fmt.Errorf("acquire advisory lock: %w", err)
	}

	if !acquired {
		return false, nil, nil
	}

	// Возвращаем функцию для освобождения блокировки
	unlock := func() {
		var released bool
		err := r.db.Get(&released, "SELECT pg_advisory_unlock($1)", lockID)
		if err != nil {
			r.logger.Error("Failed to release advisory lock",
				zap.String("table", tableName),
				zap.Uint64("lock_id", uint64(lockID)),
				zap.Error(err))
		}
	}

	return true, unlock, nil
}

// ValidateTable проверяет существование таблицы и наличие индекса по дате
func (r *postgresRepository) ValidateTable(ctx context.Context, tableName string) error {
	// Проверяем существование таблицы
	var exists bool
	err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`, tableName)
	if err != nil {
		return fmt.Errorf("check table existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// Проверяем наличие индекса по created_at
	err = r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT FROM pg_indexes
			WHERE tablename = $1
			AND indexdef LIKE '%created_at%'
		)
	`, tableName)
	if err != nil {
		return fmt.Errorf("check index existence: %w", err)
	}

	if !exists {
		r.logger.Warn("Table doesn't have index on created_at column, operation may be slow",
			zap.String("table", tableName))
	}

	return nil
}

// Вспомогательные функции

// generateLockID генерирует уникальный ID для advisory lock
func (r *postgresRepository) generateLockID(tableName string) int64 {
	h := fnv.New64a()
	h.Write([]byte(tableName))
	u := h.Sum64()
	return int64(u)
}

// isValidTableName проверяет, является ли имя таблицы безопасным для использования в SQL
func (r *postgresRepository) isValidTableName(name string) bool {
	// Простая проверка имени таблицы - только буквы, цифры и подчеркивания
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	// Проверка на SQL-инъекции
	forbidden := []string{";", "--", "/*", "*/", "drop", "delete", "insert", "update"}
	nameLower := strings.ToLower(name)
	for _, word := range forbidden {
		if strings.Contains(nameLower, word) {
			return false
		}
	}

	return true
}
