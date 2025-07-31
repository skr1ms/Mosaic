#!/bin/bash

# Скрипт для тестирования всех эндпоинтов stats
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZWVhMjQ1ODItMTU3OS00NDAzLWJhOWUtODk2NDY4MWQ1ZGY5IiwibG9naW4iOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJzdWIiOiJlZWEyNDU4Mi0xNTc5LTQ0MDMtYmE5ZS04OTY0NjgxZDVkZjkiLCJleHAiOjE3NTM5NTk5NDYsIm5iZiI6MTc1Mzk1NjM0NiwiaWF0IjoxNzUzOTU2MzQ2fQ.SCdFG3nGmwYv8CEaHr68Bm8vWP0IGK3mK328aiZFhlo"
BASE_URL="http://localhost:3000/api"
HEADER="Authorization: Bearer $TOKEN"

echo "=== Тестирование всех эндпоинтов stats ==="
echo ""

echo "1. Общая статистика:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/general" | jq .
echo ""

echo "2. Статистика всех партнеров:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/partners" | jq .
echo ""

echo "3. Временная статистика (по умолчанию):"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/time-series" | jq .
echo ""

echo "4. Временная статистика (по неделям):"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/time-series?period=week" | jq .
echo ""

echo "5. Временная статистика (с датами):"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/time-series?period=day&date_from=2025-07-01&date_to=2025-07-31" | jq .
echo ""

echo "6. Системное здоровье:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/system-health" | jq .
echo ""

echo "7. Купоны по статусам:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/coupons-by-status" | jq .
echo ""

echo "8. Купоны по размерам:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/coupons-by-size" | jq .
echo ""

echo "9. Купоны по стилям:"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/coupons-by-style" | jq .
echo ""

echo "10. Топ партнеров (лимит 5):"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/top-partners?limit=5" | jq .
echo ""

echo "11. Топ партнеров (лимит 10):"
curl -s -H "$HEADER" "$BASE_URL/admin/stats/top-partners?limit=10" | jq .
echo ""

echo "=== Все тесты завершены ==="
