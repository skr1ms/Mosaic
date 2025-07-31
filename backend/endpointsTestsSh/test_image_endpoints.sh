#!/bin/bash

# Скрипт для тестирования image processing эндпоинтов
BASE_URL="http://localhost:3000/api"

echo "=== Тестирование image processing эндпоинтов ==="
echo ""

echo "1. Получение очереди обработки изображений:"
curl -s "$BASE_URL/images/queue" | jq .
echo ""

echo "2. Получение статистики по задачам:"
curl -s "$BASE_URL/images/statistics" | jq .
echo ""

echo "3. Получение следующей задачи для обработки:"
curl -s "$BASE_URL/images/next" | jq .
echo ""

echo "4. Добавление задачи в очередь:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{
    "coupon_id":"00000000-0000-0000-0000-000000000000",
    "image_url":"https://example.com/test-image.jpg",
    "size":"21x30",
    "style":"grayscale",
    "priority":1
  }' \
  "$BASE_URL/images/queue" | jq .
echo ""

# Получаем список задач для дальнейшего тестирования
QUEUE_DATA=$(curl -s "$BASE_URL/images/queue")
TASK_ID=$(echo "$QUEUE_DATA" | jq -r '.tasks[0].id // empty')

if [ -n "$TASK_ID" ] && [ "$TASK_ID" != "null" ] && [ "$TASK_ID" != "empty" ]; then
  echo "5. Найдена задача $TASK_ID, тестируем операции с ней:"
  
  echo "  a) Получение задачи по ID:"
  curl -s "$BASE_URL/images/queue/$TASK_ID" | jq .
  echo ""
  
  echo "  b) Начало обработки задачи:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"processor_id":"test_processor"}' \
    "$BASE_URL/images/queue/$TASK_ID/start" | jq .
  echo ""
  
  echo "  c) Завершение обработки задачи:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{
      "preview_url":"https://example.com/preview.jpg",
      "schema_url":"https://example.com/schema.pdf"
    }' \
    "$BASE_URL/images/queue/$TASK_ID/complete" | jq .
  echo ""
  
  echo "  d) Попытка повтора задачи:"
  curl -s -X PUT "$BASE_URL/images/queue/$TASK_ID/retry" | jq .
  echo ""
  
  echo "  e) Провал обработки задачи:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"error":"Test error message"}' \
    "$BASE_URL/images/queue/$TASK_ID/fail" | jq .
  echo ""
  
  echo "  f) Удаление задачи:"
  curl -s -X DELETE "$BASE_URL/images/queue/$TASK_ID" | jq .
  echo ""
else
  echo "5. Тестируем операции с фиктивной задачей:"
  
  FAKE_ID="00000000-0000-0000-0000-000000000000"
  
  echo "  a) Получение несуществующей задачи:"
  curl -s "$BASE_URL/images/queue/$FAKE_ID" | jq .
  echo ""
  
  echo "  b) Попытка начала обработки несуществующей задачи:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"processor_id":"test_processor"}' \
    "$BASE_URL/images/queue/$FAKE_ID/start" | jq .
  echo ""
  
  echo "  c) Попытка завершения несуществующей задачи:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"preview_url":"https://example.com/preview.jpg"}' \
    "$BASE_URL/images/queue/$FAKE_ID/complete" | jq .
  echo ""
  
  echo "  d) Попытка удаления несуществующей задачи:"
  curl -s -X DELETE "$BASE_URL/images/queue/$FAKE_ID" | jq .
  echo ""
fi

echo "=== Image processing тесты завершены ==="
