import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText, Table, Badge,
    Modal, ModalHeader, ModalBody, ModalFooter
} from 'reactstrap';
import api from '../../../api/api';

const ActivatedCoupons = () => {
    const { t } = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [coupons, setCoupons] = useState([]);
    const [page] = useState(1);
    const [limit] = useState(20);
    const [sortKey, setSortKey] = useState('used_at'); 
    const [sortOrder, setSortOrder] = useState('desc'); 
    const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);
    const [selectedCoupon, setSelectedCoupon] = useState(null);

    const load = useCallback(async () => {
        try {
            setLoading(true);
            setError('');
            const params = new URLSearchParams();
            params.append('status', 'used');
            params.append('page', String(page));
            params.append('limit', String(limit));
            const term = searchTerm.trim();
            if (term) params.append('search', term);
            const res = await api.get(`/admin/coupons/paginated?${params.toString()}`);
            const list = res.data?.coupons || [];
            setCoupons(list);
        } catch (e) {
            setError(e?.response?.data?.error || e.message || t('coupons.failed_to_load_activated'));
        } finally {
            setLoading(false);
        }
    }, [page, limit, searchTerm, t]);

    useEffect(() => { load(); }, [page, limit, load]);

    const sizeOptions = [
        { value: '21x30', label: t('coupons.size_options.21x30') },
        { value: '30x40', label: t('coupons.size_options.30x40') },
        { value: '40x40', label: t('coupons.size_options.40x40') },
        { value: '40x50', label: t('coupons.size_options.40x50') },
        { value: '40x60', label: t('coupons.size_options.40x60') },
        { value: '50x70', label: t('coupons.size_options.50x70') }
    ];

    const styleOptions = [
        { value: 'grayscale', label: t('coupons.style_options.grayscale') },
        { value: 'skin_tones', label: t('coupons.style_options.skin_tones') },
        { value: 'pop_art', label: t('coupons.style_options.pop_art') },
        { value: 'max_colors', label: t('coupons.style_options.max_colors') }
    ];

    const filteredCoupons = useMemo(() => {
        const term = searchTerm.toLowerCase();
        const filtered = (coupons || []).filter(c => {
            const partnerName = (c.partner_name || '').toLowerCase();
            const code = (c.code || '').toLowerCase();
            const email = (c.purchase_email || '').toLowerCase();
            return !term || code.includes(term) || email.includes(term) || partnerName.includes(term);
        });

        // Сортировка на клиенте
        const sorted = [...filtered].sort((a, b) => {
            const dir = sortOrder === 'asc' ? 1 : -1;
            switch (sortKey) {
                case 'code': {
                    const av = (a.code || '').toLowerCase();
                    const bv = (b.code || '').toLowerCase();
                    return av.localeCompare(bv) * dir;
                }
                case 'created_at': {
                    const ad = new Date(a.created_at || 0).getTime();
                    const bd = new Date(b.created_at || 0).getTime();
                    return (ad - bd) * dir;
                }
                case 'used_at':
                default: {
                    const ad = new Date(a.used_at || 0).getTime();
                    const bd = new Date(b.used_at || 0).getTime();
                    return (ad - bd) * dir;
                }
            }
        });
        return sorted;
    }, [coupons, searchTerm, sortKey, sortOrder]);

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString({hour: '2-digit', minute: '2-digit'});
    };

    const handleDownloadMaterials = async (coupon) => {
        try {
            const res = await api.get(`/admin/coupons/${coupon.id}/materials`);
            if (res.data?.download_url) {
                window.open(res.data.download_url, '_blank');
            }
        } catch (e) {
            alert(e?.response?.data?.error || e.message || t('coupons.failed_to_download_materials'));
        }
    };

    const openDetailModal = (coupon) => {
        setSelectedCoupon(coupon);
        setIsDetailModalOpen(true);
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-ticket icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('coupons.activated', 'Активированные купоны')}
                            <div className="page-title-subheading">
                                {t('coupons.activated_coupons_description', 'Список всех погашенных купонов с материалами пользователей')}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('coupons.search_and_filter')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row className="align-items-center mb-3">
                                <Col md="8">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder={t('partners.search_by_coupon_number')}
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                            onKeyDown={(e) => { if (e.key === 'Enter') load(); }}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="4" className="text-end">
                                    <Button color="primary" onClick={load} disabled={loading}>
                                        <i className="pe-7s-refresh"></i> {t('common.refresh')}
                                    </Button>
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
                            <div className="d-flex justify-content-between align-items-center w-100">
                                <CardTitle className="mb-0">{t('coupons.activated_list')}</CardTitle>
                                <div className="d-flex align-items-center" style={{gap: 16}}>
                                    <div className="d-flex align-items-center" style={{marginRight: 8}}>
                                        <small className="text-muted">{t('partners.sorting')}:</small>
                                    </div>
                                    <select
                                        className="form-select form-select-sm"
                                        value={sortKey}
                                        onChange={(e) => setSortKey(e.target.value)}
                                        style={{ height: 38, minWidth: 180 }}
                                    >
                                        <option value="used_at">{t('coupons.used_at')}</option>
                                        <option value="created_at">{t('coupons.created_at')}</option>
                                        <option value="code">{t('coupons.code')}</option>
                                    </select>
                                    <select
                                        className="form-select form-select-sm"
                                        value={sortOrder}
                                        onChange={(e) => setSortOrder(e.target.value)}
                                        style={{ height: 38, minWidth: 160 }}
                                    >
                                        <option value="desc">{t('common.sort.descending')}</option>
                                        <option value="asc">{t('common.sort.ascending')}</option>
                                    </select>
                                </div>
                            </div>
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
                                <div className="table-responsive">
                                    <Table>
                                        <thead>
                                            <tr>
                                                <th>{t('coupons.code')}</th>
                                                <th>{t('partners.brand_name')}</th>
                                                <th>{t('coupons.size')}</th>
                                                <th>{t('coupons.style')}</th>
                                                <th>{t('coupons.used_at')}</th>
                                                <th>{t('coupons.purchase_email')}</th>
                                                <th>{t('actions')}</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {filteredCoupons.map((coupon) => (
                                                <tr key={coupon.id}>
                                                    <td>
                                                        <Badge color="success">{coupon.code}</Badge>
                                                    </td>
                                                    <td>{coupon.partner_name || '-'}</td>
                                                    <td>{sizeOptions.find(s => s.value === coupon.size)?.label || '-'}</td>
                                                    <td>{styleOptions.find(s => s.value === coupon.style)?.label || '-'}</td>
                                                    <td>{formatDate(coupon.used_at)}</td>
                                                    <td>{coupon.purchase_email || '-'}</td>
                                                    <td>
                                                        <div className="btn-group">
                                                            <Button
                                                                size="sm"
                                                                color="info"
                                                                onClick={() => openDetailModal(coupon)}
                                                            >
                                                                {t('view')}
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                color="primary"
                                                                onClick={() => handleDownloadMaterials(coupon)}
                                                            >
                                                                {t('coupons.download_materials')}
                                                            </Button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </Table>
                                </div>
                            )}

                            {!loading && filteredCoupons.length === 0 && (
                                <div className="text-center py-4">
                                    <p className="text-muted">{t('coupons.no_activated_coupons')}</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Modal isOpen={isDetailModalOpen} toggle={() => setIsDetailModalOpen(false)} size="lg">
                <ModalHeader toggle={() => setIsDetailModalOpen(false)}>
                    {t('coupons.coupon_details')}: {selectedCoupon?.code}
                </ModalHeader>
                <ModalBody>
                    {selectedCoupon && (
                        <div>
                            <Row>
                                <Col md="6">
                                    <p><strong>{t('coupons.code')}:</strong> {selectedCoupon.code}</p>
                                    <p><strong>{t('partners.brand_name')}:</strong> {selectedCoupon.partner_name || '-'}</p>
                                    <p><strong>{t('coupons.size')}:</strong> {sizeOptions.find(s => s.value === selectedCoupon.size)?.label || '-'}</p>
                                    <p><strong>{t('coupons.style')}:</strong> {styleOptions.find(s => s.value === selectedCoupon.style)?.label || '-'}</p>
                                </Col>
                                <Col md="6">
                                    <p><strong>{t('coupons.created_at')}:</strong> {formatDate(selectedCoupon.created_at)}</p>
                                    <p><strong>{t('coupons.used_at')}:</strong> {formatDate(selectedCoupon.used_at)}</p>
                                    <p><strong>{t('coupons.purchase_email')}:</strong> {selectedCoupon.purchase_email || '-'}</p>
                                </Col>
                            </Row>
                        </div>
                    )}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setIsDetailModalOpen(false)}>
                        {t('close')}
                    </Button>
                </ModalFooter>
            </Modal>
        </Fragment>
    );
};

export default ActivatedCoupons; 