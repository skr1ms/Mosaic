import React, { Fragment, useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Row, Col, Card, CardBody, CardTitle, CardHeader, Table, Badge, Input, Label, Button } from 'reactstrap';
import api from '../../../api/api';

const PartnerAnalytics = () => {
    const { t } = useTranslation();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [partnersStats, setPartnersStats] = useState([]);
    const [filters, setFilters] = useState({
        search: '',
        status: '',
    });
    const [sortField, setSortField] = useState('created'); 
    const [sortDir, setSortDir] = useState('desc'); 

    const processedPartners = React.useMemo(() => {
        const normalizedSearch = (filters.search || '').toLowerCase();
        const statusFilter = (filters.status || '').toLowerCase();
        const list = (partnersStats || []).filter((p) => {
            const statusOk = statusFilter ? String(p.status || '').toLowerCase() === statusFilter : true;
            const hay = `${String(p.brand_name || '')} ${String(p.partner_code || '')}`.toLowerCase();
            const searchOk = normalizedSearch ? hay.includes(normalizedSearch) : true;
            return statusOk && searchOk;
        });
        const withMetrics = list.map((p) => {
            const created = p.total_coupons ?? p.statistics?.total ?? 0;
            const activated = p.activated_coupons ?? p.statistics?.used ?? 0;
            const purchased = p.purchased_coupons ?? p.statistics?.purchased ?? 0;
            const activationRate = created > 0 ? Math.round((activated / created) * 100) : 0;
            return { p, created, activated, purchased, activationRate };
        });
        withMetrics.sort((a, b) => {
            let va = a[sortField];
            let vb = b[sortField];
            if (va < vb) return sortDir === 'asc' ? -1 : 1;
            if (va > vb) return sortDir === 'asc' ? 1 : -1;
            return 0;
        });
        return withMetrics;
    }, [partnersStats, filters, sortField, sortDir]);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const params = new URLSearchParams();
            if (filters.search) params.append('search', filters.search);
            if (filters.status) params.append('status', filters.status);
            const query = params.toString();
            const res = await api.get(`/admin/statistics/partners${query ? `?${query}` : ''}`);
            setPartnersStats(res.data?.partners || []);
        } catch (e) {
            setError(e?.response?.data?.error || e.message || t('partners.failed_to_load_partner_statistics'));
        } finally {
            setLoading(false);
        }
    }, [filters, t]);

    useEffect(() => { load(); }, [load]);

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-users icon-gradient bg-mean-fruit" />
                        </div>
                        <div>{t('partners.partner_analytics_title')}
                            <div className="page-title-subheading">{t('partners.partner_analytics_subtitle')}</div>
                        </div>
                    </div>
                </div>
            </div>

            <Row className="mb-3">
                <Col md="4">
                    <Label className="small mb-1">{t('partners.search_by_name_or_code')}</Label>
                    <Input placeholder={t('partners.search_placeholder_mosaic_or_code')} value={filters.search}
                        onChange={e => setFilters(prev => ({ ...prev, search: e.target.value }))} />
                </Col>
                <Col md="3">
                    <Label className="small mb-1">{t('partners.partner_status')}</Label>
                    <Input type="select" value={filters.status} onChange={e => setFilters(prev => ({ ...prev, status: e.target.value }))}>
                        <option value="">{t('partners.any_status')}</option>
                        <option value="active">active</option>
                        <option value="blocked">blocked</option>
                    </Input>
                </Col>
                <Col md="4">
                    <Label className="small mb-1">{t('partners.sort_by')}</Label>
                    <div className="d-flex" style={{gap: 8}}>
                        <Input type="select" value={sortField} onChange={e => setSortField(e.target.value)} style={{minWidth: 200}}>
                            <option value="created">{t('partners.sort_by_created_coupons')}</option>
                            <option value="activated">{t('partners.sort_by_activated_coupons')}</option>
                            <option value="purchased">{t('partners.sort_by_purchased_coupons')}</option>
                            <option value="activationRate">{t('partners.sort_by_activation_rate')}</option>
                        </Input>
                        <Input type="select" value={sortDir} onChange={e => setSortDir(e.target.value)} style={{minWidth: 180}}>
                            <option value="asc">{t('partners.sort_ascending')}</option>
                            <option value="desc">{t('partners.sort_descending')}</option>
                        </Input>
                    </div>
                </Col>
                <Col md="1" className="d-flex align-items-end justify-content-end" style={{marginBottom: 4}}>
                    <Button color="primary" size="sm" onClick={load} disabled={loading}>
                        <i className="pe-7s-refresh"></i> {t('refresh')}
                    </Button>
                </Col>
            </Row>
            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('partners.statistics_by_partner')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            {error && <div className="alert alert-danger">{error}</div>}
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>{t('partners.partner')}</th>
                                        <th>{t('partners.coupons_created')}</th>
                                        <th>{t('partners.coupons_activated')}</th>
                                        <th>{t('partners.purchased_on_site')}</th>
                                        <th>{t('partners.activation_percentage')}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {loading ? (
                                        <tr><td colSpan="5" className="text-center text-muted">{t('loading')}</td></tr>
                                    ) : partnersStats.length === 0 ? (
                                        <tr><td colSpan="5" className="text-center text-muted">{t('partners.no_data')}</td></tr>
                                    ) : processedPartners.length === 0 ? (
                                        <tr><td colSpan="5" className="text-center text-muted">{t('partners.nothing_found_by_filters')}</td></tr>
                                    ) : processedPartners.map(({ p, created, activated, purchased, activationRate }) => (
                                        <tr key={p.partner_id}>
                                            <td>
                                                <div className="d-flex align-items-center" style={{gap: 8}}>
                                                    <strong>{p.brand_name}</strong>
                                                    <Badge color="light" className="text-muted" style={{minWidth: 54, textAlign: 'center'}}>
                                                        {p.partner_code}
                                                    </Badge>
                                                </div>
                                            </td>
                                            <td><Badge color="secondary">{created}</Badge></td>
                                            <td><Badge color="success">{activated}</Badge></td>
                                            <td><Badge color="primary">{purchased}</Badge></td>
                                            <td><strong>{activationRate}%</strong></td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default PartnerAnalytics;