const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  // Единая конфигурация прокси для CRA дев-сервера
  // По умолчанию шлём на backend контейнер в Docker-сети
  // Если фронт запущен локально, укажите BACKEND_URL=http://localhost:8080
  // В development используем переменную из Docker Compose, в production - http://backend:8080
  const target = process.env.BACKEND_URL || 'http://backend:8080';

  // REST API без переписывания пути (бекенд ожидает префикс /api)
  app.use(
    '/api',
    createProxyMiddleware({
      target,
      changeOrigin: true,
      logLevel: 'warn',
    })
  );

  // WebSocket чата: /ws/chat -> backend /api/ws/chat
  app.use(
    '/ws',
    createProxyMiddleware({
      target,
      changeOrigin: true,
      ws: true,
      logLevel: 'warn',
      pathRewrite: {
        '^/ws': '/api/ws',
      },
    })
  );
};


