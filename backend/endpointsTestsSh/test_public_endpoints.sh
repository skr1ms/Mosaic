#!/bin/bash

# Скрипт для тестирования public API эндпоинтов
BASE_URL="http://localhost:3000/api"

echo "=== Тестирование public API эндпоинтов ==="
echo ""

echo "1. Получение доступных размеров:"
curl -s "$BASE_URL/sizes" | jq .
echo ""

echo "2. Получение доступных стилей:"
curl -s "$BASE_URL/styles" | jq .
echo ""

echo "3. Получение информации о партнере по домену:"
curl -s "$BASE_URL/partners/example.com/info" | jq .
echo ""

echo "4. Проверка несуществующего купона:"
curl -s "$BASE_URL/coupons/0000-0000-0000" | jq .
echo ""

echo "5. Попытка активации несуществующего купона:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}' \
  "$BASE_URL/coupons/0000-0000-0000/activate" | jq .
echo ""

echo "6. Попытка покупки купона онлайн:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{
    "email":"buyer@example.com",
    "size":"21x30",
    "style":"grayscale",
    "partner_domain":"example.com"
  }' \
  "$BASE_URL/coupons/purchase" | jq .
echo ""

echo "7. Тест загрузки изображения (без файла - получим ошибку):"
curl -s -X POST "$BASE_URL/images/upload" | jq .
echo ""

echo "8. Тест обработки несуществующего изображения:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"style":"grayscale","size":"21x30"}' \
  "$BASE_URL/images/00000000-0000-0000-0000-000000000000/process" | jq .
echo ""

echo "9. Тест редактирования несуществующего изображения:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"brightness":10,"contrast":5}' \
  "$BASE_URL/images/00000000-0000-0000-0000-000000000000/edit" | jq .
echo ""

echo "10. Тест генерации схемы для несуществующего изображения:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"size":"21x30"}' \
  "$BASE_URL/images/00000000-0000-0000-0000-000000000000/generate-schema" | jq .
echo ""

echo "11. Тест получения превью несуществующего изображения:"
curl -s "$BASE_URL/images/00000000-0000-0000-0000-000000000000/preview" | jq .
echo ""

echo "12. Тест получения статуса обработки несуществующего изображения:"
curl -s "$BASE_URL/images/00000000-0000-0000-0000-000000000000/status" | jq .
echo ""

echo "13. Тест скачивания схемы несуществующего изображения:"
curl -s -I "$BASE_URL/images/00000000-0000-0000-0000-000000000000/download"
echo ""

echo "14. Тест отправки схемы на email для несуществующего изображения:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}' \
  "$BASE_URL/images/00000000-0000-0000-0000-000000000000/send-email" | jq .
echo ""

echo "=== Public API тесты завершены ==="
