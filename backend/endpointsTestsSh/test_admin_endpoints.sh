#!/bin/bash

# Скрипт для тестирования всех admin эндпоинтов
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZWVhMjQ1ODItMTU3OS00NDAzLWJhOWUtODk2NDY4MWQ1ZGY5IiwibG9naW4iOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJzdWIiOiJlZWEyNDU4Mi0xNTc5LTQ0MDMtYmE5ZS04OTY0NjgxZDVkZjkiLCJleHAiOjE3NTM5NTk5NDYsIm5iZiI6MTc1Mzk1NjM0NiwiaWF0IjoxNzUzOTU2MzQ2fQ.SCdFG3nGmwYv8CEaHr68Bm8vWP0IGK3mK328aiZFhlo"
BASE_URL="http://localhost:3000/api"
HEADER="Authorization: Bearer $TOKEN"

echo "=== Тестирование admin эндпоинтов ==="
echo ""

echo "1. Получение списка администраторов:"
curl -s -H "$HEADER" "$BASE_URL/admin/admins" | jq .
echo ""

echo "2. Получение дашборда администратора:"
curl -s -H "$HEADER" "$BASE_URL/admin/dashboard" | jq .
echo ""

echo "3. Получение списка партнеров:"
curl -s -H "$HEADER" "$BASE_URL/admin/partners" | jq .
echo ""

echo "4. Получение списка купонов:"
curl -s -H "$HEADER" "$BASE_URL/admin/coupons" | jq .
echo ""

echo "5. Получение купонов с пагинацией (limit=5):"
curl -s -H "$HEADER" "$BASE_URL/admin/coupons/paginated?limit=5" | jq .
echo ""

echo "6. Создание нового администратора:"
curl -s -X POST -H "$HEADER" -H "Content-Type: application/json" \
  -d '{"login":"test_admin","password":"Test123456!"}' \
  "$BASE_URL/admin/admins" | jq .
echo ""

echo "7. Создание нового партнера:"
curl -s -X POST -H "$HEADER" -H "Content-Type: application/json" \
  -d '{
    "partner_code":"0001",
    "login":"test_partner",
    "password":"test123456",
    "domain":"test.example.com",
    "brand_name":"Test Brand",
    "email":"test@partner.com",
    "logo_url":"https://example.com/logo.png",
    "ozon_link":"https://ozon.ru/test-brand",
    "wildberries_link":"https://wildberries.ru/test-brand",
    "address":"г. Москва, ул. Тестовая, д. 1",
    "phone":"+79001234567",
    "telegram":"testbrand",
    "whatsapp":"79001234567",
    "telegram_link":"https://t.me/testbrand",
    "whatsapp_link":"https://wa.me/79001234567",
    "allow_sales":true,
    "allow_purchases":true
  }' \
  "$BASE_URL/admin/partners" | jq .
echo ""

# Сохраняем ID партнера для дальнейших тестов
PARTNER_ID=$(curl -s -H "$HEADER" "$BASE_URL/admin/partners" | jq -r '.partners[0].id // empty')

if [ -n "$PARTNER_ID" ] && [ "$PARTNER_ID" != "null" ]; then
  echo "8. Получение информации о партнере ($PARTNER_ID):"
  curl -s -H "$HEADER" "$BASE_URL/admin/partners/$PARTNER_ID" | jq .
  echo ""

  echo "9. Получение статистики партнера ($PARTNER_ID):"
  curl -s -H "$HEADER" "$BASE_URL/admin/partners/$PARTNER_ID/statistics" | jq .
  echo ""

  echo "10. Обновление партнера ($PARTNER_ID):"
  curl -s -X PUT -H "$HEADER" -H "Content-Type: application/json" \
    -d '{"brand_name":"Updated Test Brand"}' \
    "$BASE_URL/admin/partners/$PARTNER_ID" | jq .
  echo ""

  echo "11. Создание купонов для партнера ($PARTNER_ID):"
  curl -s -X POST -H "$HEADER" -H "Content-Type: application/json" \
    -d '{
      "partner_id":"'$PARTNER_ID'",
      "size":"21x30",
      "style":"grayscale",
      "count":3
    }' \
    "$BASE_URL/admin/coupons" | jq .
  echo ""

  # Получаем ID первого купона для тестов
  COUPON_ID=$(curl -s -H "$HEADER" "$BASE_URL/admin/coupons" | jq -r '.coupons[0].id // empty')
  
  if [ -n "$COUPON_ID" ] && [ "$COUPON_ID" != "null" ]; then
    echo "12. Получение информации о купоне ($COUPON_ID):"
    curl -s -H "$HEADER" "$BASE_URL/admin/coupons/$COUPON_ID" | jq .
    echo ""
  fi

  echo "13. Блокировка партнера ($PARTNER_ID):"
  curl -s -X PATCH -H "$HEADER" "$BASE_URL/admin/partners/$PARTNER_ID/block" | jq .
  echo ""

  echo "14. Разблокировка партнера ($PARTNER_ID):"
  curl -s -X PATCH -H "$HEADER" "$BASE_URL/admin/partners/$PARTNER_ID/unblock" | jq .
  echo ""
else
  echo "8-14. Пропускаем тесты с партнером - нет партнеров в системе"
  echo ""
fi

echo "15. Экспорт всех купонов:"
curl -s -H "$HEADER" "$BASE_URL/admin/coupons/export" | head -5
echo "... (показаны первые 5 строк экспорта)"
echo ""

echo "=== Admin тесты завершены ==="
