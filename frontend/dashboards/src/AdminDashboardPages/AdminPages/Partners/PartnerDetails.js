import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Row, Col, Card, CardBody, CardHeader, CardTitle,
  Badge, Button, Table
} from 'reactstrap';
import api from '../../../api/api';

const PartnerDetails = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [partner, setPartner] = useState(null);
  const [stats, setStats] = useState({ total: 0, used: 0, purchased: 0, new: 0 });

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const [pRes, sRes] = await Promise.all([
        api.get(`/admin/partners/${id}`),
        api.get(`/admin/partners/${id}/statistics`),
      ]);
      setPartner(pRes.data || {});
      const st = (sRes.data?.coupon_statistics) || {};
      setStats({
        total: st.total || 0,
        used: st.used || 0,
        purchased: st.purchased || 0,
        new: st.new || 0,
      });
    } catch (e) {
      setError(e?.response?.data?.error || e.message || t('chat.failed_to_load_partner_data'));
    } finally {
      setLoading(false);
    }
  }, [id, t]);

  useEffect(() => { load(); }, [load]);

  const activationRate = useMemo(() => {
    return stats.total > 0 ? Math.round((stats.used / stats.total) * 1000) / 10 : 0;
  }, [stats]);

  const blockToggle = async () => {
    try {
      if (!partner) return;
      const active = partner.status === 'active';
      if (active) await api.patch(`/admin/partners/${id}/block`); else await api.patch(`/admin/partners/${id}/unblock`);
      await load();
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('chat.failed_to_change_status'));
    }
  };

  const deletePartner = async () => {
    if (!window.confirm(t('chat.confirm_delete_partner'))) return;
    try {
      await api.delete(`/admin/partners/${id}?confirm=true`);
      navigate('/partners/list');
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('chat.failed_to_delete_partner'));
    }
  };

  const exportPartnerCoupons = async () => {
    try {
      const res = await api.get(`/admin/coupons/export/partner/${id}?format=txt`, { responseType: 'blob' });
      const url = URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = url;
      a.download = `partner_${partner?.brand_name || 'coupons'}_${Date.now()}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('chat.failed_to_export_partner_coupons'));
    }
  };

  if (loading) {
    return (
      <div className="text-center py-5">
        <i className="fa fa-spinner fa-spin fa-3x" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="alert alert-danger" role="alert">{error}</div>
    );
  }

  return (
    <Fragment>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-user icon-gradient bg-mean-fruit" />
            </div>
            <div>
              {t('chat.partner_details_title').replace('{brand_name}', partner?.brand_name || '')}
              <div className="page-title-subheading">{t('chat.partner_details_subtitle')}</div>
            </div>
          </div>
          <div className="page-title-actions">
            <Button color="secondary" onClick={() => navigate('/partners/list')} className="mr-2">
              <i className="pe-7s-back" /> {t('chat.back')}
            </Button>
            <Button color={partner?.status === 'active' ? 'warning' : 'success'} onClick={blockToggle} className="mr-2">
              <i className={partner?.status === 'active' ? 'pe-7s-lock' : 'pe-7s-unlock'} /> {partner?.status === 'active' ? t('chat.block') : t('chat.unblock')}
            </Button>
            <Button color="danger" onClick={deletePartner}>
              <i className="pe-7s-trash" /> {t('chat.delete')}
            </Button>
          </div>
        </div>
      </div>

      <Row>
        <Col lg="8">
          <Card className="main-card mb-3">
            <CardHeader><CardTitle>{t('chat.main_information')}</CardTitle></CardHeader>
            <CardBody>
              <Table responsive size="sm">
                <tbody>
                  <tr><td><strong>{t('chat.partner_id')}</strong></td><td>{partner?.id}</td></tr>
                  <tr><td><strong>{t('chat.partner_code')}</strong></td><td><Badge color="secondary">{partner?.partner_code}</Badge></td></tr>
                  <tr><td><strong>{t('chat.login')}</strong></td><td>{partner?.login}</td></tr>
                  <tr><td><strong>{t('chat.domain')}</strong></td><td>{partner?.domain}</td></tr>
                  <tr><td><strong>{t('common.email')}</strong></td><td>{partner?.email}</td></tr>
                  <tr><td><strong>{t('chat.phone')}</strong></td><td>{partner?.phone || '-'}</td></tr>
                  <tr><td><strong>{t('chat.status')}</strong></td><td>{partner?.status === 'active' ? <Badge color="success">{t('chat.active')}</Badge> : <Badge color="danger">{t('chat.blocked')}</Badge>}</td></tr>
                  <tr><td><strong>{t('chat.created')}</strong></td><td>{partner?.created_at}</td></tr>
                  <tr><td><strong>{t('chat.updated')}</strong></td><td>{partner?.updated_at}</td></tr>
                  <tr><td><strong>OZON</strong></td><td>{partner?.ozon_link ? <a href={partner.ozon_link} target="_blank" rel="noreferrer">{t('chat.open')}</a> : '-'}</td></tr>
                  <tr><td><strong>Wildberries</strong></td><td>{partner?.wildberries_link ? <a href={partner.wildberries_link} target="_blank" rel="noreferrer">{t('chat.open')}</a> : '-'}</td></tr>
                </tbody>
              </Table>
              <div className="d-flex gap-2">
                <Button color="primary" className="mr-2" onClick={() => navigate(`/partners/edit/${id}`)}>
                  <i className="pe-7s-edit" /> {t('chat.edit')}
                </Button>
                <Button color="info" onClick={() => navigate(`/coupons/manage`)}>
                  <i className="pe-7s-search" /> {t('chat.to_coupons')}
                </Button>
              </div>
            </CardBody>
          </Card>
        </Col>

        <Col lg="4">
          <Card className="main-card mb-3">
            <CardHeader><CardTitle>{t('chat.coupon_statistics')}</CardTitle></CardHeader>
            <CardBody>
              <div className="mb-2">{t('chat.total')}: <strong>{stats.total}</strong></div>
              <div className="mb-2 text-success">{t('chat.activated')}: <strong>{stats.used}</strong></div>
              <div className="mb-2">{t('chat.new')}: <strong>{stats.new}</strong></div>
              <div className="mb-2 text-primary">{t('chat.purchased')}: <strong>{stats.purchased}</strong></div>
              <div className="mb-3">{t('chat.activation_rate')}: <strong>{activationRate}%</strong></div>
              <Button color="primary" block onClick={exportPartnerCoupons}>
                <i className="pe-7s-download" /> {t('chat.export_partner_coupons')}
              </Button>
            </CardBody>
          </Card>
        </Col>
      </Row>
    </Fragment>
  );
};

export default PartnerDetails;


