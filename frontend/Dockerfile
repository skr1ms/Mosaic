FROM nginx:alpine

# Создаем директории для логов
RUN mkdir -p /var/log/nginx

# Копируем готовые HTML страницы из pages
COPY pages /usr/share/nginx/html

# Копируем nginx конфиг
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Устанавливаем правильные права для nginx
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/log/nginx

# Открываем только порт 80
EXPOSE 80

# Запускаем nginx
CMD ["nginx", "-g", "daemon off;"]
