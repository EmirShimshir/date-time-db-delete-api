package entities

import (
	"time"
)

// CleanupRequest представляет запрос на удаление данных
type CleanupRequest struct {
	TableName  string    `json:"table_name"`
	BeforeDate time.Time `json:"before_date"`
	BatchSize  int       `json:"batch_size"`
}

// CleanupResult представляет результат операции удаления
type CleanupResult struct {
	TableName    string        `json:"table_name"`
	RowsDeleted  int           `json:"rows_deleted"`
	ElapsedTime  time.Duration `json:"elapsed_time"`
	Status       string        `json:"status"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// Validate проверяет корректность запроса
func (r *CleanupRequest) Validate() error {
	if r.TableName == "" {
		return ErrEmptyTableName
	}

	if r.BeforeDate.IsZero() {
		return ErrInvalidDate
	}

	if r.BatchSize <= 0 {
		return ErrInvalidBatchSize
	}

	return nil
}

// Domain errors
var (
	ErrEmptyTableName   = NewDomainError("table name cannot be empty")
	ErrInvalidDate      = NewDomainError("invalid date specified")
	ErrInvalidBatchSize = NewDomainError("batch size must be positive")
)

// DomainError представляет ошибку предметной области
type DomainError struct {
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

func NewDomainError(message string) DomainError {
	return DomainError{Message: message}
}
