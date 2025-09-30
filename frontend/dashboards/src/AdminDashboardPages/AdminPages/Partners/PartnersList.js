import React, { Fragment, useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText,
    Table, Badge, Dropdown, DropdownToggle, DropdownMenu, DropdownItem
} from 'reactstrap';
import api from '../../../api/api';
import ConfirmModal from '../../../components/ConfirmModal';

const PartnersList = () => {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [isFilterOpen, setIsFilterOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [partners, setPartners] = useState([]);
    const [confirmOpen, setConfirmOpen] = useState(false);
    const [confirmTargetId, setConfirmTargetId] = useState(null);
    const [partnersStats, setPartnersStats] = useState({});
    const [sortKey, setSortKey] = useState('created_at'); 
    const [sortOrder, setSortOrder] = useState('desc'); 

    const loadPartners = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const params = new URLSearchParams();
            if (searchTerm.trim()) params.append('search', searchTerm.trim());
            if (statusFilter !== 'all') params.append('status', statusFilter);
            
            params.append('sort_by', sortKey);
            params.append('order', sortOrder);
            const res = await api.get(`/admin/partners?${params.toString()}`);
            const list = res.data?.partners || [];
            setPartners(list);
        } catch (e) {
            setError(e?.response?.data?.error || e.message || t('chat.failed_to_load_partners'));
        } finally {
            setLoading(false);
        }
    }, [searchTerm, statusFilter, sortKey, sortOrder, t]);

    useEffect(() => {
        loadPartners();
        
    }, [loadPartners]);

    useEffect(() => {
        
        const loadStats = async () => {
            try {
                const res = await api.get('/admin/statistics/partners');
                const map = {};
                (res.data?.partners || []).forEach(p => {
                    const stats = p.statistics || {};
                    map[p.partner_id] = {
                        total: Number(stats.total) || 0,
                        used: Number(stats.used) || 0,
                        purchased: Number(stats.purchased) || 0,
                    };
                });
                setPartnersStats(map);
            } catch (_) {}
        };
        loadStats();
    }, []);

    const filteredPartners = useMemo(() => {
        const term = searchTerm.toLowerCase();
        const filtered = (partners || []).filter(p => {
            const matchesSearch = !term ||
                (p.brand_name || '').toLowerCase().includes(term) ||
                (p.domain || '').toLowerCase().includes(term) ||
                (p.email || '').toLowerCase().includes(term);
            const matchesStatus = statusFilter === 'all' || p.status === statusFilter;
            return matchesSearch && matchesStatus;
        });
        
        const sorted = [...filtered].sort((a, b) => {
            if (sortKey === 'brand_name') {
                const av = (a.brand_name || '').toLowerCase();
                const bv = (b.brand_name || '').toLowerCase();
                const cmp = av.localeCompare(bv);
                return sortOrder === 'asc' ? cmp : -cmp;
            }
            const ad = new Date(a.created_at || 0).getTime();
            const bd = new Date(b.created_at || 0).getTime();
            const cmp = ad - bd;
            return sortOrder === 'asc' ? cmp : -cmp;
        });
        return sorted;
    }, [partners, searchTerm, statusFilter, sortKey, sortOrder]);

    const getStatusBadge = (status) => {
        return status === 'active' ? 
            <Badge color="success">{t('partners.active')}</Badge> : 
            <Badge color="danger">{t('partners.blocked')}</Badge>;
    };

    const handleEdit = (partnerId) => {
        window.location.href = `/#/partners/edit/${partnerId}`;
    };

    const handleView = (partnerId) => {
        window.location.href = `/#/partners/view/${partnerId}`;
    };

    const handleBlock = async (partnerId, currentStatus) => {
        try {
            const isActive = currentStatus === 'active';
            if (isActive) {
                await api.patch(`/admin/partners/${partnerId}/block`);
            } else {
                await api.patch(`/admin/partners/${partnerId}/unblock`);
            }
            await loadPartners();
        } catch (e) {
            alert(e?.response?.data?.error || e.message || t('chat.failed_to_change_partner_status'));
        }
    };

    const handleDeleteClick = (partnerId) => {
        setConfirmTargetId(partnerId);
        setConfirmOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (!confirmTargetId) return;
        try {
            await api.delete(`/admin/partners/${confirmTargetId}?confirm=true`);
            setPartners(prev => prev.filter(p => p.id !== confirmTargetId));
            setConfirmOpen(false);
            setConfirmTargetId(null);
            await loadPartners();
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
                            <i className="pe-7s-users icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('partners.partner_management')}
                            <div className="page-title-subheading">
                                {t('partners.partner_list_management')}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="primary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/partners/create'}
                        >
                            <i className="pe-7s-plus"></i> {t('partners.add_partner')}
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('partners.filters_and_search')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row className="align-items-center mb-3">
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
                                            onKeyDown={(e) => { if (e.key === 'Enter') loadPartners(); }}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="4" className="d-flex align-items-center">
                                    <Dropdown 
                                        isOpen={isFilterOpen} 
                                        toggle={() => setIsFilterOpen(!isFilterOpen)}
                                    >
                                        <DropdownToggle caret color="info" style={{ height: 36, lineHeight: '36px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                            {t('partners.status')}: {statusFilter === 'all' ? t('partners.all') : 
                                                    statusFilter === 'active' ? t('partners.active') : t('partners.blocked')}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStatusFilter('all')}>
                                                {t('partners.all')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('active')}>
                                                {t('partners.active')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('blocked')}>
                                                {t('partners.blocked')}
                                            </DropdownItem>
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2" className="text-end mt-2 mt-md-0">
                                    <div className="d-inline-flex justify-content-end align-items-center" style={{gap: 12}}>
                                        <small className="text-muted">
                                            {t('partners.found')}: {filteredPartners.length}
                                        </small>
                                        <Button color="primary" size="sm" onClick={loadPartners} disabled={loading}>
                                            <i className="pe-7s-refresh"></i> {t('partners.refresh')}
                                        </Button>
                                    </div>
                                </Col>
                            </Row>
                            <div style={{height: 8}} />
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <div className="d-flex justify-content-between align-items-center w-100" style={{gap: 20}}>
                                <CardTitle className="mb-0">{t('partners.partner_list')}</CardTitle>
                                <div className="d-flex align-items-center flex-wrap" style={{gap: 16}}>
                                    <div className="d-flex align-items-center" style={{marginRight: 8}}>
                                        <small className="text-muted" style={{letterSpacing: 0.5}}>{t('partners.sorting')}:</small>
                                    </div>
                                    <div className="d-flex align-items-center" style={{gap: 12}}>
                                        <select
                                            className="form-select form-select-sm"
                                            value={sortKey}
                                            onChange={(e) => setSortKey(e.target.value)}
                                            style={{ height: 38, minWidth: 210, textAlignLast: 'center', textAlign: 'center', paddingTop: 2, paddingBottom: 6 }}
                                        >
                                            <option value="created_at">{t('common.date')}</option>
                                            <option value="brand_name">{t('partners.brand_name')}</option>
                                        </select>
                                        <select
                                            className="form-select form-select-sm"
                                            value={sortOrder}
                                            onChange={(e) => setSortOrder(e.target.value)}
                                            style={{ height: 38, minWidth: 180, textAlignLast: 'center', textAlign: 'center', paddingTop: 2, paddingBottom: 6 }}
                                        >
                                            <option value="asc">{t('tables.first')}</option>
                                            <option value="desc">{t('tables.last')}</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                        </CardHeader>
                        <CardBody style={{paddingTop: 32}}>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>{t('common.code')}</th>
                                        <th>{t('partners.brand_name')}</th>
                                        <th>{t('partners.domain')}</th>
                                        <th>{t('common.email')}</th>
                                        <th>{t('common.phone')}</th>
                                        <th>{t('common.status')}</th>
                                        <th>{t('dashboard.total_coupons')}</th>
                                        <th>{t('common.date')}</th>
                                        <th>{t('common.actions')}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {loading ? (
                                        <tr><td colSpan="9" className="text-center text-muted">{t('common.loading')}</td></tr>
                                    ) : filteredPartners.map((partner) => (
                                        <tr key={partner.id}>
                                            <td>
                                                <Badge color="secondary">{partner.partner_code}</Badge>
                                            </td>
                                            <td>
                                                <strong>{partner.brand_name}</strong>
                                            </td>
                                            <td>
                                                <a href={`${partner.domain?.startsWith('http') ? partner.domain : 'https://' + partner.domain}`}
                                                   target="_blank" rel="noopener noreferrer">
                                                    {partner.domain}
                                                </a>
                                            </td>
                                            <td>{partner.email || '-'}</td>
                                            <td>{partner.phone || '-'}</td>
                                            <td>{getStatusBadge(partner.status)}</td>
                                            <td>
                                                {(() => {
                                                    const s = partnersStats[partner.id];
                                                    return s ? (
                                                        <small className="text-muted">
                                                            {t('common.total')}: {s.total}, {t('dashboard.activated')}: {s.used}
                                                        </small>
                                                    ) : <small className="text-muted">—</small>;
                                                })()}
                                            </td>
                                            <td>{new Date(partner.created_at).toLocaleDateString()}</td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="primary"
                                                        onClick={() => handleEdit(partner.id)}
                                                        title={t('partners.edit_partner_data')}
                                                    >
                                                        <i className="pe-7s-edit"></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color="info"
                                                        onClick={() => handleView(partner.id)}
                                                        title={t('partners.view_partner_info')}
                                                    >
                                                        <i className="pe-7s-look"></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color="secondary"
                                                        onClick={() => navigate(`/admin/partners/${partner.id}/articles`)}
                                                        title="Управление артикулами"
                                                    >
                                                        <i className="pe-7s-box2"></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color={partner.status === 'active' ? 'warning' : 'success'}
                                                        onClick={() => handleBlock(partner.id, partner.status)}
                                                        title={partner.status === 'active' ? t('partners.block_partner_action') : t('partners.unblock_partner_action')}
                                                    >
                                                        <i className={partner.status === 'active' ? 'pe-7s-lock' : 'pe-7s-unlock'}></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color="danger"
                                                        onClick={() => handleDeleteClick(partner.id)}
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

                            {!loading && filteredPartners.length === 0 && (
                                <div className="text-center py-4">
                                    <p className="text-muted">{t('partners.partners_not_found')}</p>
                                </div>
                            )}
                            {!!error && (
                                <div className="alert alert-danger mt-3" role="alert">{error}</div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
            <ConfirmModal
                isOpen={confirmOpen}
                title={t('partners.delete_partner')}
                message={t('modals.delete_warning')}
                confirmText={t('common.delete')}
                confirmColor="danger"
                onConfirm={handleConfirmDelete}
                onCancel={() => { setConfirmOpen(false); setConfirmTargetId(null); }}
            />
        </Fragment>
    );
};

export default PartnersList; 