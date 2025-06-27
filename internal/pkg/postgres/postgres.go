package postgres

import (
	"context"

	_ "github.com/jackc/pgx/v4/stdlib" // Драйвер PostgreSQL
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"data-cleaner/internal/pkg/config"
)

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*sqlx.DB, error) {
	// Создаем подключение
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.GetDBConnString())
	if err != nil {
		return nil, err
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	// Проверяем соединение
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.DBHost),
		zap.Int("port", cfg.DBPort),
		zap.String("database", cfg.DBName))

	return db, nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB(db *sqlx.DB, logger *zap.Logger) {
	if err := db.Close(); err != nil {
		logger.Error("Error closing database connection", zap.Error(err))
	} else {
		logger.Info("Database connection closed")
	}
}
