import axios from 'axios';



const ENV_API_BASE = process.env.REACT_APP_API_BASE_URL || '/api';


const API_BASE_URL = ENV_API_BASE.replace(/\/$/, '');

// Функция для декодирования JWT токена
function decodeJWT(token) {
  try {
    const tokenWithoutBearer = token.replace('Bearer ', '');
    const base64Url = tokenWithoutBearer.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

// Функция для проверки истечения токена
function isTokenExpired(token) {
  if (!token) return true;
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) return true;
  const currentTime = Math.floor(Date.now() / 1000);
  return decoded.exp < currentTime;
}

// Функция для очистки аутентификации
function clearAuth() {
  localStorage.removeItem('token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('userRole');
  localStorage.removeItem('userId');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('userName');
  localStorage.removeItem('auth_session_id');
  sessionStorage.removeItem('auth_session_id');
  window.dispatchEvent(new Event('auth-updated'));
}


function checkTokenOnInit() {
  const token = localStorage.getItem('token');
  const refreshToken = localStorage.getItem('refresh_token');
  
  
  if (!token) {
    return false;
  }
  
  
  const sessionId = sessionStorage.getItem('auth_session_id');
  const storedSessionId = localStorage.getItem('auth_session_id');
  
  
  if (!sessionId || sessionId !== storedSessionId) {
    console.warn('New browser session detected, clearing auth and requiring login');
    clearAuth();
    return false;
  }
  
  
  if (token && isTokenExpired(token)) {
    console.warn('Access token expired on init');
    
    
    if (!refreshToken || isTokenExpired(refreshToken)) {
      console.warn('Refresh token also expired, clearing auth');
      clearAuth();
      return false;
    }
    
    
    console.warn('Access token expired, requiring re-login for security');
    clearAuth();
    return false;
  }
  
  
  const decoded = decodeJWT(token);
  if (decoded && decoded.iat) {
    const tokenAge = Math.floor(Date.now() / 1000) - decoded.iat;
    
    if (tokenAge > 1800) { 
      console.warn('Token is too old, requiring fresh login');
      clearAuth();
      return false;
    }
  }
  
  return true;
}

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
});


console.log('API Base URL:', API_BASE_URL);
console.log('Environment:', process.env.REACT_APP_ENVIRONMENT);


api.interceptors.request.use(
  async (config) => {
    console.log('API Request:', config.method?.toUpperCase(), config.url, config.data);
    
    const token = localStorage.getItem('token');
    const refreshToken = localStorage.getItem('refresh_token');
    
    
    if (token && isTokenExpired(token)) {
      console.warn('Access token expired, attempting refresh');
      
      if (refreshToken && !isTokenExpired(refreshToken)) {
        try {
          const response = await axios.post(`${API_BASE_URL}/auth/refresh`, { refresh_token: refreshToken });
          const { access_token, refresh_token: newRefreshToken } = response.data || {};
          
          if (access_token) {
            localStorage.setItem('token', `Bearer ${access_token}`);
            config.headers.Authorization = `Bearer ${access_token}`;
          }
          if (newRefreshToken) {
            localStorage.setItem('refresh_token', newRefreshToken);
          }
        } catch (error) {
          console.error('Token refresh failed:', error);
          clearAuth();
          
          if (window.location.pathname !== '/login') {
            window.location.href = '/login';
          }
          return Promise.reject(error);
        }
      } else {
        console.warn('Refresh token expired, clearing auth');
        clearAuth();
        
        if (window.location.pathname !== '/login') {
          window.location.href = '/login';
        }
        return Promise.reject(new Error('Authentication expired'));
      }
    } else if (token) {
      if (token.startsWith('Bearer ')) {
        config.headers.Authorization = token;
      } else {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    
    return config;
  },
  (error) => {
    console.error('API Request Error:', error);
    return Promise.reject(error);
  }
);


let refreshPromise = null;

api.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.status, response.data);
    return response;
  },
  async (error) => {
    if (!error.response) {
      console.error('API Network Error:', error.message || 'Network error');
      return Promise.reject(error);
    }
    console.error('API Response Error:', error.response.status, error.response.data);

    const originalRequest = error.config;
    const status = error.response?.status;
    const isLoginRequest = originalRequest?.url?.includes('/login/');

    if (status === 401 && !isLoginRequest) {
      const storedRefresh = localStorage.getItem('refresh_token');
      const role = localStorage.getItem('userRole');

      if (storedRefresh && role) {
        try {
          if (!refreshPromise) {
            refreshPromise = api.post(`/auth/refresh`, { refresh_token: storedRefresh })
              .then((res) => {
                const { access_token, refresh_token } = res.data || {};
                if (access_token) {
                  localStorage.setItem('token', `Bearer ${access_token}`);
                }
                if (refresh_token) {
                  localStorage.setItem('refresh_token', refresh_token);
                }
                return access_token;
              })
              .finally(() => {
                refreshPromise = null;
              });
          }

          const newAccessToken = await refreshPromise;
          if (newAccessToken) {
            originalRequest.headers = originalRequest.headers || {};
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
            return api(originalRequest);
          }
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError.response?.data || refreshError.message);
        }
      }

      
      clearAuth();
      if (window.location.pathname !== '/' && window.location.pathname !== '/login') {
        window.location.href = '/';
      }
    }
    return Promise.reject(error);
  }
);

export { checkTokenOnInit, clearAuth };
export default api;
