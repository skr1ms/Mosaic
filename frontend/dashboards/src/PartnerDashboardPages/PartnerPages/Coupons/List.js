import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardBody, CardHeader, CardTitle, Row, Col, Input, Button, Table, Badge } from 'reactstrap';
import api from '../../../api/api';

const PartnerCoupons = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [coupons, setCoupons] = useState([]);
  const [filters, setFilters] = useState({
    code: '',
    status: '',
    size: '',
    style: '',
  });
  const [sortBy, setSortBy] = useState('created_at'); 
  const [sortOrder, setSortOrder] = useState('desc'); 

  const sizeOptions = [
    { value: '', label: t('coupons.size_options.all') },
    { value: '21x30', label: t('coupons.size_options.21x30') },
    { value: '30x40', label: t('coupons.size_options.30x40') },
    { value: '40x40', label: t('coupons.size_options.40x40') },
    { value: '40x50', label: t('coupons.size_options.40x50') },
    { value: '40x60', label: t('coupons.size_options.40x60') },
    { value: '50x70', label: t('coupons.size_options.50x70') },
  ];
  const styleOptions = [
    { value: '', label: t('coupons.style_options.all') },
    { value: 'grayscale', label: t('coupons.style_options.grayscale') },
    { value: 'skin_tones', label: t('coupons.style_options.skin_tones') },
    { value: 'pop_art', label: t('coupons.style_options.pop_art') },
    { value: 'max_colors', label: t('coupons.style_options.max_colors') },
  ];

  const fetchCoupons = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      // Бэкенд: GET /partner/coupons с фильтрами и сортировкой
      const params = new URLSearchParams();
      if (filters.code) params.append('code', filters.code.trim());
      if (filters.status) params.append('status', filters.status);
      if (filters.size) params.append('size', filters.size);
      if (filters.style) params.append('style', filters.style);
      if (sortBy) params.append('sort_by', sortBy);
      if (sortOrder) params.append('sort_order', sortOrder);
      const res = await api.get(`/partner/coupons${params.toString() ? `?${params.toString()}` : ''}`);
      const items = res.data?.coupons || [];
      setCoupons(items);
    } catch (e) {
      setError(e?.response?.data?.error || e.message || 'Не удалось загрузить купоны');
    } finally {
      setLoading(false);
    }
  }, [filters.code, filters.status, filters.size, filters.style, sortBy, sortOrder]);

  useEffect(() => {
    fetchCoupons();
  }, [fetchCoupons]);

  const filtered = useMemo(() => coupons, [coupons]);

  const formatDateTime = (iso) => {
    try {
      if (!iso) return '-';
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) return '-';
      return d.toLocaleString('ru-RU', { hour12: false });
    } catch {
      return '-';
    }
  };

  const handleDownloadMaterials = async (coupon) => {
    try {
      const res = await api.get(`/partner/coupons/${coupon.id}/download-materials`, { responseType: 'blob' });
      
      
      const blob = new Blob([res.data], { type: 'application/zip' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `coupon_${coupon.code}_materials.zip`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('coupons.failed_to_download_materials'));
    }
  };

  return (
    <div>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-ticket icon-gradient bg-mean-fruit"></i>
            </div>
            <div>
              {t('partner_dashboard.my_coupons')}
              <div className="page-title-subheading">{t('partner_dashboard.my_coupons_subtitle')}</div>
            </div>
          </div>
          
        </div>
      </div>

      {!!error && (
        <div className="alert alert-danger" role="alert">{error}</div>
      )}

      <Card className="mb-3">
        <CardHeader>
          <CardTitle>{t('partner_dashboard.filters')}</CardTitle>
        </CardHeader>
        <CardBody>
          <Row>
            <Col md="3" className="mb-2">
              <Input
                placeholder={t('partner_dashboard.search_by_coupon_number_placeholder')}
                value={filters.code}
                onChange={(e) => setFilters({ ...filters, code: e.target.value })}
              />
            </Col>
            <Col md="3" className="mb-2">
              <Input
                type="select"
                value={filters.status}
                onChange={(e) => setFilters({ ...filters, status: e.target.value })}
              >
                <option value="">{t('partner_dashboard.status_all_option')}</option>
                <option value="new">{t('partner_dashboard.status_new')}</option>
                <option value="activated">{t('partner_dashboard.status_activated')}</option>
                <option value="used">{t('partner_dashboard.status_used')}</option>
                <option value="completed">{t('partner_dashboard.status_completed')}</option>
              </Input>
            </Col>
            <Col md="3" className="mb-2">
              <Input
                type="select"
                value={filters.size}
                onChange={(e) => setFilters({ ...filters, size: e.target.value })}
              >
                {sizeOptions.map((s) => (
                  <option key={s.value || 'all'} value={s.value}>{s.label}</option>
                ))}
              </Input>
            </Col>
            <Col md="3" className="mb-2">
              <Input
                type="select"
                value={filters.style}
                onChange={(e) => setFilters({ ...filters, style: e.target.value })}
              >
                {styleOptions.map((s) => (
                  <option key={s.value || 'all'} value={s.value}>{s.label}</option>
                ))}
              </Input>
            </Col>
          </Row>
          <Row className="mt-2">
            <Col md="3" className="mb-2">
              <Input type="select" value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
                <option value="created_at">{t('partner_dashboard.sort_by_created')}</option>
                <option value="used_at">{t('partner_dashboard.sort_by_activated')}</option>
                <option value="code">{t('partner_dashboard.sort_by_code')}</option>
                <option value="status">{t('partner_dashboard.sort_by_status')}</option>
              </Input>
            </Col>
            <Col md="3" className="mb-2">
              <Input type="select" value={sortOrder} onChange={(e) => setSortOrder(e.target.value)}>
                <option value="desc">{t('partner_dashboard.sort_desc')}</option>
                <option value="asc">{t('partner_dashboard.sort_asc')}</option>
              </Input>
            </Col>
            <Col md="3" className="mb-2">
              <Button color="primary" onClick={fetchCoupons} disabled={loading}>
                <i className="pe-7s-refresh"></i> {t('partner_dashboard.refresh')}
              </Button>
            </Col>
          </Row>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('partner_dashboard.coupon_list')}</CardTitle>
        </CardHeader>
        <CardBody>
          <Table responsive size="sm">
            <thead>
              <tr>
                <th>{t('partner_dashboard.table_headers.code')}</th>
                <th>{t('partner_dashboard.table_headers.size')}</th>
                <th>{t('partner_dashboard.table_headers.style')}</th>
                <th>{t('partner_dashboard.table_headers.status')}</th>
                <th>{t('partner_dashboard.table_headers.created')}</th>
                <th>{t('partner_dashboard.table_headers.activated')}</th>
                <th>{t('partner_dashboard.table_headers.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((c, idx) => (
                <tr key={idx}>
                  <td><code>{c.code}</code></td>
                  <td>{c.size || '-'}</td>
                  <td>{(styleOptions.find(s => s.value === (c.style || ''))?.label) || '-'}</td>
                  <td>
                    {(() => {
                      const st = String(c.status || '').toLowerCase();
                      if (st === 'used') return <Badge color="success" className="fw-bold">{t('partner_dashboard.status_badges.used')}</Badge>;
                      if (st === 'activated') return <Badge color="primary" className="fw-bold">{t('partner_dashboard.status_badges.activated')}</Badge>;
                      if (st === 'completed') return <Badge color="warning" className="fw-bold">{t('partner_dashboard.status_badges.completed')}</Badge>;
                      return <Badge color="info" className="fw-bold">{t('partner_dashboard.status_badges.new')}</Badge>;
                    })()}
                  </td>
                  <td>{formatDateTime(c.created_at)}</td>
                  <td>{formatDateTime(c.used_at)}</td>
                  <td>
                    {String(c.status || '').toLowerCase() === 'completed' && (
                      <Button
                        size="sm"
                        color="success"
                        onClick={() => handleDownloadMaterials(c)}
                        title={t('coupons.download_materials')}
                      >
                        <i className="pe-7s-download"></i>
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
              {!loading && filtered.length === 0 && (
                <tr>
                  <td colSpan="7" className="text-center text-muted">{t('partner_dashboard.no_data')}</td>
                </tr>
              )}
            </tbody>
          </Table>
        </CardBody>
      </Card>
    </div>
  );
};

export default PartnerCoupons;


