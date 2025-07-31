#!/bin/bash

# Скрипт для тестирования auth эндпоинтов
BASE_URL="http://localhost:3000/api"

echo "=== Тестирование auth эндпоинтов ==="
echo ""

echo "1. Авторизация администратора (admin:admin123):"
ADMIN_AUTH=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"admin123"}' \
  "$BASE_URL/login/admin")
echo "$ADMIN_AUTH" | jq .

# Извлекаем access_token для дальнейшего использования
ADMIN_TOKEN=$(echo "$ADMIN_AUTH" | jq -r '.access_token // empty')
ADMIN_REFRESH=$(echo "$ADMIN_AUTH" | jq -r '.refresh_token // empty')
echo ""

if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "null" ]; then
  echo "2. Обновление токена администратора:"
  curl -s -X POST -H "Content-Type: application/json" \
    -d '{"refresh_token":"'$ADMIN_REFRESH'"}' \
    "$BASE_URL/refresh/admin" | jq .
  echo ""
else
  echo "2. Пропускаем обновление токена - авторизация не удалась"
  echo ""
fi

echo "3. Попытка авторизации несуществующего партнера:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"login":"test_partner","password":"test123"}' \
  "$BASE_URL/login/partner" | jq .
echo ""

echo "4. Попытка авторизации с неверными данными:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"wrongpassword"}' \
  "$BASE_URL/login/admin" | jq .
echo ""

echo "5. Попытка обновления с неверным refresh токеном:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"refresh_token":"invalid_token"}' \
  "$BASE_URL/refresh/admin" | jq .
echo ""

echo "=== Auth тесты завершены ==="
