import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import api from '../api/api';
import eyeIcon from './eye_icon.png';

const Login = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    role: 'admin' // по умолчанию админ
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showTestData, setShowTestData] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  // Проверяем, есть ли уже токен при загрузке страницы
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token && token.startsWith('Bearer ')) {
      // Если есть токен, перенаправляем на dashboard
      navigate('/dashboard', { replace: true });
    }
  }, [navigate]);

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      console.log('Отправляем запрос:', {
        url: `/login/${formData.role}`,
        data: {
          login: formData.email,
          password: formData.password
        }
      });

      const response = await api.post(`/login/${formData.role}`, {
        login: formData.email, // API ожидает поле 'login', а не 'email'
        password: formData.password
      });

      console.log('Ответ от сервера:', response.data);

      const { access_token, user } = response.data;
      
      // Сохраняем токен и информацию о пользователе
      localStorage.setItem('token', `Bearer ${access_token}`);
      localStorage.setItem('userRole', user.role);
      localStorage.setItem('userId', user.id);
      localStorage.setItem('userEmail', user.email || user.login);
      localStorage.setItem('userName', user.name || user.login);

      // Перенаправляем на dashboard
      navigate('/dashboard', { replace: true });
    } catch (error) {
      console.error('Ошибка входа:', error);
      console.error('Детали ошибки:', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        message: error.message
      });
      
      let errorMessage = 'Ошибка входа. Проверьте логин и пароль.';
      
      if (error.response?.status === 0) {
        errorMessage = 'Не удается подключиться к серверу. Убедитесь, что backend запущен на http://localhost:8080';
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      } else if (error.response?.data?.message) {
        errorMessage = error.response.data.message;
      } else if (error.message) {
        errorMessage = error.message;
      }
      
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const fillTestData = (role) => {
    if (role === 'admin') {
      setFormData({
        email: 'admin',
        password: 'admin123',
        role: 'admin'
      });
    } else if (role === 'partner') {
      setFormData({
        email: 'quicktest',
        password: 'Partner123!',
        role: 'partner'
      });
    }
  };

  return (
    <div className="app-container app-theme-white">
      <div className="d-flex justify-content-center align-items-center min-vh-100 bg-light">
        <div className="card shadow-lg" style={{ width: '400px', maxWidth: '90vw' }}>
          <div className="card-header text-center bg-primary text-white">
            <h4 className="mb-0">
              <i className="pe-7s-user mr-2"></i>
              Вход в систему
            </h4>
          </div>
          <div className="card-body p-4">
            {error && (
              <div className="alert alert-danger" role="alert">
                <i className="pe-7s-attention mr-2"></i>
                {error}
              </div>
            )}
            
            <form onSubmit={handleSubmit}>
              <div className="mb-3">
                <label htmlFor="role" className="form-label">
                  <i className="pe-7s-id mr-2"></i>
                  Роль пользователя
                </label>
                <select
                  className="form-select"
                  id="role"
                  name="role"
                  value={formData.role}
                  onChange={handleChange}
                  required
                >
                  <option value="admin">Администратор</option>
                  <option value="partner">Партнер</option>
                </select>
              </div>

              <div className="mb-3">
                <label htmlFor="email" className="form-label d-flex align-items-center">
                  <i className="pe-7s-mail mr-2" style={{ lineHeight: '1', verticalAlign: 'middle', display: 'inline-flex', alignItems: 'center', marginTop: '2px' }}></i>
                  Email/Логин
                </label>
                <input
                  type="text"
                  className="form-control"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  placeholder="Введите ваш email или логин"
                  required
                />
              </div>

              <div className="mb-4">
                <label htmlFor="password" className="form-label">
                  <i className="pe-7s-lock mr-2"></i>
                  Пароль
                </label>
                <div className="input-group">
                  <input
                    type={showPassword ? "text" : "password"}
                    className="form-control border-end-0"
                    id="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    placeholder="Введите ваш пароль"
                    required
                  />
                  <span 
                    className="input-group-text bg-transparent border-start-0" 
                    style={{ cursor: 'pointer' }}
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    <img 
                      src={eyeIcon} 
                      alt={showPassword ? "Скрыть пароль" : "Показать пароль"}
                      style={{ 
                        width: '16px', 
                        height: '16px',
                        filter: showPassword ? 'brightness(0.7)' : 'brightness(1)',
                        transition: 'filter 0.2s ease'
                      }}
                    />
                  </span>
                </div>
              </div>

              <button
                type="submit"
                className="btn btn-primary w-100 btn-lg"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                    Вход в систему...
                  </>
                ) : (
                  <>
                    <i className="pe-7s-right-arrow mr-2"></i>
                    Войти
                  </>
                )}
              </button>
            </form>

            {/* Тестовые данные */}
            <div className="mt-4">
              <button
                type="button"
                className="btn btn-outline-info btn-sm w-100"
                onClick={() => setShowTestData(!showTestData)}
              >
                <i className="pe-7s-info mr-2"></i>
                {showTestData ? 'Скрыть' : 'Показать'} тестовые данные
              </button>
              
              {showTestData && (
                <div className="mt-3 p-3 bg-light rounded">
                  <h6 className="text-muted mb-2">
                    <i className="pe-7s-info mr-2"></i>
                    Тестовые учетные данные:
                  </h6>
                  
                  <div className="mb-2">
                    <strong>Администратор:</strong>
                    <div className="small text-muted">
                      Логин: <code>admin</code><br/>
                      Пароль: <code>admin123</code>
                    </div>
                    <button
                      type="button"
                      className="btn btn-outline-primary btn-sm mt-1"
                      onClick={() => fillTestData('admin')}
                    >
                      <i className="pe-7s-check mr-1"></i>
                      Заполнить
                    </button>
                  </div>
                  
                  <div className="mb-2">
                    <strong>Партнер:</strong>
                    <div className="small text-muted">
                      Логин: <code>quicktest</code><br/>
                      Пароль: <code>Partner123!</code>
                    </div>
                    <button
                      type="button"
                      className="btn btn-outline-primary btn-sm mt-1"
                      onClick={() => fillTestData('partner')}
                    >
                      <i className="pe-7s-check mr-1"></i>
                      Заполнить
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
