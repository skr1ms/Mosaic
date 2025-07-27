import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

const ResetPassword: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  React.useEffect(() => {
    if (!token) {
      setError('Токен сброса пароля не найден. Пожалуйста, запросите новую ссылку.');
    }
  }, [token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!token) {
      setError('Токен сброса пароля не найден');
      return;
    }

    if (newPassword !== confirmPassword) {
      setError('Пароли не совпадают');
      return;
    }

    if (newPassword.length < 8) {
      setError('Пароль должен содержать минимум 8 символов');
      return;
    }

    setIsLoading(true);
    setError('');
    setMessage('');

    try {
      const response = await fetch('/api/partner/reset-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token,
          new_password: newPassword,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setMessage(data.message);
        // Перенаправляем на страницу входа через 3 секунды
        setTimeout(() => {
          navigate('/partner/login');
        }, 3000);
      } else {
        setError(data.error || 'Произошла ошибка при сбросе пароля');
      }
    } catch (err) {
      setError('Произошла ошибка при сбросе пароля');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-lg shadow-md p-6">
        <div className="space-y-4">
          <button
            onClick={() => navigate('/partner/login')}
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center space-x-1"
          >
            <span>←</span>
            <span>Вернуться к входу</span>
          </button>
          
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900">Сброс пароля</h2>
            <p className="mt-2 text-sm text-gray-600">
              Введите новый пароль для вашего аккаунта
            </p>
          </div>
        </div>

        {!token ? (
          <div className="mt-6 bg-red-50 border border-red-200 rounded-md p-4">
            <p className="text-sm text-red-600">
              Недействительная ссылка для сброса пароля. 
              <button
                onClick={() => navigate('/partner/forgot-password')}
                className="ml-1 underline hover:no-underline"
              >
                Запросить новую ссылку
              </button>
            </p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="mt-6 space-y-4">
            <div>
              <label htmlFor="newPassword" className="block text-sm font-medium text-gray-700 mb-1">
                Новый пароль
              </label>
              <input
                id="newPassword"
                type="password"
                placeholder="Введите новый пароль"
                value={newPassword}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
                minLength={8}
              />
              <p className="text-xs text-gray-500 mt-1">
                Минимум 8 символов
              </p>
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-1">
                Подтвердите пароль
              </label>
              <input
                id="confirmPassword"
                type="password"
                placeholder="Подтвердите новый пароль"
                value={confirmPassword}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setConfirmPassword(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
                minLength={8}
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-md p-3">
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            {message && (
              <div className="bg-green-50 border border-green-200 rounded-md p-3">
                <p className="text-sm text-green-600">{message}</p>
                <p className="text-xs text-green-500 mt-1">
                  Перенаправление на страницу входа...
                </p>
              </div>
            )}

            <button
              type="submit"
              className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
              disabled={isLoading || !token}
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                  <span>Обновление пароля...</span>
                </>
              ) : (
                <span>Сбросить пароль</span>
              )}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default ResetPassword;
