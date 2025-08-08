import axios from 'axios';

// Создаем экземпляр axios с базовой конфигурацией
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Логируем базовый URL для отладки
console.log('API Base URL:', process.env.REACT_APP_API_URL || 'http://localhost:8080/api');

// Интерцептор для добавления токена авторизации
api.interceptors.request.use(
  (config) => {
    console.log('API Request:', config.method?.toUpperCase(), config.url, config.data);
    const token = localStorage.getItem('token');
    if (token && token.startsWith('Bearer ')) {
      config.headers.Authorization = token;
    } else if (token) {
      // Если токен есть, но не в формате Bearer, добавляем префикс
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    console.error('API Request Error:', error);
    return Promise.reject(error);
  }
);

// Интерцептор для обработки ответов
api.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.status, response.data);
    return response;
  },
  (error) => {
    console.error('API Response Error:', error.response?.status, error.response?.data);
    if (error.response?.status === 401) {
      // Если токен истек, очищаем localStorage и перенаправляем на главную страницу
      localStorage.removeItem('token');
      localStorage.removeItem('userRole');
      localStorage.removeItem('userId');
      localStorage.removeItem('userEmail');
      localStorage.removeItem('userName');
      
      // Проверяем, что мы не на главной странице
      if (window.location.pathname !== '/' && window.location.pathname !== '/login') {
        window.location.href = '/';
      }
    }
    return Promise.reject(error);
  }
);

export default api;
