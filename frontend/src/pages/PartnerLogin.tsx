import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';

const PartnerLogin: React.FC = () => {
  const navigate = useNavigate();
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await fetch('/api/partner/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          login,
          password,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        // Сохраняем токены
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        localStorage.setItem('partner', JSON.stringify(data.partner));
        
        // Перенаправляем на дашборд
        navigate('/partner/dashboard');
      } else {
        setError(data.error || 'Ошибка авторизации');
      }
    } catch (err) {
      setError('Произошла ошибка при входе в систему');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-lg shadow-md p-6">
        <div className="space-y-4">
          <button
            onClick={() => navigate('/')}
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center space-x-1"
          >
            <span>←</span>
            <span>На главную</span>
          </button>
          
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900">Вход в партнерскую панель</h2>
            <p className="mt-2 text-sm text-gray-600">
              Войдите в свой аккаунт для управления купонами
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <div>
            <label htmlFor="login" className="block text-sm font-medium text-gray-700 mb-1">
              Логин
            </label>
            <input
              id="login"
              type="text"
              placeholder="Введите логин"
              value={login}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setLogin(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">
              Пароль
            </label>
            <input
              id="password"
              type="password"
              placeholder="Введите пароль"
              value={password}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
            />
          </div>

          <div className="text-right">
            <Link
              to="/partner/forgot-password"
              className="text-sm text-blue-600 hover:text-blue-700 hover:underline"
            >
              Забыли пароль?
            </Link>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <button
            type="submit"
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                <span>Вход...</span>
              </>
            ) : (
              <span>Войти</span>
            )}
          </button>
        </form>

        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600">
            Нет аккаунта? Обратитесь к администратору для создания партнерского аккаунта.
          </p>
        </div>
      </div>
    </div>
  );
};

export default PartnerLogin;
