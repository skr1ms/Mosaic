FROM node:18-alpine AS builder

WORKDIR /app

# Копируем package файлы
COPY package*.json ./

# Устанавливаем зависимости
RUN npm ci --only=production

# Копируем исходный код
COPY . .

# Собираем приложение
RUN npm run build

# Продакшен образ
FROM nginx:alpine

# Создаем директории для логов
RUN mkdir -p /var/log/nginx

# Копируем собранное приложение
COPY --from=builder /app/dist /usr/share/nginx/html

# Копируем nginx конфиг
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Создаем пользователя для nginx
RUN addgroup -g 1001 -S nginx && \
    adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx

# Устанавливаем правильные права
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/log/nginx

# Переключаемся на пользователя nginx
USER nginx

# Открываем только порт 80
EXPOSE 80

# Запускаем nginx
CMD ["nginx", "-g", "daemon off;"]
