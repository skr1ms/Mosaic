#!/bin/bash

# Скрипт для предварительной загрузки Docker образов на GitLab Runner
# Запустите этот скрипт на сервере где установлен GitLab Runner

echo "🚀 Подготавливаем Docker образы для CI/CD..."

# Базовые образы для build процессов
echo "📥 Загружаем базовые образы..."
docker pull golang:1.24-alpine
docker pull node:20-alpine
docker pull nginx:alpine
docker pull docker:latest
docker pull ubuntu:22.04

# Образы для services (база данных, кеш)
echo "📥 Загружаем образы для сервисов..."
docker pull postgres:17-alpine
docker pull redis:8

echo "✅ Все образы загружены! Список:"
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

echo ""
echo "🔧 Настройка завершена! Теперь GitLab CI не будет обращаться к Docker Hub."
