import React from 'react';
import { Link } from 'react-router-dom';

const Home = () => {
  return (
    <div className="min-h-screen bg-gray-100">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="py-6">
            <h1 className="text-3xl font-bold text-gray-900">
              Mosaic - Система управления купонами
            </h1>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          
          {/* Админская панель */}
          <Link
            to="/login?type=admin"
            className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
          >
            <div className="text-4xl mb-4">👨‍💼</div>
            <h2 className="text-xl font-semibold mb-2">Админ панель</h2>
            <p className="text-gray-600">
              Просмотр статистики, управление партнерами и купонами
            </p>
          </Link>

          {/* Панель партнера */}
          <Link
            to="/login?type=partner"
            className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
          >
            <div className="text-4xl mb-4">🤝</div>
            <h2 className="text-xl font-semibold mb-2">Панель партнера</h2>
            <p className="text-gray-600">
              Генерация купонов, статистика продаж
            </p>
          </Link>

          {/* Активация купонов */}
          <div className="bg-white rounded-lg shadow-md p-6 opacity-50">
            <div className="text-4xl mb-4">🎫</div>
            <h2 className="text-xl font-semibold mb-2">Активация</h2>
            <p className="text-gray-600">
              Активация купонов пользователями (в разработке)
            </p>
          </div>

        </div>

        {/* API информация */}
        <div className="mt-12 bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4">🔗 API Endpoints</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <h3 className="font-medium text-green-600 mb-2">Админ API:</h3>
              <ul className="space-y-1 text-gray-600">
                <li><code>GET /admin/dashboard</code></li>
                <li><code>GET /admin/system-stats</code></li>
                <li><code>GET /admin/analytics</code></li>
              </ul>
            </div>
            <div>
              <h3 className="font-medium text-blue-600 mb-2">Partner API:</h3>
              <ul className="space-y-1 text-gray-600">
                <li><code>GET /partner/coupons</code></li>
                <li><code>POST /partner/generate</code></li>
                <li><code>GET /partner/stats</code></li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home; 