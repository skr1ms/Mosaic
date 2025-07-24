import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Table, Badge,
    UncontrolledDropdown, DropdownToggle, DropdownMenu, DropdownItem
} from 'reactstrap';

const PartnerAnalytics = () => {
    const [selectedPeriod, setSelectedPeriod] = useState('month');

    // Моковые данные статистики партнеров
    const partnerStats = [
        {
            id: 1,
            partnerCode: '1001',
            brandName: 'Мозаика Арт',
            totalCoupons: 1500,
            activatedCoupons: 1245,
            activationRate: 83.0,
            purchasedOnSite: 234,
            lastActivity: '2023-12-15',
            revenue: 125000
        },
        {
            id: 2,
            partnerCode: '1002',
            brandName: 'Алмазные картины',
            totalCoupons: 800,
            activatedCoupons: 456,
            activationRate: 57.0,
            purchasedOnSite: 89,
            lastActivity: '2023-12-10',
            revenue: 78000
        }
    ];

    const totalStats = {
        totalPartners: 12,
        activePartners: 9,
        totalCoupons: 15420,
        activatedCoupons: 8965,
        totalRevenue: 1250000
    };

    const periodOptions = [
        { value: 'week', label: 'За неделю' },
        { value: 'month', label: 'За месяц' },
        { value: 'year', label: 'За год' }
    ];

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-users icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Статистика по партнерам
                            <div className="page-title-subheading">
                                Детальная аналитика активности всех партнеров системы
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <UncontrolledDropdown>
                            <DropdownToggle caret color="primary">
                                <i className="pe-7s-date"></i> {periodOptions.find(p => p.value === selectedPeriod)?.label}
                            </DropdownToggle>
                            <DropdownMenu right>
                                {periodOptions.map(period => (
                                    <DropdownItem 
                                        key={period.value}
                                        onClick={() => setSelectedPeriod(period.value)}
                                        active={selectedPeriod === period.value}
                                    >
                                        {period.label}
                                    </DropdownItem>
                                ))}
                            </DropdownMenu>
                        </UncontrolledDropdown>
                    </div>
                </div>
            </div>

            {/* Общая статистика */}
            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Общие показатели</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row>
                                <Col md="2">
                                    <div className="text-center">
                                        <h3 className="text-primary">{totalStats.totalPartners}</h3>
                                        <small className="text-muted">Всего партнеров</small>
                                    </div>
                                </Col>
                                <Col md="2">
                                    <div className="text-center">
                                        <h3 className="text-success">{totalStats.activePartners}</h3>
                                        <small className="text-muted">Активных</small>
                                    </div>
                                </Col>
                                <Col md="3">
                                    <div className="text-center">
                                        <h3 className="text-info">{totalStats.totalCoupons.toLocaleString()}</h3>
                                        <small className="text-muted">Купонов создано</small>
                                    </div>
                                </Col>
                                <Col md="3">
                                    <div className="text-center">
                                        <h3 className="text-warning">{totalStats.activatedCoupons.toLocaleString()}</h3>
                                        <small className="text-muted">Купонов активировано</small>
                                    </div>
                                </Col>
                                <Col md="2">
                                    <div className="text-center">
                                        <h3 className="text-dark">{(totalStats.totalRevenue / 1000).toFixed(0)}K ₽</h3>
                                        <small className="text-muted">Общий доход</small>
                                    </div>
                                </Col>
                            </Row>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            {/* Детальная статистика по партнерам */}
            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Статистика по каждому партнеру</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Партнер</th>
                                        <th>Купоны создано</th>
                                        <th>Купоны активировано</th>
                                        <th>% активации</th>
                                        <th>Продано на сайте</th>
                                        <th>Доход</th>
                                        <th>Последняя активность</th>
                                        <th>Статус</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {partnerStats.map((partner) => (
                                        <tr key={partner.id}>
                                            <td>
                                                <div>
                                                    <strong>{partner.brandName}</strong>
                                                    <br/>
                                                    <small className="text-muted">
                                                        Код: {partner.partnerCode}
                                                    </small>
                                                </div>
                                            </td>
                                            <td>
                                                <Badge color="info" className="badge-pill">
                                                    {partner.totalCoupons.toLocaleString()}
                                                </Badge>
                                            </td>
                                            <td>
                                                <Badge color="success" className="badge-pill">
                                                    {partner.activatedCoupons.toLocaleString()}
                                                </Badge>
                                            </td>
                                            <td>
                                                <div className="d-flex align-items-center">
                                                    <div 
                                                        className="progress" 
                                                        style={{width: '100px', height: '8px'}}
                                                    >
                                                        <div 
                                                            className="progress-bar bg-success" 
                                                            style={{width: `${partner.activationRate}%`}}
                                                        />
                                                    </div>
                                                    <span className="ml-2 text-muted">
                                                        {partner.activationRate}%
                                                    </span>
                                                </div>
                                            </td>
                                            <td>
                                                <strong>{partner.purchasedOnSite}</strong>
                                            </td>
                                            <td>
                                                <strong className="text-success">
                                                    {(partner.revenue / 1000).toFixed(0)}K ₽
                                                </strong>
                                            </td>
                                            <td>
                                                <small>{partner.lastActivity}</small>
                                            </td>
                                            <td>
                                                <Badge 
                                                    color={partner.activationRate > 70 ? 'success' : 
                                                           partner.activationRate > 40 ? 'warning' : 'danger'}
                                                >
                                                    {partner.activationRate > 70 ? 'Отлично' :
                                                     partner.activationRate > 40 ? 'Средне' : 'Низко'}
                                                </Badge>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            {/* Графики и дополнительная аналитика */}
            <Row>
                <Col md="6">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Топ партнеры по активации</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="text-center py-4">
                                <i className="pe-7s-graph1 text-primary" style={{fontSize: '3em'}}></i>
                                <p className="text-muted mt-3">
                                    График будет добавлен позже<br/>
                                    <small>Интеграция с Grafana/Prometheus</small>
                                </p>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
                <Col md="6">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Динамика продаж</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="text-center py-4">
                                <i className="pe-7s-cash text-success" style={{fontSize: '3em'}}></i>
                                <p className="text-muted mt-3">
                                    График динамики продаж<br/>
                                    <small>Интеграция с Grafana/Prometheus</small>
                                </p>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default PartnerAnalytics; 