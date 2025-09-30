import React, { Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button
} from 'reactstrap';

const Dashboard = () => {
    const { t } = useTranslation();

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-home icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('dashboard.admin_panel_title')}
                            <div className="page-title-subheading">
                                {t('dashboard.main_control_panel')}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="6" md="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('dashboard.coupon_statistics')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-success">
                                        <b>1,247</b>
                                    </div>
                                    <div className="widget-subheading">
                                        {t('dashboard.total_created_coupons')}
                                    </div>
                                    <div className="widget-description text-primary">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>58%</span>
                                        </span>
                                        {t('dashboard.activated')}
                                    </div>
                                </div>
                            </div>
                        </CardBody>
                    </Card>
                </Col>
                
                <Col lg="6" md="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('dashboard.active_partners')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-chart-content">
                                    <div className="widget-numbers text-warning">
                                        <b>23</b>
                                    </div>
                                    <div className="widget-subheading">
                                        {t('dashboard.active_partners_count')}
                                    </div>
                                    <div className="widget-description text-success">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>12%</span>
                                        </span>
                                        {t('dashboard.growth_per_month')}
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
                            <CardTitle>{t('dashboard.quick_actions_title')}</CardTitle>
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
                                        <i className="pe-7s-plus"></i> {t('dashboard.add_partner')}
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="success" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/create'}
                                    >
                                        <i className="pe-7s-ticket"></i> {t('dashboard.create_coupons')}
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="info" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/coupons/manage'}
                                    >
                                        <i className="pe-7s-search"></i> {t('dashboard.search_coupons')}
                                    </Button>
                                </Col>
                                <Col md="3">
                                    <Button 
                                        color="warning" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/analytics'}
                                    >
                                        <i className="pe-7s-graph1"></i> {t('dashboard.statistics')}
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