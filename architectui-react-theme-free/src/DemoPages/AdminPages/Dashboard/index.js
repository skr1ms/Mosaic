import React, { Fragment } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button
} from 'reactstrap';

const Dashboard = () => {
    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-home icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Админ панель алмазной мозаики
                            <div className="page-title-subheading">
                                Главная панель управления системой
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="6" md="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Статистика купонов</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-success">
                                        <b>1,247</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Всего создано купонов
                                    </div>
                                    <div className="widget-description text-primary">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>58%</span>
                                        </span>
                                        активировано
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
                
                <Col lg="6" md="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Активные партнеры</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-warning">
                                        <b>23</b>
                                    </div>
                                    <div className="widget-subheading">
                                        Активных партнеров
                                    </div>
                                    <div className="widget-description text-success">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>12%</span>
                                        </span>
                                        рост за месяц
                                    </div>
                                </div>
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
                                <Col md="3">
                                    <Button 
                                        color="primary" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/partners/create'}
                                    >
                                        <i className="pe-7s-plus"></i> Добавить партнера
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="success" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/create'}
                                    >
                                        <i className="pe-7s-ticket"></i> Создать купоны
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="info" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/manage'}
                                    >
                                        <i className="pe-7s-search"></i> Поиск купонов
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="warning" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/analytics'}
                                    >
                                        <i className="pe-7s-graph1"></i> Статистика
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