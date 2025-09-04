import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardBody, CardHeader, CardTitle, Row, Col, Input, Button } from 'reactstrap';
import api from '../../../api/api';

const PartnerExport = () => {
  const { t } = useTranslation();
  const [format, setFormat] = useState('txt');
  const [status, setStatus] = useState(''); // '' | new | activated | used | completed
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const exportCoupons = async () => {
    try {
      setLoading(true);
      setError('');
      // Бэкенд: GET /partner/coupons/export?format=txt|csv
      const params = new URLSearchParams();
      params.append('format', format);
      if (status) params.append('status', status);
      const res = await api.get(`/partner/coupons/export?${params.toString()}`, { responseType: 'blob' });
      const blob = new Blob([res.data], { type: format === 'csv' ? 'text/csv' : 'text/plain' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `partner_coupons.${format}`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (e) {
      setError(e?.response?.data?.error || e.message || t('partner_dashboard.failed_to_export'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-download icon-gradient bg-mean-fruit"></i>
            </div>
            <div>
              {t('partner_dashboard.export_coupons')}
              <div className="page-title-subheading">{t('partner_dashboard.export_coupons_subtitle')}</div>
            </div>
          </div>
        </div>
      </div>

      {!!error && (
        <div className="alert alert-danger" role="alert">{error}</div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('partner_dashboard.export_parameters')}</CardTitle>
        </CardHeader>
        <CardBody>
          <Row className="align-items-center">
            <Col md="3" className="mb-2">
              <Input type="select" value={format} onChange={(e) => setFormat(e.target.value)}>
                <option value="txt">TXT</option>
                <option value="csv">CSV</option>
              </Input>
            </Col>
            <Col md="4" className="mb-2">
              <Input type="select" value={status} onChange={(e) => setStatus(e.target.value)}>
                <option value="">{t('partner_dashboard.status_all')}</option>
                <option value="new">{t('partner_dashboard.status_new')}</option>
                <option value="activated">{t('partner_dashboard.status_activated')}</option>
                <option value="used">{t('partner_dashboard.status_used')}</option>
                <option value="completed">{t('partner_dashboard.status_completed')}</option>
              </Input>
            </Col>
            <Col md="3" className="mb-2">
              <Button color="primary" onClick={exportCoupons} disabled={loading}>
                <i className="pe-7s-download"></i> {t('partner_dashboard.export_button')}
              </Button>
            </Col>
          </Row>
        </CardBody>
      </Card>
    </div>
  );
};

export default PartnerExport;


