#!/bin/bash

echo "🚀 Запуск фронтенда для photo.doyoupaint.com..."

# Создаем директорию для логов если её нет
mkdir -p logs/nginx

# Останавливаем существующие контейнеры
echo "📦 Останавливаем существующие контейнеры..."
docker-compose down

# Удаляем старые образы
echo "🧹 Удаляем старые образы..."
docker-compose down --rmi all

# Собираем и запускаем
echo "🔨 Собираем и запускаем новый образ..."
docker-compose up --build -d

# Проверяем статус
echo "✅ Проверяем статус контейнеров..."
docker-compose ps

echo "🎉 Фронтенд запущен!"
echo "🌐 Сайт доступен по адресу: https://photo.doyoupaint.com"
echo "📋 Логи: docker-compose logs -f frontend"
