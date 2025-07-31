#!/bin/bash
echo "=== Краткий тест партнерских эндпоинтов ==="

# Проверим создание партнера (всегда нужно обновлять токен после логина в админке)
ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZWVhMjQ1ODItMTU3OS00NDAzLWJhOWUtODk2NDY4MWQ1ZGY5IiwibG9naW4iOiJhZG1pbiIsInJvbGUiOiJhZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJzdWIiOiJlZWEyNDU4Mi0xNTc5LTQ0MDMtYmE5ZS04OTY0NjgxZDVkZjkiLCJleHAiOjE3NTM5Njg2MjcsIm5iZiI6MTc1Mzk2NTAyNywiaWF0IjoxNzUzOTY1MDI3fQ.Gv0GQhZ5xT211fxGiAavOMDB-8ygTu3qDidZnK7IeY0"

echo "1. Создание партнера (минимальные данные):"
curl -s -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "partner_code":"0007",
    "login":"quicktest",
    "password":"Partner123!",
    "domain":"https://test.com",
    "brand_name":"Quick Test",
    "email":"quick@test.com"
  }' \
  "http://localhost:3000/api/admin/partners"

echo -e "\n2. Авторизация партнера:"
curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"login":"quicktest","password":"Partner123!"}' \
  "http://localhost:3000/api/login/partner"

echo -e "\n=== Тест завершен ==="
