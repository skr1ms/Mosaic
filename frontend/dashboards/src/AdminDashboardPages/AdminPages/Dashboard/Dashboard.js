import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Badge, Table, Button
} from 'reactstrap';
import api from '../../../api/api';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';

const Dashboard = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [userRole, setUserRole] = useState('');
    const [data, setData] = useState({
        partners: { total: 0, active: 0 },
        coupons: { total: 0, used: 0, new: 0, purchased: 0 },
        image_processing: { total: 0, processing: 0, completed: 0, failed: 0 },
        recent_activations: []
    });

    // Инициализация userRole при монтировании компонента
    useEffect(() => {
        const storedRole = localStorage.getItem('userRole');
        if (storedRole) {
            setUserRole(storedRole);
            console.log('User role set to:', storedRole);
        } else {
            console.error('No user role found in localStorage');
            setError('Роль пользователя не определена. Пожалуйста, войдите в систему заново.');
            setLoading(false);
        }
    }, []);

    const fetchDashboard = useCallback(async () => {
        try {
            console.log('Starting dashboard fetch...');
            setLoading(true);
            setError('');
            
            // Проверяем наличие токена
            const token = localStorage.getItem('token');
            if (!token) {
                console.error('No token found');
                setError('Токен авторизации не найден. Пожалуйста, войдите в систему.');
                setLoading(false);
                return;
            }
            
            const endpoint = (userRole === 'admin' || userRole === 'main_admin') ? '/admin/dashboard' : '/partner/dashboard';
            console.log('Fetching from endpoint:', endpoint);
            console.log('User role:', userRole);
            console.log('Token exists:', !!token);
            
            const res = await api.get(endpoint);
            console.log('Dashboard response:', res);
            
            const payload = res.data || {};
            console.log('Dashboard payload:', payload);
            
            if (userRole === 'admin' || userRole === 'main_admin') {
                setData({
                    partners: payload.partners || { total: 0, active: 0 },
                    coupons: payload.coupons || { total: 0, used: 0, new: 0, purchased: 0 },
                    image_processing: payload.image_processing || { total: 0, processing: 0, completed: 0, failed: 0 },
                    recent_activations: payload.recent_activations || [],
                });
            } else {
                setData({
                    partners: undefined,
                    coupons: payload.coupons || { total: 0, used: 0, new: 0, purchased: 0 },
                    image_processing: undefined,
                    recent_activations: payload.recent_activations || [],
                });
            }
            
            console.log('Data set successfully');
        } catch (e) {
            console.error('Dashboard fetch error:', e);
            console.error('Error response:', e?.response);
            console.error('Error message:', e?.message);
            console.error('Error status:', e?.response?.status);
            
            let errorMessage = t('dashboard.failed_to_load_data');
            
            if (e?.response?.status === 401) {
                errorMessage = 'Ошибка авторизации. Пожалуйста, войдите в систему заново.';
                
                localStorage.removeItem('token');
                localStorage.removeItem('refresh_token');
                localStorage.removeItem('userRole');
                localStorage.removeItem('userId');
                localStorage.removeItem('userEmail');
                localStorage.removeItem('userName');
                window.location.href = '/';
                return;
            } else if (e?.response?.status === 403) {
                errorMessage = 'Доступ запрещен. Недостаточно прав.';
            } else if (e?.response?.status === 404) {
                errorMessage = 'Эндпоинт не найден. Обратитесь к администратору.';
            } else if (e?.response?.status >= 500) {
                errorMessage = 'Ошибка сервера. Попробуйте позже.';
            } else if (e?.message) {
                errorMessage = e.message;
            }
            
            setError(errorMessage);
        } finally {
            console.log('Setting loading to false');
            setLoading(false);
        }
    }, [userRole, t]);

    useEffect(() => {
        if (userRole) {
            fetchDashboard();
        }
    }, [userRole, fetchDashboard]);

    const systemStats = useMemo(() => {
        const totalPartners = (userRole === 'admin' || userRole === 'main_admin') ? (data.partners?.total || 0) : undefined;
        const activePartners = (userRole === 'admin' || userRole === 'main_admin') ? (data.partners?.active || 0) : undefined;
        const blockedPartners = (userRole === 'admin' || userRole === 'main_admin') ? Math.max((totalPartners || 0) - (activePartners || 0), 0) : undefined;
        const totalCoupons = data.coupons?.total || 0;
        const activatedCoupons = data.coupons?.used || data.coupons?.activated || 0;
        const activationRate = totalCoupons > 0 ? Math.round((activatedCoupons / totalCoupons) * 1000) / 10 : 0;
        return {
            totalPartners,
            activePartners,
            blockedPartners,
            totalCoupons,
            activatedCoupons,
            activationRate,
        };
    }, [data, userRole]);

    const imageProcessingChart = useMemo(() => {
        const items = [
            { 
                name: t('dashboard.in_processing'), 
                value: data.image_processing?.processing || 0, 
                color: '#0d6efd' 
            },
            { 
                name: t('dashboard.ready'), 
                value: data.image_processing?.completed || 0, 
                color: '#198754' 
            },
            { 
                name: t('dashboard.error'), 
                value: data.image_processing?.failed || 0, 
                color: '#dc3545' 
            },
        ];
        return items;
    }, [data, t]);

    const formatDateTime = (iso) => {
        try {
            const d = new Date(iso);
            if (Number.isNaN(d.getTime())) return '-';
            return d.toLocaleString();
        } catch {
            return '-';
        }
    };

    
    if (!userRole) {
        return (
            <div className="alert alert-danger" role="alert">
                <i className="pe-7s-attention mr-2"></i>
                {error || 'Инициализация...'}
            </div>
        );
    }

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-home icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('dashboard.main_panel')}
                            <div className="page-title-subheading">
                                {t('dashboard.system_overview')}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button color="primary" size="lg" onClick={fetchDashboard} disabled={loading}>
                            <i className="pe-7s-refresh"></i> {t('dashboard.refresh_data')}
                        </Button>
                    </div>
                </div>
            </div>

            {!!error && (
                <div className="alert alert-danger" role="alert">
                    <i className="pe-7s-attention mr-2"></i>
                    {error}
                </div>
            )}

            <Row>
                {(userRole === 'admin' || userRole === 'main_admin') && (
                    <Col xl="3" md="6">
                        <Card className="main-card mb-3">
                            <CardBody>
                                <div className="d-flex align-items-center">
                                    <div className="icon-wrapper rounded-circle text-primary mr-3">
                                        <i className="pe-7s-users" style={{ fontSize: '2em' }}></i>
                                    </div>
                                    <div>
                                        <div className="text-muted small">{t('dashboard.total_partners')}</div>
                                        <div className="text-dark font-weight-bold h4">{systemStats.totalPartners}</div>
                                        <div>
                                            <Badge color="success">{systemStats.activePartners} {t('dashboard.active')}</Badge>
                                            <Badge color="danger" className="ml-2">{systemStats.blockedPartners} {t('dashboard.blocked')}</Badge>
                                        </div>
                                    </div>
                                </div>
                            </CardBody>
                        </Card>
                    </Col>
                )}

                <Col xl="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="d-flex align-items-center">
                                <div className="icon-wrapper rounded-circle text-info mr-3">
                                    <i className="pe-7s-ticket" style={{ fontSize: '2em' }}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">{t('dashboard.total_coupons')}</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.totalCoupons.toLocaleString()}</div>
                                    <div className="small text-muted">
                                        {t('dashboard.used_coupons')}: <strong>{(data.coupons?.used || 0).toLocaleString()}</strong> • {t('dashboard.new_coupons')}: <strong>{(data.coupons?.new || 0).toLocaleString()}</strong>
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>

                <Col xl="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="d-flex align-items-center">
                                <div className="icon-wrapper rounded-circle text-success mr-3">
                                    <i className="pe-7s-check" style={{ fontSize: '2em' }}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">{t('dashboard.activated_coupons')}</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.activatedCoupons.toLocaleString()}</div>
                                    <div className="small text-muted">{t('dashboard.purchased_coupons')}: <strong>{(data.coupons?.purchased || 0).toLocaleString()}</strong></div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>

                <Col xl="3" md="6">
                    <Card className="main-card mb-3" style={{ minHeight: '101px' }}>
                        <CardBody>
                            <div className="d-flex align-items-center">
                                <div className="icon-wrapper rounded-circle text-warning mr-3">
                                    <i className="pe-7s-graph1" style={{ fontSize: '2em' }}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">{t('dashboard.activation_coefficient')}</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.activationRate}%</div>
                                    <div className="progress" style={{ height: '4px' }}>
                                        <div
                                            className="progress-bar bg-warning"
                                            style={{ width: `${systemStats.activationRate}%` }}
                                        />
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg={(userRole === 'admin' || userRole === 'main_admin') ? "8" : "12"}>
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('dashboard.recent_coupon_activations')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            {loading ? (
                                <div className="text-center text-muted">{t('common.loading')}</div>
                            ) : (data.recent_activations?.length ? (
                                <Table size="sm" responsive>
                                    <thead>
                                        <tr>
                                            <th>{t('dashboard.coupon_code')}</th>
                                            <th>{t('dashboard.status')}</th>
                                            <th>{t('dashboard.activated')}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {data.recent_activations.map((c, idx) => (
                                            <tr key={idx}>
                                                <td><code>{c.code || c.Code}</code></td>
                                                <td><Badge color="success">used</Badge></td>
                                                <td>{formatDateTime(c.used_at || c.UsedAt)}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </Table>
                            ) : (
                                <div className="text-muted">{t('dashboard.no_recent_activations')}</div>
                            ))}
                            <div className="text-center mt-3">
                                <Button color="outline-primary" size="sm">
                                    {t('dashboard.refresh_list')}
                                </Button>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
                {(userRole === 'admin' || userRole === 'main_admin') && (
                    <Col lg="4">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('dashboard.image_processing_status')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <div style={{ width: '100%', height: 260 }}>
                                    <ResponsiveContainer>
                                        <PieChart>
                                            <Pie dataKey="value" data={imageProcessingChart} outerRadius={90} label>
                                                {imageProcessingChart.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip />
                                            <Legend />
                                        </PieChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardBody>
                        </Card>
                    </Col>
                )}
            </Row>

            {(userRole === 'admin' || userRole === 'main_admin') && (
                <Row>
                    <Col lg="12">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('dashboard.quick_actions')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="3" className="mb-2">
                                        <Button
                                            color="primary"
                                            size="lg"
                                            block
                                            onClick={() => window.location.href = '/#/partners/create'}
                                            style={{ height: '80px' }}
                                        >
                                            <i className="pe-7s-add-user" style={{ fontSize: '24px' }}></i><br />
                                            {t('dashboard.add_partner')}
                                        </Button>
                                    </Col>
                                    <Col md="3" className="mb-2">
                                        <Button
                                            color="info"
                                            size="lg"
                                            block
                                            onClick={() => window.location.href = '/#/coupons/create'}
                                            style={{ height: '80px' }}
                                        >
                                            <i className="pe-7s-ticket" style={{ fontSize: '24px' }}></i><br />
                                            {t('dashboard.create_coupons')}
                                        </Button>
                                    </Col>
                                    <Col md="3" className="mb-2">
                                        <Button
                                            color="success"
                                            size="lg"
                                            block
                                            onClick={() => window.location.href = '/#/analytics'}
                                            style={{ height: '80px' }}
                                        >
                                            <i className="pe-7s-graph1" style={{ fontSize: '24px' }}></i><br />
                                            {t('dashboard.statistics')}
                                        </Button>
                                    </Col>
                                    <Col md="3" className="mb-2">
                                        <Button
                                            color="warning"
                                            size="lg"
                                            block
                                            onClick={() => window.location.href = '/#/coupons/manage'}
                                            style={{ height: '80px' }}
                                        >
                                            <i className="pe-7s-ticket" style={{ fontSize: '24px' }}></i><br />
                                            {t('dashboard.coupon_management')}
                                        </Button>
                                    </Col>
                                </Row>
                            </CardBody>
                        </Card>
                    </Col>
                </Row>
            )}
        </Fragment>
    );
};

export default Dashboard; 