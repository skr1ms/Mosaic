import React, { useState, useEffect } from 'react';
import api from '../../../api/api';

const SystemStatus = () => {
  const [systemStatus, setSystemStatus] = useState({
    uptime: '',
    activeUsers: 0,
    todayActivations: 0,
    systemHealth: 'healthy',
    frontendStatus: 'online',
    backendStatus: 'online',
    databaseStatus: 'online'
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSystemStatus = async () => {
      try {
        const response = await api.get('/admin/system-status');
        setSystemStatus(response.data);
      } catch (error) {
        console.error('Ошибка загрузки состояния системы:', error);
        setSystemStatus(prev => ({
          ...prev,
          systemHealth: 'error',
          backendStatus: 'offline'
        }));
      } finally {
        setLoading(false);
      }
    };

    fetchSystemStatus();
    // Обновляем статус каждые 30 секунд
    const interval = setInterval(fetchSystemStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="app-main__inner">
        <div className="row">
          <div className="col-12">
            <div className="card mb-3">
              <div className="card-body">
                <div className="text-center">
                  <div className="spinner-border" role="status">
                    <span className="sr-only">Загрузка...</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'online':
      case 'healthy':
        return 'text-success';
      case 'warning':
        return 'text-warning';
      case 'offline':
      case 'error':
        return 'text-danger';
      default:
        return 'text-muted';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'online':
      case 'healthy':
        return '🟢';
      case 'warning':
        return '🟡';
      case 'offline':
      case 'error':
        return '🔴';
      default:
        return '⚪';
    }
  };

  return (
    <div className="app-main__inner">
      <div className="row">
        <div className="col-12">
          <div className="card mb-3">
            <div className="card-header">
              <h5 className="card-title mb-0">Состояние системы</h5>
            </div>
            <div className="card-body">
              <div className="row">
                <div className="col-md-6">
                  <h6 className="mb-3">Основные показатели</h6>
                  <div className="list-group list-group-flush">
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Время работы системы:</span>
                      <span className="font-weight-bold">{systemStatus.uptime}</span>
                    </div>
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Активные пользователи:</span>
                      <span className="font-weight-bold">{systemStatus.activeUsers}</span>
                    </div>
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Активаций сегодня:</span>
                      <span className="font-weight-bold">{systemStatus.todayActivations}</span>
                    </div>
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Общее состояние:</span>
                      <span className={`font-weight-bold ${getStatusColor(systemStatus.systemHealth)}`}>
                        {getStatusIcon(systemStatus.systemHealth)} {systemStatus.systemHealth === 'healthy' ? 'Здоровая' : 'Проблемы'}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="col-md-6">
                  <h6 className="mb-3">Статус сервисов</h6>
                  <div className="list-group list-group-flush">
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Фронтенд (localhost:3000):</span>
                      <div className="d-flex align-items-center">
                        <span className={`font-weight-bold ${getStatusColor(systemStatus.frontendStatus)}`}>
                          {getStatusIcon(systemStatus.frontendStatus)} {systemStatus.frontendStatus === 'online' ? 'Онлайн' : 'Офлайн'}
                        </span>
                        <a 
                          href="http://localhost:3000" 
                          target="_blank" 
                          rel="noopener noreferrer"
                          className="btn btn-sm btn-primary ml-2"
                        >
                          <i className="pe-7s-external-link mr-1"></i>
                          Открыть фронтенд
                        </a>
                      </div>
                    </div>
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>Бэкенд (localhost:8080):</span>
                      <span className={`font-weight-bold ${getStatusColor(systemStatus.backendStatus)}`}>
                        {getStatusIcon(systemStatus.backendStatus)} {systemStatus.backendStatus === 'online' ? 'Онлайн' : 'Офлайн'}
                      </span>
                    </div>
                    <div className="list-group-item d-flex justify-content-between align-items-center">
                      <span>База данных:</span>
                      <span className={`font-weight-bold ${getStatusColor(systemStatus.databaseStatus)}`}>
                        {getStatusIcon(systemStatus.databaseStatus)} {systemStatus.databaseStatus === 'online' ? 'Онлайн' : 'Офлайн'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SystemStatus;
