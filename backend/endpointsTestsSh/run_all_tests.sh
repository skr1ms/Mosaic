#!/bin/bash

# Главный скрипт для запуска всех тестов эндпоинтов
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:3000/api"

echo "=== Запуск всех тестов эндпоинтов Mosaic API ==="
echo "Базовый URL: $BASE_URL"
echo "Директория скриптов: $SCRIPT_DIR"
echo ""

# Проверяем доступность API
echo "Проверка доступности API..."
if curl -s --max-time 5 "$BASE_URL" > /dev/null; then
    echo "✅ API доступен"
else
    echo "❌ API недоступен по адресу $BASE_URL"
    echo "Убедитесь, что контейнер backend запущен"
    exit 1
fi
echo ""

# Делаем все скрипты исполняемыми
chmod +x "$SCRIPT_DIR"/*.sh

# Список всех тестовых скриптов
TESTS=(
    "test_auth_endpoints.sh:Auth (авторизация)"
    "test_public_endpoints.sh:Public API (публичные эндпоинты)"
    "test_admin_endpoints.sh:Admin (административные функции)"
    "test_partner_endpoints.sh:Partner (партнерские функции)"
    "test_coupon_endpoints.sh:Coupon (купоны)"
    "test_payment_endpoints.sh:Payment (платежи)"
    "test_image_endpoints.sh:Image Processing (обработка изображений)"
    "test_stats_endpoints.sh:Stats (статистика)"
)

# Запускаем тесты
for test_info in "${TESTS[@]}"; do
    IFS=':' read -r script_name description <<< "$test_info"
    
    echo "🔄 Запуск тестов: $description"
    echo "📄 Скрипт: $script_name"
    echo "⏰ Время запуска: $(date)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [ -f "$SCRIPT_DIR/$script_name" ]; then
        cd "$SCRIPT_DIR"
        ./"$script_name"
        echo ""
        echo "✅ Тест '$description' завершен"
    else
        echo "❌ Скрипт $script_name не найден"
    fi
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Небольшая пауза между тестами
    sleep 1
done

echo "🎉 Все тесты завершены!"
echo "⏰ Время завершения: $(date)"
echo ""
echo "📋 Краткая сводка:"
echo "   • Auth: тестирование авторизации и токенов"
echo "   • Public API: публичные эндпоинты без авторизации"
echo "   • Admin: административные функции (требует admin токен)"
echo "   • Partner: функции партнеров (требует создание партнера)"
echo "   • Coupon: работа с купонами"
echo "   • Payment: платежная система"
echo "   • Image Processing: обработка изображений"
echo "   • Stats: статистика системы"
echo ""
echo "💡 Для запуска отдельного теста используйте:"
echo "   cd $SCRIPT_DIR && ./имя_скрипта.sh"
