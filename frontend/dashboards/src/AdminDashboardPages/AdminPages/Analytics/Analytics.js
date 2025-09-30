import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Row, Col, Card, CardBody, CardTitle, CardHeader, Table, Badge, Progress } from 'reactstrap';
import api from '../../../api/api';

const Analytics = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [couponStats, setCouponStats] = useState({ total: 0, new: 0, used: 0, purchased: 0 });
    const [partnersStats, setPartnersStats] = useState([]);

    const fetchStats = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const [overallRes, partnersRes] = await Promise.all([
                api.get('/admin/statistics'),
                api.get('/admin/statistics/partners'),
            ]);
            setCouponStats(overallRes.data?.coupon_statistics || { total: 0, new: 0, used: 0, purchased: 0 });
            setPartnersStats(partnersRes.data?.partners || []);
        } catch (e) {
            setError(e?.response?.data?.error || e.message || t('analytics.failed_to_load_statistics'));
        } finally {
            setLoading(false);
        }
    }, [t]);

    useEffect(() => { fetchStats(); }, [fetchStats]);

    const overallStats = useMemo(() => {
        const total = Number(couponStats.total || 0);
        const used = Number(couponStats.used || 0);
        const activationRate = total > 0 ? Math.round((used / total) * 100) : 0;
        return {
            totalCoupons: total,
            activatedCoupons: used,
            activationRate,
            activePartners: undefined,
            partnersGrowth: undefined,
        };
    }, [couponStats]);

    

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-graph1 icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('analytics.title')}
                            <div className="page-title-subheading">
                                {t('analytics.partner_performance_overview')}
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
                                        {t('analytics.total_created_coupons')}
                                    </div>
                                    <div className="widget-description text-success">
                                        <span className="pr-1">
                                            <i className="fa fa-angle-up"></i>
                                            <span>{overallStats.activationRate}%</span>
                                        </span>
                                        {t('analytics.activated_coupons')}
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
                                        {t('analytics.activated_coupons')}
                                    </div>
                                    <div className="widget-description text-primary">
                                        {t('analytics.last_month')}
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
                                        <b>{partnersStats.length}</b>
                                    </div>
                                    <div className="widget-subheading">
                                        {t('analytics.partners')}
                                    </div>
                                    <div className="widget-description text-info">
                                        {t('analytics.active_partners')}
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
                                        {t('analytics.activation_rate')}
                                    </div>
                                    <div className="widget-description text-success">
                                        {t('analytics.activated_coupons')}
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
                            <CardTitle>{t('analytics.detailed_partner_statistics')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            {loading ? (
                                <div className="text-center">
                                    <div className="spinner-border" role="status">
                                        <span className="sr-only">{t('loading')}</span>
                                    </div>
                                </div>
                            ) : error ? (
                                <div className="alert alert-danger">{error}</div>
                            ) : (
                                <Table responsive>
                                    <thead>
                                        <tr>
                                            <th>{t('partners.brand_name')}</th>
                                            <th>{t('partners.domain')}</th>
                                            <th>{t('partners.status')}</th>
                                            <th>{t('analytics.total_coupons')}</th>
                                            <th>{t('analytics.activated_coupons')}</th>
                                            <th>{t('analytics.activation_rate')}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {partnersStats.map((partner, index) => {
                                            const stats = partner.statistics || {};
                                            const total = Number(stats.total) || 0;
                                            const used = Number(stats.used) || 0;
                                            const rate = total > 0 ? Math.round((used / total) * 100) : 0;
                                            
                                            return (
                                                <tr key={partner.partner_id || index}>
                                                    <td>{partner.brand_name || '-'}</td>
                                                    <td>{partner.domain || '-'}</td>
                                                    <td>
                                                        <Badge color={partner.status === 'active' ? 'success' : 'danger'}>
                                                            {t(`partners.${partner.status}`)}
                                                        </Badge>
                                                    </td>
                                                    <td>{total.toLocaleString()}</td>
                                                    <td>{used.toLocaleString()}</td>
                                                    <td>
                                                        <div className="d-flex align-items-center">
                                                            <span className="mr-2">{rate}%</span>
                                                            <Progress 
                                                                value={rate} 
                                                                className="flex-grow-1" 
                                                                style={{ height: '6px' }}
                                                            />
                                                        </div>
                                                    </td>
                                                </tr>
                                            );
                                        })}
                                    </tbody>
                                </Table>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default Analytics; 