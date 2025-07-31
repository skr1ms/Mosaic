#!/bin/bash

# Скрипт для тестирования payment эндпоинтов
BASE_URL="http://localhost:3000/api"

echo "=== Тестирование payment эндпоинтов ==="
echo ""

echo "1. Получение доступных опций оплаты:"
curl -s "$BASE_URL/payment/options" | jq .
echo ""

echo "2. Попытка покупки купона (без реальных данных):"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{
    "coupon_id":"00000000-0000-0000-0000-000000000000",
    "email":"buyer@example.com",
    "size":"21x30",
    "style":"grayscale"
  }' \
  "$BASE_URL/payment/purchase" | jq .
echo ""

echo "3. Проверка статуса несуществующего заказа:"
curl -s "$BASE_URL/payment/orders/ORDER123456/status" | jq .
echo ""

echo "4. Тест возврата с оплаты (без параметров):"
curl -s "$BASE_URL/payment/return" | jq .
echo ""

echo "5. Тест возврата с оплаты (с фиктивными параметрами):"
curl -s "$BASE_URL/payment/return?order_id=TEST123&status=success" | jq .
echo ""

echo "6. Тест уведомления о платеже (POST notification):"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"order_id":"TEST123","status":"paid","amount":"1000"}' \
  "$BASE_URL/payment/notification" | jq .
echo ""

echo "=== Payment тесты завершены ==="
