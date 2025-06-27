package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger создает и настраивает новый логгер
func NewLogger(isDevelopment bool) (*zap.Logger, error) {
	var config zap.Config

	if isDevelopment {
		// Для разработки используем более читаемый формат
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Для продакшна используем JSON формат
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Настраиваем логирование стеков ошибок
	config.DisableStacktrace = false

	// Создаем логгер
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Заменяем глобальный логгер
	zap.ReplaceGlobals(logger)

	return logger, nil
}
