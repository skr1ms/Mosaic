import React, { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardBody, CardHeader, CardTitle, Row, Col, Badge } from 'reactstrap';
import api from '../../../api/api';

const PartnerProfile = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [profile, setProfile] = useState(null);

  const fetchProfile = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      // Бэкенд: GET /partner/profile
      const res = await api.get('/partner/profile');
      setProfile(res.data || null);
    } catch (e) {
      setError(e?.response?.data?.error || e.message || t('profile.failed_to_load_profile'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => { fetchProfile(); }, [fetchProfile]);

  return (
    <div>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-user icon-gradient bg-mean-fruit"></i>
            </div>
            <div>
              {t('profile.my_profile')}
              <div className="page-title-subheading">{t('profile.profile_readonly_info')}</div>
            </div>
          </div>
        </div>
      </div>

      {!!error && (
        <div className="alert alert-danger" role="alert">{error}</div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{t('profile.my_data')}</CardTitle>
        </CardHeader>
        <CardBody>
          {loading ? (
            <div className="text-muted">{t('profile.loading')}</div>
          ) : profile ? (
            <Row>
              <Col md="6" className="mb-3">
                <div><strong>{t('profile.brand')}:</strong> {profile.brand_name}</div>
                <div><strong>{t('profile.domain')}:</strong> {profile.domain}</div>
                <div><strong>{t('profile.email')}:</strong> {profile.email}</div>
                <div><strong>{t('profile.phone')}:</strong> {profile.phone || '-'}</div>
                <div><strong>{t('profile.address')}:</strong> {profile.address || '-'}</div>
              </Col>
              <Col md="6" className="mb-3">
                <div><strong>{t('profile.telegram')}:</strong> {profile.telegram || '-'}</div>
                <div><strong>{t('profile.whatsapp')}:</strong> {profile.whatsapp || '-'}</div>
                <div><strong>{t('profile.status')}:</strong> {profile.status === 'active' ? <Badge color="success">active</Badge> : <Badge color="danger">blocked</Badge>}</div>
                <div><strong>{t('profile.created')}:</strong> {new Date(profile.created_at).toLocaleString('ru-RU', { hour12: false })}</div>
                <div><strong>{t('profile.updated')}:</strong> {new Date(profile.updated_at).toLocaleString('ru-RU', { hour12: false })}</div>
              </Col>
              {Array.isArray(profile.brand_colors) && profile.brand_colors.length > 0 && (
                <Col md="12" className="mt-2">
                  <div style={{ fontWeight: 600, marginBottom: 8 }}>{t('profile.brand_palette')}:</div>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
                    {profile.brand_colors.map((c, idx) => (
                      <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <span style={{ display: 'inline-block', width: 24, height: 24, borderRadius: 4, backgroundColor: c, border: '1px solid rgba(0,0,0,0.1)' }} />
                        <code>{c}</code>
                      </div>
                    ))}
                  </div>
                </Col>
              )}
            </Row>
          ) : (
            <div className="text-muted">{t('profile.no_data')}</div>
          )}
        </CardBody>
      </Card>
    </div>
  );
};

export default PartnerProfile;


