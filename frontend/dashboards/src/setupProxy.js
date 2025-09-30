const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  
  
  
  
  const target = process.env.BACKEND_URL || 'http://backend:8080';

  
  app.use(
    '/api',
    createProxyMiddleware({
      target,
      changeOrigin: true,
      logLevel: 'warn',
    })
  );

  
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


