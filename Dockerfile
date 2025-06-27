FROM golang:1.23-alpine as builder

WORKDIR /app

# Копируем зависимости и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o data-cleaner ./cmd/api

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем исполняемый файл из промежуточного образа
COPY --from=builder /app/data-cleaner .

# Создаем непривилегированного пользователя
RUN adduser -D -g '' appuser
USER appuser

# Запускаем приложение
CMD ["./data-cleaner"]