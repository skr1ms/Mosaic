import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
});

// Добавляем интерцептор для автоматического добавления токена
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Добавляем интерцептор для обработки ошибок авторизации
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Токен истек, пытаемся обновить
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const userRole = localStorage.getItem('user_role');
          const response = await axios.post(`http://localhost:8080/api/refresh/${userRole}`, {
            refresh_token: refreshToken
          });
          
          localStorage.setItem('access_token', response.data.access_token);
          localStorage.setItem('refresh_token', response.data.refresh_token);
          
          // Повторяем оригинальный запрос
          error.config.headers.Authorization = `Bearer ${response.data.access_token}`;
          return api.request(error.config);
        } catch (refreshError) {
          // Не удалось обновить токен, перенаправляем на логин
          localStorage.clear();
          window.location.href = '/login';
        }
      } else {
        // Нет refresh токена, перенаправляем на логин
        localStorage.clear();
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default api; 