import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api/api';
import Chat from '../components/Chat';

interface PartnerStats {
  totalCoupons: number;
  activeCoupons: number;
  redeemedCoupons: number;
  totalRevenue: number;
  conversionRate: number;
}

interface PartnerInfo {
  id: string;
  name: string;
  email: string;
  status: string;
  registrationDate: string;
}

const PartnerDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [partnerStats, setPartnerStats] = useState<PartnerStats | null>(null);
  const [partnerInfo, setPartnerInfo] = useState<PartnerInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState<'admin' | 'partner'>('partner');
  const [userId, setUserId] = useState<string>('');

  useEffect(() => {
    // Проверяем авторизацию
    const token = localStorage.getItem('access_token');
    const role = localStorage.getItem('user_role') as 'admin' | 'partner';
    const id = localStorage.getItem('user_id');

    if (!token || !role || !id) {
      navigate('/login');
      return;
    }

    setUserRole(role);
    setUserId(id);

    // Если это админ, перенаправляем на админскую панель
    if (role === 'admin') {
      navigate('/admin/dashboard');
      return;
    }

    const fetchData = async () => {
      try {
        const [statsRes, infoRes] = await Promise.all([
          api.get('/partner/dashboard'),
          api.get('/partner/profile')
        ]);
        
        setPartnerStats(statsRes.data);
        setPartnerInfo(infoRes.data);
      } catch (error) {
        console.error('Ошибка загрузки данных:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [navigate]);

  const handleLogout = () => {
    localStorage.clear();
    navigate('/login');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-xl">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <h1 className="text-3xl font-bold text-gray-900">
              Панель партнера
            </h1>
            <div className="flex items-center space-x-4">
              <span className="text-gray-600">
                Партнер: {localStorage.getItem('user_login')}
              </span>
              <button
                onClick={handleLogout}
                className="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg"
              >
                Выйти
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Основная статистика */}
        {partnerStats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
            <StatCard
              title="Всего купонов"
              value={partnerStats.totalCoupons.toString()}
              icon="🎫"
              color="bg-blue-500"
            />
            <StatCard
              title="Активные"
              value={partnerStats.activeCoupons.toString()}
              icon="✅"
              color="bg-green-500"
            />
            <StatCard
              title="Использовано"
              value={partnerStats.redeemedCoupons.toString()}
              icon="🔥"
              color="bg-yellow-500"
            />
            <StatCard
              title="Доход"
              value={`₽${partnerStats.totalRevenue.toLocaleString()}`}
              icon="💰"
              color="bg-purple-500"
            />
            <StatCard
              title="Конверсия"
              value={`${partnerStats.conversionRate.toFixed(1)}%`}
              icon="📈"
              color="bg-indigo-500"
            />
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Информация о партнере */}
          {partnerInfo && (
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Информация о партнере</h2>
              <div className="space-y-4">
                <div className="flex justify-between">
                  <span>Имя:</span>
                  <span className="font-semibold">{partnerInfo.name}</span>
                </div>
                <div className="flex justify-between">
                  <span>Email:</span>
                  <span className="font-semibold">{partnerInfo.email}</span>
                </div>
                <div className="flex justify-between">
                  <span>Статус:</span>
                  <span className={`font-semibold ${
                    partnerInfo.status === 'active' ? 'text-green-600' : 'text-red-600'
                  }`}>
                    {partnerInfo.status === 'active' ? '🟢 Активный' : '🔴 Неактивный'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Дата регистрации:</span>
                  <span className="font-semibold">
                    {new Date(partnerInfo.registrationDate).toLocaleDateString('ru-RU')}
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Быстрые действия */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Быстрые действия</h2>
            <div className="space-y-3">
              <button className="w-full text-left p-3 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors">
                <div className="flex items-center">
                  <span className="text-2xl mr-3">🎫</span>
                  <div>
                    <div className="font-medium">Создать купон</div>
                    <div className="text-sm text-gray-600">Добавить новый купон</div>
                  </div>
                </div>
              </button>
              
              <button className="w-full text-left p-3 bg-green-50 hover:bg-green-100 rounded-lg transition-colors">
                <div className="flex items-center">
                  <span className="text-2xl mr-3">📊</span>
                  <div>
                    <div className="font-medium">Аналитика</div>
                    <div className="text-sm text-gray-600">Просмотр статистики</div>
                  </div>
                </div>
              </button>
              
              <button className="w-full text-left p-3 bg-purple-50 hover:bg-purple-100 rounded-lg transition-colors">
                <div className="flex items-center">
                  <span className="text-2xl mr-3">⚙️</span>
                  <div>
                    <div className="font-medium">Настройки</div>
                    <div className="text-sm text-gray-600">Управление профилем</div>
                  </div>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Чат */}
      <Chat userRole={userRole} userId={userId} />
    </div>
  );
};

interface StatCardProps {
  title: string;
  value: string;
  icon: string;
  color: string;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, icon, color }) => (
  <div className="bg-white rounded-lg shadow p-6">
    <div className="flex items-center">
      <div className={`${color} rounded-md p-3 text-white text-2xl mr-4`}>
        {icon}
      </div>
      <div>
        <p className="text-sm font-medium text-gray-600">{title}</p>
        <p className="text-2xl font-semibold text-gray-900">{value}</p>
      </div>
    </div>
  </div>
);

export default PartnerDashboard; 