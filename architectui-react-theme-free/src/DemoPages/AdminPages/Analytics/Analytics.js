import React, { Fragment } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Table, Badge, Progress
} from 'reactstrap';

const Analytics = () => {
    const overallStats = {
        totalCoupons: 1247,
        activatedCoupons: 723,
        activationRate: 58,
        activePartners: 23,
        partnersGrowth: 12
    };

    const partnerStats = [
        {
            id: '1001',
            name: 'Мозаика Арт',
            created: 150,
            activated: 89,
            purchased: 25,
            activationRate: 59
        },
        {
            id: '1002',
            name: 'Алмазная студия',
            created: 75,
            activated: 42,
            purchased: 15,
            activationRate: 56
        },
        {
            id: '1003',
            name: 'Творческая мастерская',
            created: 200,
            activated: 156,
            purchased: 45,
            activationRate: 78
        },
        {
            id: '0000',
            name: 'Собственные купоны',
            created: 822,
            activated: 436,
            purchased: 0,
            activationRate: 53
        }
    ];

    const recentActivity = [
        {
            id: 1,
            type: 'coupon_activated',
            description: 'Активирован купон 1001-2345-6789',
            time: '2 часа назад',
            partner: 'Мозаика Арт'
        },
        {
            id: 2,
            type: 'partner_created',
            description: 'Создан новый партнер "Алмазные картины"',
            time: '5 часов назад',
            partner: null
        },
        {
            id: 3,
            type: 'coupons_generated',
            description: 'Создано 50 купонов для партнера',
            time: '1 день назад',
            partner: 'Творческая мастерская'
        },
        {
            id: 4,
            type: 'coupon_purchased',
            description: 'Куплен купон через платформу',
            time: '1 день назад',
            partner: 'Алмазная студия'
        }
    ];

    const getActivityIcon = (type) => {
        switch (type) {
            case 'coupon_activated':
                return <i className="pe-7s-check text-success"></i>;
            case 'partner_created':
                return <i className="pe-7s-user text-primary"></i>;
            case 'coupons_generated':
                return <i className="pe-7s-ticket text-info"></i>;
            case 'coupon_purchased':
                return <i className="pe-7s-cash text-warning"></i>;
            default:
                return <i className="pe-7s-info text-muted"></i>;
        }
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-graph1 icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Статистика и аналитика
                            <div className="page-title-subheading">
                                Обзор производительности системы и активности партнеров
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-primary">
                                        <b>{overallStats.totalCoupons.toLocaleString()}</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Всего создано купонов
                                    </div>
                                    <div className="widget-description text-success">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>{overallStats.activationRate}%</span>
                                        </span>
                                        активировано
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
                
                <Col lg="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-success">
                                        <b>{overallStats.activatedCoupons}</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Активированных купонов
                                    </div>
                                    <div className="widget-description text-primary">
                                        За последний месяц
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-warning">
                                        <b>{overallStats.activePartners}</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Активных партнеров
                                    </div>
                                    <div className="widget-description text-success">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>{overallStats.partnersGrowth}%</span>
                                        </span>
                                        рост за месяц
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-info">
                                        <b>{overallStats.activationRate}%</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Средний процент активации
                                    </div>
                                    <Progress 
                                        value={overallStats.activationRate} 
                                        color="info" 
                                        className="mt-2"
                                    />
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg="8">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Графики активности (будут добавлены через Prometheus + Grafana)</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="text-center py-5">
                                <i className="pe-7s-graph text-muted" style={{fontSize: '4em'}}></i>
                                <h5 className="text-muted mt-3">Графики Grafana</h5>
                                <p className="text-muted">
                                    Здесь будут отображаться интерактивные графики от Grafana:
                                </p>
                                <ul className="list-unstyled text-muted">
                                    <li>• Динамика создания и активации купонов</li>
                                    <li>• Активность по партнерам</li>
                                    <li>• Популярные размеры и стили</li>
                                    <li>• Конверсия по времени</li>
                                </ul>
                            </div>
                        </CardBody>
                    </Card>

                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Детальная статистика по партнерам</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Партнер</th>
                                        <th>Создано</th>
                                        <th>Активировано</th>
                                        <th>Куплено</th>
                                        <th>% активации</th>
                                        <th>Прогресс</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {partnerStats.map((partner) => (
                                        <tr key={partner.id}>
                                            <td>
                                                <strong>{partner.name}</strong>
                                                <br/>
                                                <small className="text-muted">
                                                    Код: {partner.id}
                                                </small>
                                            </td>
                                            <td>
                                                <Badge color="secondary">
                                                    {partner.created}
                                                </Badge>
                                            </td>
                                            <td>
                                                <Badge color="success">
                                                    {partner.activated}
                                                </Badge>
                                            </td>
                                            <td>
                                                <Badge color="primary">
                                                    {partner.purchased}
                                                </Badge>
                                            </td>
                                            <td>
                                                <strong>{partner.activationRate}%</strong>
                                            </td>
                                            <td style={{width: '200px'}}>
                                                <Progress 
                                                    value={partner.activationRate} 
                                                    color={
                                                        partner.activationRate >= 70 ? 'success' :
                                                        partner.activationRate >= 50 ? 'warning' : 'danger'
                                                    }
                                                >
                                                    {partner.activationRate}%
                                                </Progress>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="4">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Последняя активность</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="timeline-wrapper">
                                {recentActivity.map((activity) => (
                                    <div key={activity.id} className="timeline-item">
                                        <div className="timeline-item-icon">
                                            {getActivityIcon(activity.type)}
                                        </div>
                                        <div className="timeline-item-content">
                                            <div className="timeline-item-description">
                                                {activity.description}
                                            </div>
                                            {activity.partner && (
                                                <div className="timeline-item-partner">
                                                    <small className="text-primary">
                                                        {activity.partner}
                                                    </small>
                                                </div>
                                            )}
                                            <div className="timeline-item-time">
                                                <small className="text-muted">
                                                    {activity.time}
                                                </small>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </CardBody>
                    </Card>

                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Быстрые действия</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="d-grid gap-2">
                                <button 
                                    className="btn btn-primary"
                                    onClick={() => window.location.href = '/#/coupons/create'}
                                >
                                    <i className="pe-7s-ticket"></i> Создать купоны
                                </button>
                                <button 
                                    className="btn btn-info"
                                    onClick={() => window.location.href = '/#/partners/create'}
                                >
                                    <i className="pe-7s-user"></i> Добавить партнера
                                </button>
                                <button 
                                    className="btn btn-warning"
                                    onClick={() => window.location.href = '/#/coupons/manage'}
                                >
                                    <i className="pe-7s-search"></i> Найти купон
                                </button>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <style jsx>{`
                .timeline-wrapper {
                    position: relative;
                }
                .timeline-item {
                    display: flex;
                    align-items: flex-start;
                    margin-bottom: 20px;
                    padding-bottom: 20px;
                    border-bottom: 1px solid #f0f0f0;
                }
                .timeline-item:last-child {
                    border-bottom: none;
                    margin-bottom: 0;
                    padding-bottom: 0;
                }
                .timeline-item-icon {
                    width: 30px;
                    height: 30px;
                    border-radius: 50%;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    background: #f8f9fa;
                    margin-right: 15px;
                    flex-shrink: 0;
                }
                .timeline-item-content {
                    flex: 1;
                }
                .timeline-item-description {
                    font-weight: 500;
                    margin-bottom: 5px;
                }
                .timeline-item-partner {
                    margin-bottom: 5px;
                }
                .timeline-item-time {
                    font-size: 0.875em;
                }
            `}</style>
        </Fragment>
    );
};

export default Analytics; 