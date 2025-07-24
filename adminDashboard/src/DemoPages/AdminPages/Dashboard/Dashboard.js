import React, { Fragment } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Badge, Table, Button
} from 'reactstrap';

const Dashboard = () => {
    const systemStats = {
        totalPartners: 12,
        activePartners: 9,
        blockedPartners: 3,
        totalCoupons: 15420,
        activatedCoupons: 8965,
        activationRate: 58.1,
        newCouponsToday: 45,
        activationsToday: 23
    };

    const recentActivities = [
        {
            id: 1,
            type: 'coupon_activated',
            message: 'Купон 1001-2345-6789 активирован пользователем',
            time: '5 минут назад',
            icon: 'pe-7s-check',
            color: 'success'
        },
        {
            id: 2,
            type: 'partner_created',
            message: 'Новый партнер "Мозаика Мастер" зарегистрирован',
            time: '2 часа назад',
            icon: 'pe-7s-add-user',
            color: 'info'
        },
        {
            id: 3,
            type: 'coupons_generated',
            message: 'Создано 100 новых купонов для партнера "АртМозаика"',
            time: '3 часа назад',
            icon: 'pe-7s-ticket',
            color: 'warning'
        },
        {
            id: 4,
            type: 'partner_blocked',
            message: 'Партнер "Творческая мастерская" заблокирован',
            time: '1 день назад',
            icon: 'pe-7s-lock',
            color: 'danger'
        }
    ];

    const topPartners = [
        { name: 'Мозаика Арт', activations: 1245, rate: 83 },
        { name: 'Алмазные картины', activations: 856, rate: 72 },
        { name: 'АртМозаика', activations: 634, rate: 65 }
    ];

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-home icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Главная панель
                            <div className="page-title-subheading">
                                Обзор системы алмазной мозаики и основная статистика
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button color="primary" size="lg">
                            <i className="pe-7s-refresh"></i> Обновить данные
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col xl="3" md="6">
                    <Card className="main-card mb-3">
                        <CardBody>
                            <div className="d-flex align-items-center">
                                <div className="icon-wrapper rounded-circle text-primary mr-3">
                                    <i className="pe-7s-users" style={{fontSize: '2em'}}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">Всего партнеров</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.totalPartners}</div>
                                    <div>
                                        <Badge color="success">{systemStats.activePartners} активных</Badge>
                                        <Badge color="danger" className="ml-2">{systemStats.blockedPartners} заблокированных</Badge>
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
                                <div className="icon-wrapper rounded-circle text-info mr-3">
                                    <i className="pe-7s-ticket" style={{fontSize: '2em'}}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">Всего купонов</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.totalCoupons.toLocaleString()}</div>
                                    <div>
                                        <Badge color="primary">+{systemStats.newCouponsToday} сегодня</Badge>
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
                                    <i className="pe-7s-check" style={{fontSize: '2em'}}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">Активированных</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.activatedCoupons.toLocaleString()}</div>
                                    <div>
                                        <Badge color="success">+{systemStats.activationsToday} сегодня</Badge>
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
                                <div className="icon-wrapper rounded-circle text-warning mr-3">
                                    <i className="pe-7s-graph1" style={{fontSize: '2em'}}></i>
                                </div>
                                <div>
                                    <div className="text-muted small">Коэффициент активации</div>
                                    <div className="text-dark font-weight-bold h4">{systemStats.activationRate}%</div>
                                    <div className="progress" style={{height: '4px'}}>
                                        <div 
                                            className="progress-bar bg-warning" 
                                            style={{width: `${systemStats.activationRate}%`}}
                                        />
                                    </div>
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
                            <CardTitle>Последние активности</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="timeline">
                                {recentActivities.map((activity) => (
                                    <div key={activity.id} className="timeline-item mb-3">
                                        <div className="d-flex align-items-center">
                                            <div className={`icon-wrapper rounded-circle text-${activity.color} mr-3`} style={{padding: '8px'}}>
                                                <i className={activity.icon}></i>
                                            </div>
                                            <div className="flex-grow-1">
                                                <div className="font-weight-bold">{activity.message}</div>
                                                <small className="text-muted">{activity.time}</small>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <div className="text-center mt-3">
                                <Button color="outline-primary" size="sm">
                                    Показать все активности
                                </Button>
                            </div>
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="4">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Топ партнеры</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table size="sm">
                                <thead>
                                    <tr>
                                        <th>Партнер</th>
                                        <th>Активации</th>
                                        <th>%</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {topPartners.map((partner, index) => (
                                        <tr key={index}>
                                            <td>
                                                <div className="d-flex align-items-center">
                                                    <Badge 
                                                        color={index === 0 ? 'warning' : index === 1 ? 'secondary' : 'light'}
                                                        className="mr-2"
                                                    >
                                                        {index + 1}
                                                    </Badge>
                                                    <small>{partner.name}</small>
                                                </div>
                                            </td>
                                            <td>
                                                <strong>{partner.activations}</strong>
                                            </td>
                                            <td>
                                                <Badge color="success">{partner.rate}%</Badge>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                            <div className="text-center mt-3">
                                <Button 
                                    color="outline-primary" 
                                    size="sm"
                                    onClick={() => window.location.href = '/#/analytics/partners'}
                                >
                                    Вся статистика
                                </Button>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Быстрые действия</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row>
                                <Col md="3" className="mb-2">
                                    <Button 
                                        color="primary" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/partners/create'}
                                        style={{height: '80px'}}
                                    >
                                        <i className="pe-7s-add-user" style={{fontSize: '24px'}}></i><br/>
                                        Добавить партнера
                                    </Button>
                                </Col>
                                <Col md="3" className="mb-2">
                                    <Button 
                                        color="info" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/create'}
                                        style={{height: '80px'}}
                                    >
                                        <i className="pe-7s-ticket" style={{fontSize: '24px'}}></i><br/>
                                        Создать купоны
                                    </Button>
                                </Col>
                                <Col md="3" className="mb-2">
                                    <Button 
                                        color="success" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/analytics'}
                                        style={{height: '80px'}}
                                    >
                                        <i className="pe-7s-graph1" style={{fontSize: '24px'}}></i><br/>
                                        Статистика
                                    </Button>
                                </Col>
                                <Col md="3" className="mb-2">
                                    <Button 
                                        color="warning" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/manage'}
                                        style={{height: '80px'}}
                                    >
                                        <i className="pe-7s-config" style={{fontSize: '24px'}}></i><br/>
                                        Управление купонами
                                    </Button>
                                </Col>
                            </Row>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default Dashboard; 