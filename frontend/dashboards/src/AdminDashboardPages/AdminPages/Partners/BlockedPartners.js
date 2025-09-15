import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText,
    Table
} from 'reactstrap';
import api from '../../../api/api';

const BlockedPartners = () => {
    const { t } = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [partners, setPartners] = useState([]);

    const loadBlockedPartners = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const res = await api.get('/admin/partners', { params: { status: 'blocked', search: searchTerm || '' } });
            setPartners(res.data?.partners || []);
        } catch (e) {
            setError(e?.response?.data?.error || e.message || t('chat.failed_to_load_blocked_partners'));
        } finally {
            setLoading(false);
        }
    }, [searchTerm, t]);

    useEffect(() => { loadBlockedPartners(); }, [loadBlockedPartners]);
    useEffect(() => { const t = setTimeout(loadBlockedPartners, 250); return () => clearTimeout(t); }, [searchTerm, loadBlockedPartners]);

    const filteredPartners = useMemo(() => {
        const term = (searchTerm || '').toLowerCase();
        return (partners || []).filter(p =>
            !term || (p.brand_name || '').toLowerCase().includes(term) || (p.domain || '').toLowerCase().includes(term) || (p.email || '').toLowerCase().includes(term)
        );
    }, [partners, searchTerm]);

    const handleUnblock = async (partnerId) => {
        try {
            await api.patch(`/admin/partners/${partnerId}/unblock`);
            await loadBlockedPartners();
        } catch (e) {
            alert(e?.response?.data?.error || e.message || t('chat.failed_to_unblock_partner'));
        }
    };

    const handleDelete = async (partnerId) => {
        if (!window.confirm(t('partners.confirm_delete_partner'))) return;
        try {
            await api.delete(`/admin/partners/${partnerId}?confirm=true`);
            await loadBlockedPartners();
        } catch (e) {
            alert(e?.response?.data?.error || e.message || t('chat.failed_to_delete_partner'));
        }
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-lock icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('partners.blocked_partners_title')}
                            <div className="page-title-subheading">
                                {t('partners.blocked_partners_description')}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/partners/list'}
                        >
                            <i className="pe-7s-back"></i> {t('partners.all_partners')}
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('common.search')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row>
                                <Col md="6">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder={t('partners.search_by_name_domain_email')}
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                        />
                                    </InputGroup>
                                </Col>
                            </Row>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('partners.blocked_partners_list')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            {loading ? (
                                <div className="text-center">
                                    <div className="spinner-border" role="status">
                                        <span className="sr-only">{t('common.loading')}</span>
                                    </div>
                                </div>
                            ) : error ? (
                                <div className="alert alert-danger">{error}</div>
                            ) : (
                                <div className="table-responsive">
                                    <Table>
                                        <thead>
                                            <tr>
                                                <th>{t('partners.brand_name')}</th>
                                                <th>{t('partners.domain')}</th>
                                                <th>{t('common.email')}</th>
                                                <th>{t('common.phone')}</th>
                                                <th>{t('partners.blocked_date')}</th>
                                                <th>{t('common.actions')}</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {filteredPartners.map((partner) => (
                                                <tr key={partner.id}>
                                                    <td>
                                                        <strong>{partner.brand_name}</strong>
                                                        <br/>
                                                        <small className="text-muted">
                                                            {t('common.code')}: {partner.partner_code}
                                                        </small>
                                                    </td>
                                                    <td>
                                                        <a href={`${partner.domain?.startsWith('http') ? partner.domain : 'https://' + partner.domain}`}
                                                           target="_blank" rel="noopener noreferrer">
                                                            {partner.domain}
                                                        </a>
                                                    </td>
                                                    <td>{partner.email || '-'}</td>
                                                    <td>{partner.phone || '-'}</td>
                                                    <td>
                                                        <small>{new Date(partner.updated_at).toLocaleDateString()}</small>
                                                    </td>
                                                    <td>
                                                        <div className="btn-group">
                                                            <Button
                                                                size="sm"
                                                                color="success"
                                                                onClick={() => handleUnblock(partner.id)}
                                                                title={t('partners.unblock_partner_action')}
                                                            >
                                                                <i className="pe-7s-unlock"></i>
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                color="danger"
                                                                onClick={() => handleDelete(partner.id)}
                                                                title={t('partners.delete_partner')}
                                                            >
                                                                <i className="pe-7s-trash"></i>
                                                            </Button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </Table>
                                </div>
                            )}

                            {!loading && filteredPartners.length === 0 && (
                                <div className="text-center py-4">
                                    <p className="text-muted">{t('partners.no_blocked_partners')}</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default BlockedPartners; 