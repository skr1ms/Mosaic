import React, { useState, useEffect } from 'react';
import api from '../api/api';

interface DashboardStats {
  totalPartners: number;
  totalCoupons: number;
  activeCoupons: number;
  redeemedCoupons: number;
  conversionRate: number;
}

interface SystemStats {
  uptime: string;
  activeUsers: number;
  todayActivations: number;
  systemHealth: string;
}

interface PartnerAnalytics {
  topPartners: Array<{
    name: string;
    coupons: number;
    revenue: number;
  }>;
  monthlyGrowth: number;
  churnRate: number;
}

const AdminDashboard: React.FC = () => {
  const [dashboardStats, setDashboardStats] = useState<DashboardStats | null>(null);
  const [systemStats, setSystemStats] = useState<SystemStats | null>(null);
  const [analytics, setAnalytics] = useState<PartnerAnalytics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [dashboardRes, systemRes, analyticsRes] = await Promise.all([
          api.get('/admin/dashboard'),
          api.get('/admin/system-stats'),
          api.get('/admin/analytics')
        ]);
        
        setDashboardStats(dashboardRes.data);
        setSystemStats(systemRes.data);
        setAnalytics(analyticsRes.data);
      } catch (error) {
        console.error('Ошибка загрузки данных:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-xl">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="py-6">
            <h1 className="text-3xl font-bold text-gray-900">
              Панель администратора
            </h1>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Основная статистика */}
        {dashboardStats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6 mb-8">
            <StatCard
              title="Партнеры"
              value={dashboardStats.totalPartners.toString()}
              icon="👥"
              color="bg-blue-500"
            />
            <StatCard
              title="Всего купонов"
              value={dashboardStats.totalCoupons.toString()}
              icon="🎫"
              color="bg-green-500"
            />
            <StatCard
              title="Активные"
              value={dashboardStats.activeCoupons.toString()}
              icon="✅"
              color="bg-yellow-500"
            />
            <StatCard
              title="Использовано"
              value={dashboardStats.redeemedCoupons.toString()}
              icon="🔥"
              color="bg-red-500"
            />
            <StatCard
              title="Конверсия"
              value={`${dashboardStats.conversionRate.toFixed(1)}%`}
              icon="📈"
              color="bg-purple-500"
            />
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Системная информация */}
          {systemStats && (
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Состояние системы</h2>
              <div className="space-y-4">
                <div className="flex justify-between">
                  <span>Время работы:</span>
                  <span className="font-semibold">{systemStats.uptime}</span>
                </div>
                <div className="flex justify-between">
                  <span>Активные пользователи:</span>
                  <span className="font-semibold">{systemStats.activeUsers}</span>
                </div>
                <div className="flex justify-between">
                  <span>Активаций сегодня:</span>
                  <span className="font-semibold">{systemStats.todayActivations}</span>
                </div>
                <div className="flex justify-between">
                  <span>Состояние:</span>
                  <span className={`font-semibold ${
                    systemStats.systemHealth === 'healthy' ? 'text-green-600' : 'text-red-600'
                  }`}>
                    {systemStats.systemHealth === 'healthy' ? '🟢 Здоровая' : '🔴 Проблемы'}
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Аналитика партнеров */}
          {analytics && (
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-semibold mb-4">Аналитика</h2>
              
              <div className="mb-6">
                <h3 className="text-lg font-medium mb-3">Топ партнеры</h3>
                <div className="space-y-2">
                  {analytics.topPartners.map((partner, index) => (
                    <div key={index} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span>{partner.name}</span>
                      <div className="text-sm text-gray-600">
                        {partner.coupons} купонов | ₽{partner.revenue}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-3 bg-green-50 rounded">
                  <div className="text-2xl font-bold text-green-600">
                    +{analytics.monthlyGrowth}%
                  </div>
                  <div className="text-sm text-gray-600">Рост за месяц</div>
                </div>
                <div className="text-center p-3 bg-red-50 rounded">
                  <div className="text-2xl font-bold text-red-600">
                    {analytics.churnRate}%
                  </div>
                  <div className="text-sm text-gray-600">Отток</div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
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

export default AdminDashboard; 