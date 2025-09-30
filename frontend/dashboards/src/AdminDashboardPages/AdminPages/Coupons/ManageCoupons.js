import React, { Fragment, useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText, Table, Badge,
    Dropdown, DropdownToggle, DropdownMenu, DropdownItem,
    Modal, ModalHeader, ModalBody, ModalFooter, UncontrolledTooltip, Alert
} from 'reactstrap';
import api from '../../../api/api';

const ManageCoupons = () => {
    const { t } = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [partnerFilter, setPartnerFilter] = useState('all');
    const [sizeFilter, setSizeFilter] = useState('all');
    const [styleFilter, setStyleFilter] = useState('all');
    const [createdFrom, setCreatedFrom] = useState('');
    const [createdTo, setCreatedTo] = useState('');
    const [usedFrom, setUsedFrom] = useState('');
    const [usedTo, setUsedTo] = useState('');

    const applyDateTimeMask = (raw) => {
        const digits = (raw || '').replace(/\D/g, '').slice(0, 12); // dd mm yyyy hh mm → 12 цифр
        let out = '';
        for (let i = 0; i < digits.length; i += 1) {
            out += digits[i];
            if (i === 1 || i === 3) out += '.';        // после DD и MM
            if (i === 7) out += ' ';                   // после YYYY
            if (i === 9) out += ':';                   // после HH
        }
        return out;
    };

    const parseRuDateTime = (value) => {
        if (!value) return null;
        const match = value.match(/^(\d{2})\.(\d{2})\.(\d{4})\s+(\d{2}):(\d{2})$/);
        if (!match) return null;
        const [, dd, mm, yyyy, HH, MM] = match;
        const date = new Date(Number(yyyy), Number(mm) - 1, Number(dd), Number(HH), Number(MM));
        return isNaN(date.getTime()) ? null : date;
    };
    const [isStatusFilterOpen, setIsStatusFilterOpen] = useState(false);
    const [isPartnerFilterOpen, setIsPartnerFilterOpen] = useState(false);
    const [isSizeFilterOpen, setIsSizeFilterOpen] = useState(false);
    const [isStyleFilterOpen, setIsStyleFilterOpen] = useState(false);
    const [selectedCoupon, setSelectedCoupon] = useState(null);
    const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);
    const [page, setPage] = useState(1);
    const [limit, setLimit] = useState(20);
    const [pagination, setPagination] = useState({ total: 0, total_pages: 0 });

    const [loading, setLoading] = useState(false);
    const [actionLoading, setActionLoading] = useState(false);

    const [successMessage, setSuccessMessage] = useState('');
    const [coupons, setCoupons] = useState([]);
    const [partners, setPartners] = useState([]);
    const [selectedIds, setSelectedIds] = useState([]);
    const [selectAll, setSelectAll] = useState(false);
    const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
    const [deleteTargetId, setDeleteTargetId] = useState(null);
    const [confirmBatchOpen, setConfirmBatchOpen] = useState(false);
    const [confirmBatchResetOpen, setConfirmBatchResetOpen] = useState(false);
    const [confirmResetOpen, setConfirmResetOpen] = useState(false);
    const [resetTargetId, setResetTargetId] = useState(null);

    // Export modal state
    const [isExportModalOpen, setIsExportModalOpen] = useState(false);
    const [exportFormat, setExportFormat] = useState('txt'); 
    const [exportScope, setExportScope] = useState('all'); 
    const [partnerSearch, setPartnerSearch] = useState('');
    const [selectedPartnerCodes, setSelectedPartnerCodes] = useState([]); // array of 4-char codes
    const [exportLoading, setExportLoading] = useState(false);

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

    const loadCoupons = useCallback(async () => {
        try {
            setLoading(true);
            const params = new URLSearchParams();
            
            if (searchTerm && searchTerm.trim()) {
                params.append('search', searchTerm.trim());
            }
            
            if (statusFilter === 'new' || statusFilter === 'activated' || statusFilter === 'used' || statusFilter === 'completed') {
                params.append('status', statusFilter);
            } else if (statusFilter === 'blocked') {
                params.append('is_blocked', 'true');
            }
            if (partnerFilter !== 'all') params.append('partner_id', partnerFilter);
            if (sizeFilter !== 'all') params.append('size', sizeFilter);
            if (styleFilter !== 'all') params.append('style', styleFilter);
            params.append('page', String(page));
            params.append('limit', String(limit));
            
            const cf = parseRuDateTime(createdFrom);
            const ct = parseRuDateTime(createdTo);
            const uf = parseRuDateTime(usedFrom);
            const ut = parseRuDateTime(usedTo);
            if (cf) params.append('created_from', cf.toISOString());
            if (ct) params.append('created_to', ct.toISOString());
            if (uf) params.append('used_from', uf.toISOString());
            if (ut) params.append('used_to', ut.toISOString());

            const res = await api.get(`/admin/coupons/paginated?${params.toString()}`);
            setCoupons(res.data?.coupons || []);
            if (res.data?.pagination) setPagination(res.data.pagination);
        } catch (e) {
            console.error('Ошибка загрузки купонов:', e);
        } finally {
            setLoading(false);
        }
    }, [searchTerm, statusFilter, partnerFilter, sizeFilter, styleFilter, page, limit, createdFrom, createdTo, usedFrom, usedTo]);

    useEffect(() => {
        const loadPartners = async () => {
            try {
                const res = await api.get('/admin/partners');
                const list = res.data?.partners || [];
                
                const filtered = list.filter(p => p?.partner_code !== '0000');
                setPartners([{ id: '0000', brand_name: t('coupons.own_coupons'), partner_code: '0000' }, ...filtered]);
            } catch (_) {}
        };
        loadPartners();
    }, [t]);

    useEffect(() => {
        loadCoupons();
    }, [loadCoupons]);

    
    const filteredCoupons = coupons;

    const getStatusBadge = (status) => {
        switch (status) {
            case 'new':
                return <Badge color="success">{t('coupons.status_new')}</Badge>;
            case 'activated':
                return <Badge color="info">{t('coupons.status_activated')}</Badge>;
            case 'used':
                return <Badge color="primary">{t('coupons.status_used')}</Badge>;
            case 'completed':
                return <Badge color="secondary">{t('coupons.status_completed')}</Badge>;
            default:
                return <Badge color="warning">{status}</Badge>;
        }
    };

    const formatDate = (dateString) => {
        if (!dateString) return '-';
        const date = new Date(dateString);
        const pad = (n) => String(n).padStart(2, '0');
        const day = pad(date.getDate());
        const month = pad(date.getMonth() + 1);
        const year = date.getFullYear();
        const hours = pad(date.getHours());
        const minutes = pad(date.getMinutes());
        return `${day}.${month}.${year} ${hours}:${minutes}`;
    };

    const handleViewDetails = (coupon) => {
        setSelectedCoupon(coupon);
        setIsDetailModalOpen(true);
    };

    const handleReset = async (couponId) => {
        try {
            setActionLoading(true);
            await api.patch(`/admin/coupons/${couponId}/reset`);
            await loadCoupons();
            setSuccessMessage('Купон успешно сброшен');
        } catch (e) {
            alert(e?.response?.data?.error || e.message || 'Не удалось сбросить купон');
        } finally {
            setActionLoading(false);
        }
    };

    const openResetModal = (couponId) => {
        setResetTargetId(couponId);
        setConfirmResetOpen(true);
    };

    const handleDelete = async (couponId) => {
        try {
            setActionLoading(true);
            await api.delete(`/admin/coupons/${couponId}?confirm=true`);
            
            setCoupons(prev => prev.filter(c => c.id !== couponId));
            setSelectedIds(prev => prev.filter(id => id !== couponId));
            await loadCoupons();
            setSuccessMessage('Купон удалён');
        } catch (e) {
            alert(e?.response?.data?.error || e.message || 'Не удалось удалить купон');
        } finally {
            setActionLoading(false);
        }
    };

    const openDeleteModal = (couponId) => {
        setDeleteTargetId(couponId);
        setConfirmDeleteOpen(true);
    };

    const toggleSelectOne = (id, checked) => {
        setSelectedIds(prev => {
            const set = new Set(prev);
            if (checked) set.add(id); else set.delete(id);
            return Array.from(set);
        });
    };

    const toggleSelectAll = (checked) => {
        setSelectAll(checked);
        if (checked) {
            setSelectedIds(filteredCoupons.map(c => c.id));
        } else {
            setSelectedIds([]);
        }
    };

    const batchDelete = async () => {
        if (selectedIds.length === 0) return;
        try {
            setActionLoading(true);
            await api.post('/admin/coupons/batch-delete', { coupon_ids: selectedIds, confirm: true });
            setSuccessMessage('Выбранные купоны удалены');
            setSelectedIds([]);
            setSelectAll(false);
            await loadCoupons();
        } catch (e) {
            alert(e?.response?.data?.error || e.message || 'Не удалось удалить выбранные купоны');
        } finally {
            setActionLoading(false);
        }
    };

    const openBatchDeleteModal = () => {
        if (selectedIds.length === 0) return;
        setConfirmBatchOpen(true);
    };

    const batchReset = async () => {
        if (selectedIds.length === 0) return;
        try {
            setActionLoading(true);
            await api.post('/admin/coupons/batch/reset', { coupon_ids: selectedIds });
            alert('Выбранные купоны сброшены');
            setSelectedIds([]);
            setSelectAll(false);
            await loadCoupons();
        } catch (e) {
            alert(e?.response?.data?.error || e.message || 'Не удалось сбросить выбранные купоны');
        } finally {
            setActionLoading(false);
        }
    };

    const openBatchResetModal = () => {
        if (selectedIds.length === 0) return;
        setConfirmBatchResetOpen(true);
    };

    const handleDownloadMaterials = async (coupon) => {
        try {
            
            if (coupon?.schemaUrl) {
                
                const link = document.createElement('a');
                link.href = coupon.schemaUrl;
                link.setAttribute('download', `schema_${coupon.code}.zip`);
                link.target = '_blank';
                document.body.appendChild(link);
                link.click();
                link.remove();
                return;
            }
            
            
            const response = await api.get(`/admin/coupons/${coupon.id}/download-materials`, { responseType: 'blob' });
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', `schema_${coupon.code}.zip`);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (e) {
            alert(e?.response?.data?.error || e.message || t('coupons.failed_to_download_materials'));
        }
    };

    const openExportModal = () => setIsExportModalOpen(true);
    const closeExportModal = () => {
        if (exportLoading) return;
        setIsExportModalOpen(false);
    };

    const togglePartnerCode = (code) => {
        setSelectedPartnerCodes(prev => {
            const set = new Set(prev);
            if (set.has(code)) set.delete(code); else set.add(code);
            return Array.from(set);
        });
    };

    const filteredPartnerCodes = useMemo(() => {
        const term = partnerSearch.trim().toLowerCase();
        const items = (partners || []).map(p => ({
            code: p.partner_code || (p.id === '0000' ? '0000' : ''),
            name: p.brand_name || '',
            id: p.id,
        })).filter(p => p.code);
        if (!term) return items;
        return items.filter(p => p.code.toLowerCase().includes(term) || p.name.toLowerCase().includes(term));
    }, [partners, partnerSearch]);

    const downloadBlob = (blob, filename) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const handleExport = async () => {
        try {
            setExportLoading(true);
            if (exportScope === 'all') {
                const req = {
                    format: 'admin',
                    file_format: exportFormat,
                    include_header: true,
                    delimiter: exportFormat === 'csv' ? ';' : ',',
                };
                const res = await api.post('/admin/coupons/export-advanced', req, { responseType: 'blob' });
                downloadBlob(res.data, `coupons_export_all_${Date.now()}.${exportFormat}`);
            } else {
                if (selectedPartnerCodes.length === 0) {
                    alert('Выберите хотя бы один партнёрский код');
                    return;
                }
                const req = {
                    format: 'admin',
                    file_format: exportFormat,
                    include_header: true,
                    delimiter: exportFormat === 'csv' ? ';' : ',',
                    partner_codes: selectedPartnerCodes,
                };
                const res = await api.post('/admin/coupons/export-advanced', req, { responseType: 'blob' });
                downloadBlob(res.data, `coupons_export_${selectedPartnerCodes.join('-')}_${Date.now()}.${exportFormat}`);
            }
            setIsExportModalOpen(false);
        } catch (e) {
            alert(e?.response?.data?.error || e.message || 'Не удалось экспортировать купоны');
        } finally {
            setExportLoading(false);
        }
    };

    return (
        <Fragment>
            {successMessage && (
                <Alert color="success" className="mb-3">{successMessage}</Alert>
            )}
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-search icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('coupons.manage_coupons_title')}
                            <div className="page-title-subheading">
                                {t('coupons.manage_coupons_description')}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="success" 
                            size="lg"
                            onClick={openExportModal}
                        >
                            <i className="pe-7s-download"></i> {t('coupons.export_list_button')}
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('coupons.filters_and_search')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row className="mb-3">
                                <Col md="4">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder={t('coupons.search_placeholder')}
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isStatusFilterOpen} 
                                        toggle={() => setIsStatusFilterOpen(!isStatusFilterOpen)}
                                    >
                                        <DropdownToggle caret color="info" block>
                                            {statusFilter === 'all' ? t('coupons.all_statuses') : 
                                             statusFilter === 'new' ? t('coupons.new_coupons') : 
                                             statusFilter === 'activated' ? t('coupons.activated_coupons') :
                                             statusFilter === 'used' ? t('coupons.used_coupons') : 
                                             statusFilter === 'completed' ? t('coupons.completed_coupons') :
                                             t('coupons.blocked_coupons')}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStatusFilter('all')}>
                                                {t('coupons.all_statuses')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('new')}>
                                                {t('coupons.new_coupons')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('activated')}>
                                                {t('coupons.activated_coupons')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('used')}>
                                                {t('coupons.used_coupons')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('completed')}>
                                                {t('coupons.completed_coupons')}
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('blocked')}>
                                                {t('coupons.blocked_coupons')}
                                            </DropdownItem>
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isPartnerFilterOpen} 
                                        toggle={() => setIsPartnerFilterOpen(!isPartnerFilterOpen)}
                                    >
                                        <DropdownToggle caret color="secondary" block>
                                            {partnerFilter === 'all' ? t('coupons.all_partners') : 
                                              (partners.find(p => p.id === partnerFilter)?.brand_name || '—')}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setPartnerFilter('all')}>
                                                {t('coupons.all_partners')}
                                            </DropdownItem>
                                             {partners.map(partner => (
                                                <DropdownItem 
                                                    key={partner.id}
                                                    onClick={() => setPartnerFilter(partner.id)}
                                                >
                                                    {partner.brand_name}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isSizeFilterOpen} 
                                        toggle={() => setIsSizeFilterOpen(!isSizeFilterOpen)}
                                    >
                                        <DropdownToggle caret color="warning" block>
                                            {sizeFilter === 'all' ? t('coupons.all_sizes') : 
                                             sizeOptions.find(s => s.value === sizeFilter)?.label}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setSizeFilter('all')}>
                                                {t('coupons.all_sizes')}
                                            </DropdownItem>
                                            {sizeOptions.map(size => (
                                                <DropdownItem 
                                                    key={size.value}
                                                    onClick={() => setSizeFilter(size.value)}
                                                >
                                                    {size.label}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isStyleFilterOpen} 
                                        toggle={() => setIsStyleFilterOpen(!isStyleFilterOpen)}
                                    >
                                        <DropdownToggle caret color="primary" block>
                                            {styleFilter === 'all' ? t('coupons.all_styles') : 
                                             styleOptions.find(s => s.value === styleFilter)?.label}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStyleFilter('all')}>
                                                {t('coupons.all_styles')}
                                            </DropdownItem>
                                            {styleOptions.map(style => (
                                                <DropdownItem 
                                                    key={style.value}
                                                    onClick={() => setStyleFilter(style.value)}
                                                >
                                                    {style.label}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                            </Row>
                            <Row className="mt-3">
                                <Col md="3">
                                    <label className="form-label">{t('coupons.created_from')}</label>
                                    <Input
                                        type="text"
                                        value={createdFrom}
                                        onChange={(e) => { setPage(1); setCreatedFrom(applyDateTimeMask(e.target.value)); }}
                                        placeholder={t('coupons.created_from_placeholder')}
                                    />
                                </Col>
                                <Col md="3">
                                    <label className="form-label">{t('coupons.created_to')}</label>
                                    <Input
                                        type="text"
                                        value={createdTo}
                                        onChange={(e) => { setPage(1); setCreatedTo(applyDateTimeMask(e.target.value)); }}
                                        placeholder={t('coupons.created_to_placeholder')}
                                    />
                                </Col>
                                <Col md="3">
                                    <label className="form-label">{t('coupons.activated_from')}</label>
                                    <Input
                                        type="text"
                                        value={usedFrom}
                                        onChange={(e) => { setPage(1); setUsedFrom(applyDateTimeMask(e.target.value)); }}
                                        placeholder={t('coupons.activated_from_placeholder')}
                                    />
                                </Col>
                                <Col md="3">
                                    <label className="form-label">{t('coupons.activated_to')}</label>
                                    <Input
                                        type="text"
                                        value={usedTo}
                                        onChange={(e) => { setPage(1); setUsedTo(applyDateTimeMask(e.target.value)); }}
                                        placeholder={t('coupons.activated_to_placeholder')}
                                    />
                                </Col>
                            </Row>
                            <Row>
                                <Col md="12">
                                    <div className="d-flex justify-content-between align-items-center">
                                        <small className="text-muted">
                                            {t('coupons.found_coupons')}: <strong>{filteredCoupons.length}</strong>
                                        </small>
                                        <Button 
                                            color="primary" 
                                            size="sm"
                                            className="mt-1"
                                            onClick={() => {
                                                setSearchTerm('');
                                                setStatusFilter('all');
                                                setPartnerFilter('all');
                                                setSizeFilter('all');
                                                setStyleFilter('all');
                                            }}
                                        >
                                            <i className="pe-7s-refresh"></i> {t('coupons.reset_filters')}
                                        </Button>
                                    </div>
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
                                <CardTitle className="mb-0">{t('coupons.coupon_list')}</CardTitle>
                                <div className="btn-group">
                                    <Button color="danger" size="sm" disabled={selectedIds.length === 0 || actionLoading} onClick={openBatchDeleteModal}>
                                        <i className="pe-7s-trash"></i> {t('coupons.delete_selected')}
                                    </Button>
                                    <Button color="warning" size="sm" disabled={selectedIds.length === 0 || actionLoading} onClick={openBatchResetModal}>
                                        <i className="pe-7s-refresh"></i> {t('coupons.reset_selected')}
                                    </Button>
                                </div>
                            </div>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th style={{width:'36px'}}>
                                            <input type="checkbox" checked={selectAll} onChange={(e) => toggleSelectAll(e.target.checked)} />
                                        </th>
                                        <th>{t('coupons.coupon_number')}</th>
                                        <th>{t('coupons.partner')}</th>
                                        <th>{t('coupons.size')}</th>
                                        <th>{t('coupons.style')}</th>
                                        <th>{t('coupons.status')}</th>
                                        <th>{t('coupons.created_at')}</th>
                                        <th>{t('coupons.activated_at')}</th>
                                        <th>{t('coupons.is_blocked')}</th>
                                        <th>{t('coupons.actions')}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {loading ? (
                                        <tr><td colSpan="9" className="text-center text-muted">Загрузка...</td></tr>
                                    ) : filteredCoupons.map((coupon) => (
                                        <tr key={coupon.id}>
                                            <td>
                                                <input type="checkbox" checked={selectedIds.includes(coupon.id)} onChange={(e) => toggleSelectOne(coupon.id, e.target.checked)} />
                                            </td>
                                            <td>
                                                <Badge color="secondary" className="badge-pill">
                                                    {coupon.code}
                                                </Badge>
                                            </td>
                                            <td>
                                                {(() => {
                                                    const isOwn = !coupon.partner_id || coupon.partner_id === '00000000-0000-0000-0000-000000000000';
                                                    const partnerObj = isOwn ? { partner_code: '0000', brand_name: t('coupons.own_coupons') } : partners.find(p => p.id === coupon.partner_id);
                                                    const partnerName = coupon.partner_name || partnerObj?.brand_name || '—';
                                                    const partnerCode = partnerObj?.partner_code || '0000';
                                                    return (
                                                        <>
                                                            <strong>{partnerName}</strong>
                                                            <br/>
                                                            <small className="text-muted">
                                                                {t('coupons.partner_code')}: {partnerCode}
                                                            </small>
                                                        </>
                                                    );
                                                })()}
                                            </td>
                                            <td>{sizeOptions.find(s => s.value === coupon.size)?.label}</td>
                                            <td>{styleOptions.find(s => s.value === coupon.style)?.label}</td>
                                            <td>{getStatusBadge(coupon.status)}</td>
                                            <td>
                                                <small>{formatDate(coupon.created_at)}</small>
                                            </td>
                                            <td>
                                                <small>{formatDate(coupon.used_at)}</small>
                                            </td>
                                            <td>
                                                {(() => {
                                                    const partnerObj = partners.find(p => p.id === coupon.partner_id);
                                                    const partnerBlocked = partnerObj?.status === 'blocked';
                                                    const blocked = Boolean(coupon.is_blocked) || Boolean(partnerBlocked);
                                                    return blocked ? (
                                                    <Badge color="danger">{t('common.yes')}</Badge>
                                                ) : (
                                                    <Badge color="success">{t('common.no')}</Badge>
                                                );
                                                })()}
                                            </td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="info"
                                                        onClick={() => handleViewDetails(coupon)}
                                                        id={`details-${coupon.id}`}
                                                    >
                                                        <i className="pe-7s-look"></i>
                                                    </Button>
                                                    <UncontrolledTooltip target={`details-${coupon.id}`}>
                                                        {t('coupons.view_details')}
                                                    </UncontrolledTooltip>

                                                    {coupon.status === 'used' && (
                                                        <>
                                                            <Button 
                                                                size="sm" 
                                                                color="success"
                                                                onClick={() => handleDownloadMaterials(coupon)}
                                                                id={`download-${coupon.id}`}
                                                            >
                                                                <i className="pe-7s-download"></i>
                                                            </Button>
                                                            <UncontrolledTooltip target={`download-${coupon.id}`}>
                                                                {t('coupons.download_materials')}
                                                            </UncontrolledTooltip>
                                                        </>
                                                    )}

                                                    <Button 
                                                        size="sm" 
                                                        color="warning"
                                                        onClick={() => openResetModal(coupon.id)}
                                                        id={`reset-${coupon.id}`}
                                                    >
                                                        <i className="pe-7s-refresh"></i>
                                                    </Button>
                                                    <UncontrolledTooltip target={`reset-${coupon.id}`}>
                                                        {t('coupons.reset_coupon')}
                                                    </UncontrolledTooltip>

                                                    <Button 
                                                        size="sm" 
                                                        color="danger"
                                                        onClick={() => openDeleteModal(coupon.id)}
                                                        id={`delete-${coupon.id}`}
                                                    >
                                                        <i className="pe-7s-trash"></i>
                                                    </Button>
                                                    <UncontrolledTooltip target={`delete-${coupon.id}`}>
                                                        {t('coupons.delete_coupon')}
                                                    </UncontrolledTooltip>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>

                            <div className="d-flex justify-content-between align-items-center mt-3">
                                <div>
                                    <small className="text-muted">{t('coupons.total')}: {pagination.total || filteredCoupons.length}</small>
                                </div>
                                <div className="d-flex align-items-center">
                                    <div className="mr-2">
                                        <select className="form-select form-select-sm" value={limit} onChange={(e) => { setPage(1); setLimit(Number(e.target.value)); }}>
                                            <option value={10}>10</option>
                                            <option value={20}>20</option>
                                            <option value={50}>50</option>
                                            <option value={100}>100</option>
                                        </select>
                                    </div>
                                    <div className="btn-group">
                                        <Button size="sm" color="light" disabled={page <= 1} onClick={() => setPage(p => Math.max(1, p - 1))}>
                                            <i className="pe-7s-angle-left" />
                                        </Button>
                                        <Button size="sm" color="light" disabled={!pagination.has_next && (pagination.total_pages ? page >= pagination.total_pages : true)} onClick={() => setPage(p => p + 1)}>
                                            <i className="pe-7s-angle-right" />
                                        </Button>
                                    </div>
                                    <div className="ml-2">
                                        <small className="text-muted">{t('coupons.page')}: {page}{pagination.total_pages ? ` / ${pagination.total_pages}` : ''}</small>
                                    </div>
                                </div>
                            </div>

                            {filteredCoupons.length === 0 && (
                                <div className="text-center py-4">
                                    <p className="text-muted">{t('coupons.no_coupons_found')}</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Modal isOpen={isDetailModalOpen} toggle={() => setIsDetailModalOpen(false)} size="lg">
                <ModalHeader toggle={() => setIsDetailModalOpen(false)}>{t('coupons.coupon_details')}: {selectedCoupon?.code}</ModalHeader>
                <ModalBody>
                    {selectedCoupon && (
                        <Row>
                            <Col md="6">
                                <h6>{t('coupons.main_info')}</h6>
                                <table className="table table-sm">
                                    <tbody>
                                        <tr>
                                            <td><strong>{t('coupons.coupon_number')}:</strong></td>
                                            <td>{selectedCoupon.code}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>{t('coupons.partner')}:</strong></td>
                                            <td>
                                                {(() => {
                                                    const isOwn = !selectedCoupon.partner_id || selectedCoupon.partner_id === '00000000-0000-0000-0000-000000000000';
                                                    const partnerObj = isOwn ? { partner_code: '0000', brand_name: t('coupons.own_coupons') } : partners.find(p => p.id === selectedCoupon.partner_id);
                                                    const partnerName = selectedCoupon.partner_name || partnerObj?.brand_name || '—';
                                                    const partnerCode = partnerObj?.partner_code || '0000';
                                                    return (
                                                        <>
                                                            <strong>{partnerName}</strong>
                                                            <br/>
                                                            <small className="text-muted">{t('coupons.partner_code')}: {partnerCode}</small>
                                                        </>
                                                    );
                                                })()}
                                            </td>
                                        </tr>
                                        <tr>
                                            <td><strong>{t('coupons.size')}:</strong></td>
                                            <td>{sizeOptions.find(s => s.value === selectedCoupon.size)?.label}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>{t('coupons.style')}:</strong></td>
                                            <td>{styleOptions.find(s => s.value === selectedCoupon.style)?.label}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>{t('coupons.status')}:</strong></td>
                                            <td>{getStatusBadge(selectedCoupon.status)}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>{t('coupons.created_at')}:</strong></td>
                                            <td>{formatDate(selectedCoupon.created_at)}</td>
                                        </tr>
                                        {selectedCoupon.status === 'used' && (
                                            <tr>
                                                <td><strong>{t('coupons.activated_at')}:</strong></td>
                                                <td>{formatDate(selectedCoupon.used_at)}</td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            </Col>
                            <Col md="6">
                                <div className="text-center text-muted py-4">
                                    <i className="pe-7s-info" style={{fontSize: '2em'}}></i>
                                    <p>{t('coupons.coupon_details_info')}</p>
                                </div>
                            </Col>
                        </Row>
                    )}
                </ModalBody>
                <ModalFooter>
                    {(selectedCoupon?.status === 'used' || selectedCoupon?.status === 'completed') && (
                        <Button
                            color="success"
                            onClick={() => handleDownloadMaterials(selectedCoupon)}
                        >
                            <i className="pe-7s-download"/> {t('coupons.download_materials')}
                        </Button>
                    )}
                    <Button color="secondary" onClick={() => setIsDetailModalOpen(false)}>{t('coupons.close')}</Button>
                </ModalFooter>
            </Modal>

            {}
            <Modal isOpen={isExportModalOpen} toggle={closeExportModal}>
                <ModalHeader toggle={closeExportModal}>{t('coupons.export_coupons')}</ModalHeader>
                <ModalBody>
                    <div className="mb-3">
                        <label className="form-label"><strong>{t('coupons.file_format')}</strong></label>
                        <div className="d-flex gap-3">
                            <label className="form-check-label">
                                <input
                                    type="radio"
                                    className="form-check-input"
                                    name="exportFormat"
                                    value="txt"
                                    checked={exportFormat === 'txt'}
                                    onChange={() => setExportFormat('txt')}
                                /> {t('coupons.txt_format')}
                            </label>
                            <label className="form-check-label">
                                <input
                                    type="radio"
                                    className="form-check-input"
                                    name="exportFormat"
                                    value="csv"
                                    checked={exportFormat === 'csv'}
                                    onChange={() => setExportFormat('csv')}
                                /> {t('coupons.csv_format')}
                            </label>
                        </div>
                    </div>

                    <div className="mb-3">
                        <label className="form-label"><strong>{t('coupons.export_scope')}</strong></label>
                        <div className="d-flex flex-column gap-2">
                            <label className="form-check-label">
                                <input
                                    type="radio"
                                    className="form-check-input"
                                    name="exportScope"
                                    value="all"
                                    checked={exportScope === 'all'}
                                    onChange={() => setExportScope('all')}
                                /> {t('coupons.export_all_coupons')}
                            </label>
                            <label className="form-check-label">
                                <input
                                    type="radio"
                                    className="form-check-input"
                                    name="exportScope"
                                    value="by_partner"
                                    checked={exportScope === 'by_partner'}
                                    onChange={() => setExportScope('by_partner')}
                                /> {t('coupons.export_by_partner_code')}
                            </label>
                        </div>
                    </div>

                    {exportScope === 'by_partner' && (
                        <div className="mb-3">
                            <label className="form-label">{t('coupons.search_partner_by_name_or_code')}</label>
                            <Input
                                type="text"
                                value={partnerSearch}
                                onChange={(e) => setPartnerSearch(e.target.value)}
                                placeholder={t('coupons.search_partner_placeholder')}
                            />
                            <div className="mt-2" style={{ maxHeight: 240, overflowY: 'auto', border: '1px solid #eee', borderRadius: 4, padding: 8 }}>
                                {filteredPartnerCodes.length === 0 ? (
                                    <div className="text-muted">{t('coupons.nothing_found')}</div>
                                ) : (
                                    filteredPartnerCodes.map(p => (
                                        <div key={`${p.id}-${p.code}`} className="form-check">
                                            <input
                                                className="form-check-input"
                                                type="checkbox"
                                                id={`code-${p.code}`}
                                                checked={selectedPartnerCodes.includes(p.code)}
                                                onChange={() => togglePartnerCode(p.code)}
                                            />
                                            <label className="form-check-label" htmlFor={`code-${p.code}`}>
                                                <strong>{p.code}</strong> — {p.name}
                                            </label>
                                        </div>
                                    ))
                                )}
                            </div>
                            <small className="text-muted">{t('coupons.can_select_multiple_codes')}</small>
                        </div>
                    )}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={closeExportModal} disabled={exportLoading}>{t('coupons.cancel')}</Button>
                    <Button color="success" onClick={handleExport} disabled={exportLoading}>
                        {exportLoading ? (<><i className="fa fa-spinner fa-spin"></i> {t('coupons.exporting')}</>) : (<>{t('coupons.export')}</>)}
                    </Button>
                </ModalFooter>
            </Modal>

            {}
            <Modal isOpen={confirmDeleteOpen} toggle={() => setConfirmDeleteOpen(false)}>
                <ModalHeader toggle={() => setConfirmDeleteOpen(false)}>{t('coupons.delete_coupon')}</ModalHeader>
                <ModalBody>
                    {t('coupons.delete_coupon_confirmation')}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setConfirmDeleteOpen(false)}>{t('coupons.cancel')}</Button>
                    <Button color="danger" onClick={() => { const id = deleteTargetId; setConfirmDeleteOpen(false); if (id) handleDelete(id); }}>{t('coupons.delete')}</Button>
                </ModalFooter>
            </Modal>

            {}
            <Modal isOpen={confirmBatchOpen} toggle={() => setConfirmBatchOpen(false)}>
                <ModalHeader toggle={() => setConfirmBatchOpen(false)}>{t('coupons.delete_selected')}</ModalHeader>
                <ModalBody>
                    {t('coupons.selected_coupons_to_delete')}: {selectedIds.length}. {t('coupons.delete_selected_confirmation')}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setConfirmBatchOpen(false)}>{t('coupons.cancel')}</Button>
                    <Button color="danger" onClick={() => { setConfirmBatchOpen(false); batchDelete(); }}>{t('coupons.delete')}</Button>
                </ModalFooter>
            </Modal>

            {}
            <Modal isOpen={confirmBatchResetOpen} toggle={() => setConfirmBatchResetOpen(false)}>
                <ModalHeader toggle={() => setConfirmBatchResetOpen(false)}>{t('coupons.reset_selected')}</ModalHeader>
                <ModalBody>
                    {t('coupons.selected_coupons_to_reset')}: {selectedIds.length}. {t('coupons.reset_selected_confirmation')}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setConfirmBatchResetOpen(false)}>{t('coupons.cancel')}</Button>
                    <Button color="warning" onClick={() => { setConfirmBatchResetOpen(false); batchReset(); }}>{t('coupons.reset')}</Button>
                </ModalFooter>
            </Modal>

            {}
            <Modal isOpen={confirmResetOpen} toggle={() => setConfirmResetOpen(false)}>
                <ModalHeader toggle={() => setConfirmResetOpen(false)}>{t('coupons.reset_coupon')}</ModalHeader>
                <ModalBody>
                    {t('coupons.reset_coupon_confirmation')}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setConfirmResetOpen(false)}>{t('coupons.cancel')}</Button>
                    <Button color="warning" onClick={() => { const id = resetTargetId; setConfirmResetOpen(false); if (id) handleReset(id); }}>{t('coupons.reset')}</Button>
                </ModalFooter>
            </Modal>
        </Fragment>
    );
};

export default ManageCoupons; 