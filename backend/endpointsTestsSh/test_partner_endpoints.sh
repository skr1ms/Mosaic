#!/bin/bash

# Скрипт для тестирования partner эндпоинтов
ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZWVhMjQ1ODItMTU3OS00NDAzLWJhOWUtODk2NDY4MWQ1ZGY5IiwibG9naW4iOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJzdWIiOiJlZWEyNDU4Mi0xNTc5LTQ0MDMtYmE5ZS04OTY0NjgxZDVkZjkiLCJleHAiOjE3NTM5NTk5NDYsIm5iZiI6MTc1Mzk1NjM0NiwiaWF0IjoxNzUzOTU2MzQ2fQ.SCdFG3nGmwYv8CEaHr68Bm8vWP0IGK3mK328aiZFhlo"
BASE_URL="http://localhost:3000/api"
ADMIN_HEADER="Authorization: Bearer $ADMIN_TOKEN"

echo "=== Тестирование partner эндпоинтов ==="
echo ""

# Сначала создаем партнера для тестов
echo "Создаем тестового партнера..."
PARTNER_CREATE=$(curl -s -X POST -H "$ADMIN_HEADER" -H "Content-Type: application/json" \
  -d '{
    "partner_code":"0002",
    "login":"test_partner_endpoints",
    "password":"partner123456",
    "domain":"testpartner.example.com",
    "brand_name":"Test Partner Brand",
    "email":"testpartner@example.com",
    "logo_url":"https://example.com/partner-logo.png",
    "ozon_link":"https://ozon.ru/test-partner",
    "wildberries_link":"https://wildberries.ru/test-partner",
    "address":"г. Санкт-Петербург, ул. Партнерская, д. 2",
    "phone":"+79005678901",
    "telegram":"testpartner",
    "whatsapp":"79005678901",
    "telegram_link":"https://t.me/testpartner",
    "whatsapp_link":"https://wa.me/79005678901",
    "allow_sales":true,
    "allow_purchases":true
  }' \
  "$BASE_URL/admin/partners")

echo "Партнер создан: $PARTNER_CREATE" | jq .
echo ""

# Авторизуемся как партнер
echo "1. Авторизация партнера:"
PARTNER_AUTH=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"login":"test_partner_endpoints","password":"partner123456"}' \
  "$BASE_URL/login/partner")
echo "$PARTNER_AUTH" | jq .

PARTNER_TOKEN=$(echo "$PARTNER_AUTH" | jq -r '.access_token // empty')
echo ""

if [ -n "$PARTNER_TOKEN" ] && [ "$PARTNER_TOKEN" != "null" ]; then
  PARTNER_HEADER="Authorization: Bearer $PARTNER_TOKEN"

  echo "2. Получение дашборда партнера:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/dashboard" | jq .
  echo ""

  echo "3. Получение профиля партнера:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/profile" | jq .
  echo ""

  echo "4. Обновление профиля партнера:"
  curl -s -X PUT -H "$PARTNER_HEADER" -H "Content-Type: application/json" \
    -d '{"brand_name":"Updated Partner Brand","phone":"+7900123456"}' \
    "$BASE_URL/partner/profile" | jq .
  echo ""

  echo "5. Получение купонов партнера:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/coupons" | jq .
  echo ""

  echo "6. Получение статистики партнера:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/statistics" | jq .
  echo ""

  echo "7. Получение статистики продаж:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/statistics/sales" | jq .
  echo ""

  echo "8. Получение статистики использования:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/statistics/usage" | jq .
  echo ""

  echo "9. Экспорт купонов партнера:"
  curl -s -H "$PARTNER_HEADER" "$BASE_URL/partner/coupons/export" | head -3
  echo "... (показаны первые 3 строки экспорта)"
  echo ""

  echo "10. Обновление пароля партнера:"
  curl -s -X PUT -H "$PARTNER_HEADER" -H "Content-Type: application/json" \
    -d '{"current_password":"partner123456","new_password":"newpassword123"}' \
    "$BASE_URL/partner/update/password" | jq .
  echo ""
else
  echo "Пропускаем тесты партнера - авторизация не удалась"
  echo ""
fi

echo "11. Запрос на сброс пароля:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"email":"testpartner@example.com"}' \
  "$BASE_URL/partner/forgot" | jq .
echo ""

echo "12. Попытка сброса пароля (с фиктивным токеном):"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"token":"fake_token","new_password":"newpass123"}' \
  "$BASE_URL/partner/reset" | jq .
echo ""

echo "=== Partner тесты завершены ==="
