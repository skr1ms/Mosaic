#!/bin/bash

# Скрипт для тестирования coupon эндпоинтов (публичные API)
BASE_URL="http://localhost:3000/api"

echo "=== Тестирование coupon эндпоинтов ==="
echo ""

echo "1. Получение списка купонов (публичный API):"
curl -s "$BASE_URL/coupons/" | jq .
echo ""

echo "2. Получение купонов с пагинацией:"
curl -s "$BASE_URL/coupons/paginated?limit=5&offset=0" | jq .
echo ""

echo "3. Получение статистики по купонам (без параметров):"
curl -s "$BASE_URL/coupons/statistics" | jq .
echo ""

echo "3a. Получение статистики по купонам с partner_id:"
curl -s "$BASE_URL/coupons/statistics?partner_id=dbc783f7-7986-4976-9f24-3f88b4d00e42" | jq .
echo ""

echo "4. Попытка получения купона по несуществующему ID:"
curl -s "$BASE_URL/coupons/00000000-0000-0000-0000-000000000000" | jq .
echo ""

echo "5. Попытка получения купона по несуществующему коду:"
curl -s "$BASE_URL/coupons/code/0000-0000-0000" | jq .
echo ""

echo "6. Валидация несуществующего купона:"
curl -s -X POST "$BASE_URL/coupons/code/0000-0000-0000/validate" | jq .
echo ""

echo "7. Экспорт купонов (без параметров):"
curl -s "$BASE_URL/coupons/export" | head -3
echo ""

echo "7a. Экспорт купонов с partner_id:"
curl -s "$BASE_URL/coupons/export?partner_id=dbc783f7-7986-4976-9f24-3f88b4d00e42" | head -3
echo ""

echo "8. Получение купонов по несуществующему партнеру:"
curl -s "$BASE_URL/coupons/partner/00000000-0000-0000-0000-000000000000" | jq .
echo ""

# Если есть купоны в системе, тестируем операции с ними
# Сначала получаем список купонов
COUPONS_RESPONSE=$(curl -s "$BASE_URL/coupons/")

# Проверяем структуру ответа - API возвращает массив, а не объект с полем coupons
COUPON_ID=$(echo "$COUPONS_RESPONSE" | jq -r '.[0].id // empty' 2>/dev/null)
COUPON_CODE=$(echo "$COUPONS_RESPONSE" | jq -r '.[0].code // empty' 2>/dev/null)

if [ -n "$COUPON_ID" ] && [ "$COUPON_ID" != "null" ] && [ "$COUPON_ID" != "empty" ]; then
  echo "9. Найден купон $COUPON_ID, тестируем операции с ним:"
  
  echo "  a) Получение купона по ID:"
  curl -s "$BASE_URL/coupons/$COUPON_ID" | jq .
  echo ""
  
  echo "  b) Получение купона по коду ($COUPON_CODE):"
  curl -s "$BASE_URL/coupons/code/$COUPON_CODE" | jq .
  echo ""
  
  echo "  c) Валидация купона:"
  curl -s -X POST "$BASE_URL/coupons/code/$COUPON_CODE/validate" | jq .
  echo ""
  
  echo "  d) Попытка активации купона (с URL изображений):"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{
      "original_image_url":"https://example.com/original.jpg",
      "preview_url":"https://example.com/preview.jpg",
      "schema_url":"https://example.com/schema.jpg"
    }' \
    "$BASE_URL/coupons/$COUPON_ID/activate" | jq .
  echo ""
  
  echo "  e) Попытка отправки схемы:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"email":"test@example.com"}' \
    "$BASE_URL/coupons/$COUPON_ID/send-schema" | jq .
  echo ""
  
  echo "  f) Попытка пометки как купленного:"
  curl -s -X PUT -H "Content-Type: application/json" \
    -d '{"purchase_email":"buyer@example.com"}' \
    "$BASE_URL/coupons/$COUPON_ID/purchase" | jq .
  echo ""
  
  echo "  g) Скачивание материалов купона (показать заголовки):"
  curl -s -I "$BASE_URL/coupons/$COUPON_ID/download-materials"
  echo ""
  
  echo "  h) Сброс купона:"
  curl -s -X PUT "$BASE_URL/coupons/$COUPON_ID/reset" | jq .
  echo ""
else
  echo "9. Нет купонов в системе для детального тестирования"
  echo ""
fi

echo "=== Coupon тесты завершены ==="
