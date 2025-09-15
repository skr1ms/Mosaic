import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../../../api/api';

const SystemStatus = () => {
  const { t } = useTranslation();
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
        const response = await api.get('/admin/statistics/system');
        const data = response.data || {};
        
        setSystemStatus(prev => ({
          uptime: `${data.system_health?.uptime_hours ?? 0}h`,
          activeUsers: data.quick_stats?.active_users_online ?? prev.activeUsers,
          todayActivations: data.quick_stats?.coupons_today ?? prev.todayActivations,
          systemHealth: data.system_health?.status || 'healthy',
          frontendStatus: prev.frontendStatus,
          backendStatus: 'online',
          databaseStatus: data.database?.status === 'connected' ? 'online' : 'offline',
        }));
      } catch (error) {
        console.error('Error loading system status:', error);
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
                    <span className="sr-only">{t('common.loading')}</span>
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
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-server icon-gradient bg-mean-fruit" />
            </div>
            <div>
              {t('system.title')}
              <div className="page-title-subheading">
                {t('system.system_status')}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="card mb-3">
            <div className="card-header">
              <h5 className="card-title">{t('system.main_indicators')}</h5>
            </div>
            <div className="card-body">
              <div className="row">
                <div className="col-md-3">
                  <div className="widget-chart-box">
                    <div className="widget-chart-content">
                      <div className="widget-numbers text-primary">
                        <b>{systemStatus.uptime}</b>
                      </div>
                      <div className="widget-subheading">
                        {t('system.system_uptime')}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="col-md-3">
                  <div className="widget-chart-box">
                    <div className="widget-chart-content">
                      <div className="widget-numbers text-success">
                        <b>{systemStatus.activeUsers}</b>
                      </div>
                      <div className="widget-subheading">
                        {t('system.active_users')}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="col-md-3">
                  <div className="widget-chart-box">
                    <div className="widget-chart-content">
                      <div className="widget-numbers text-warning">
                        <b>{systemStatus.todayActivations}</b>
                      </div>
                      <div className="widget-subheading">
                        {t('system.today_activations')}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="col-md-3">
                  <div className="widget-chart-box">
                    <div className="widget-chart-content">
                      <div className="widget-numbers text-info">
                        <b>{getStatusIcon(systemStatus.systemHealth)}</b>
                      </div>
                      <div className="widget-subheading">
                        {t('system.system_health')}
                      </div>
                      <div className={`widget-description ${getStatusColor(systemStatus.systemHealth)}`}>
                        {t(`system.${systemStatus.systemHealth}`)}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="card mb-3">
            <div className="card-header">
              <h5 className="card-title">{t('system.service_status')}</h5>
            </div>
            <div className="card-body">
              <div className="row">
                <div className="col-md-4">
                  <div className="d-flex align-items-center mb-3">
                    <span className="mr-3">{getStatusIcon(systemStatus.frontendStatus)}</span>
                    <div>
                      <h6 className="mb-0">{t('system.frontend')}</h6>
                      <small className={`${getStatusColor(systemStatus.frontendStatus)}`}>
                        {t(`system.${systemStatus.frontendStatus}`)}
                      </small>
                    </div>
                  </div>
                </div>
                <div className="col-md-4">
                  <div className="d-flex align-items-center mb-3">
                    <span className="mr-3">{getStatusIcon(systemStatus.backendStatus)}</span>
                    <div>
                      <h6 className="mb-0">{t('system.backend')}</h6>
                      <small className={`${getStatusColor(systemStatus.backendStatus)}`}>
                        {t(`system.${systemStatus.backendStatus}`)}
                      </small>
                    </div>
                  </div>
                </div>
                <div className="col-md-4">
                  <div className="d-flex align-items-center mb-3">
                    <span className="mr-3">{getStatusIcon(systemStatus.databaseStatus)}</span>
                    <div>
                      <h6 className="mb-0">{t('system.database')}</h6>
                      <small className={`${getStatusColor(systemStatus.databaseStatus)}`}>
                        {t(`system.${systemStatus.databaseStatus}`)}
                      </small>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="row">
        <div className="col-12">
          <div className="card mb-3">
            <div className="card-header">
              <h5 className="card-title">{t('system.monitoring')}</h5>
            </div>
            <div className="card-body">
              <div className="d-grid gap-2 d-md-block">
                <button 
                  className="btn btn-primary"
                  onClick={() => window.open('/grafana', '_blank')}
                >
                  <i className="pe-7s-graph1"></i> {t('system.grafana')}
                </button>
                <button 
                  className="btn btn-info"
                  onClick={() => window.open('/', '_blank')}
                >
                  <i className="pe-7s-home"></i> {t('system.open_frontend')}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SystemStatus;
