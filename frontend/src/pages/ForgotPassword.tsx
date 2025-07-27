import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const ForgotPassword: React.FC = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [captchaToken, setCaptchaToken] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  // Простая captcha для демо
  const generateCaptcha = () => {
    const token = `captcha_${Math.random().toString(36).substring(2)}_${Date.now()}`;
    setCaptchaToken(token);
  };

  React.useEffect(() => {
    generateCaptcha();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    setMessage('');

    try {
      const response = await fetch('/api/partner/forgot-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          captcha_token: captchaToken,
        }),
      });

      const data = await response.json();

      if (response.ok) {
        setMessage(data.message);
      } else {
        setError(data.error || 'Произошла ошибка при отправке запроса');
      }
    } catch (err) {
      setError('Произошла ошибка при отправке запроса');
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
            <h2 className="text-2xl font-bold text-gray-900">Восстановление пароля</h2>
            <p className="mt-2 text-sm text-gray-600">
              Введите ваш email адрес и мы отправим вам ссылку для восстановления пароля
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              Email адрес
            </label>
            <div className="relative">
              <input
                id="email"
                type="email"
                placeholder="partner@example.com"
                value={email}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Проверка на робота
            </label>
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="captcha"
                className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500"
                required
              />
              <label htmlFor="captcha" className="text-sm text-gray-700">
                Я не робот
              </label>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              Токен captcha: {captchaToken.substring(0, 20)}...
            </p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          {message && (
            <div className="bg-green-50 border border-green-200 rounded-md p-3">
              <p className="text-sm text-green-600">{message}</p>
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
                <span>Отправка...</span>
              </>
            ) : (
              <span>Отправить ссылку для восстановления</span>
            )}
          </button>
        </form>
      </div>
    </div>
  );
};

export default ForgotPassword;
