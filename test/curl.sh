#!/bin/sh

BASE_URL="http://localhost:8080"
TABLE_NAME="users"
BEFORE_DATE_1="2025-04-01T00:00:00Z"
BEFORE_DATE_2="2025-06-01T00:00:00Z"
BATCH_SIZE=10000

echo
echo "Ожидание 5 секунд..."
sleep 5

echo "\n=== Проверяем состояние сервиса ==="
curl -s "$BASE_URL/api/v1/health" | jq .

echo "\n=== Запускаем синхронную очистку ==="
curl -s -X POST "$BASE_URL/api/v1/cleanup" \
  -H "Content-Type: application/json" \
  -d "{\"table_name\":\"$TABLE_NAME\",\"before_date\":\"$BEFORE_DATE_1\",\"batch_size\":$BATCH_SIZE}" | jq .

echo "=== Запускаем асинхронную очистку ==="
async_response=$(curl -s -X POST "$BASE_URL/api/v1/cleanup/async" \
  -H "Content-Type: application/json" \
  -d "{\"table_name\":\"$TABLE_NAME\",\"before_date\":\"$BEFORE_DATE_2\",\"batch_size\":$BATCH_SIZE}")

echo "Ответ асинхронной очистки:"
echo "$async_response" | jq .

# Извлекаем task_id с помощью sed
task_id=$(echo "$async_response" | sed -n 's/.*"task_id":"\([^"]*\)".*/\1/p')

if [ -z "$task_id" ]; then
  echo "Ошибка: не удалось получить task_id"
  exit 1
fi

echo "Извлечён task_id: $task_id"

echo "=== Запрос статуса задачи ==="
curl -s "$BASE_URL/api/v1/cleanup/$task_id" | jq .

echo
echo "Ожидание 15 секунд..."
sleep 15

echo
echo "=== Запрос статуса задачи (второй раз) ==="
curl -s "$BASE_URL/api/v1/cleanup/$task_id" | jq .