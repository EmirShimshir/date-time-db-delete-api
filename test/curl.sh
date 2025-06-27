#!/bin/sh

BASE_URL="http://localhost:8080"
BATCH_SIZE=10000

run_cleanup() {
  TABLE_NAME=$1
  BEFORE_DATE_1="2025-04-01T00:00:00Z"
  BEFORE_DATE_2="2025-06-01T00:00:00Z"

  echo "\n=== Проверяем состояние сервиса для $TABLE_NAME ==="
  curl -s "$BASE_URL/api/v1/health" | jq .

  echo "\n=== Запускаем синхронную очистку для $TABLE_NAME ==="
  curl -s -X POST "$BASE_URL/api/v1/cleanup" \
    -H "Content-Type: application/json" \
    -d "{\"table_name\":\"$TABLE_NAME\",\"before_date\":\"$BEFORE_DATE_1\",\"batch_size\":$BATCH_SIZE}" | jq .

  echo "=== Запускаем асинхронную очистку для $TABLE_NAME ==="
  async_response=$(curl -s -X POST "$BASE_URL/api/v1/cleanup/async" \
    -H "Content-Type: application/json" \
    -d "{\"table_name\":\"$TABLE_NAME\",\"before_date\":\"$BEFORE_DATE_2\",\"batch_size\":$BATCH_SIZE}")

  echo "Ответ асинхронной очистки для $TABLE_NAME:"
  echo "$async_response" | jq .

  task_id=$(echo "$async_response" | sed -n 's/.*"task_id":"\([^"]*\)".*/\1/p')

  if [ -z "$task_id" ]; then
    echo "Ошибка: не удалось получить task_id для $TABLE_NAME"
    return 1
  fi

  echo "Извлечён task_id для $TABLE_NAME: $task_id"

  echo "=== Запрос статуса задачи для $TABLE_NAME (первый раз) ==="
  curl -s "$BASE_URL/api/v1/cleanup/$task_id" | jq .

  echo "\nОжидание 15 секунд для $TABLE_NAME..."
  sleep 15

  echo "\n=== Запрос статуса задачи для $TABLE_NAME (второй раз) ==="
  curl -s "$BASE_URL/api/v1/cleanup/$task_id" | jq .
}

# Запускаем одновременно для двух таблиц в фоне
run_cleanup users &
run_cleanup products &

# Ждём завершения обоих процессов
wait

echo "Все очистки завершены."
