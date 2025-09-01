import React, { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardBody, CardHeader, CardTitle, Row, Col, Badge } from 'reactstrap';
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';
import api from '../../../api/api';

const Analytics = () => {
  const { t } = useTranslation();
  const [, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [stats, setStats] = useState({ total: 0, used: 0, new: 0, purchased: 0 });
  const [comparison, setComparison] = useState({ used: [], purchased: [] });

  const fetchStats = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      console.log('Fetching partner statistics...');
      
      // Бэкенд: GET /partner/statistics
      const res = await api.get('/partner/statistics');
      console.log('Statistics API response:', res.data);
      
      const payload = res.data || {};
      const statsData = {
        total: payload.total || payload.coupons?.total || 0,
        used: payload.used || payload.coupons?.used || 0,
        new: payload.new || payload.coupons?.new || 0,
        purchased: payload.purchased || payload.coupons?.purchased || 0,
      };
      
      console.log('Processed stats data:', statsData);
      setStats(statsData);
    } catch (e) {
      console.error('Error fetching statistics:', e);
      setError(e?.response?.data?.error || e.message || t('chat.failed_to_get_statistics'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  const fetchComparison = useCallback(async () => {
    try {
      console.log('Fetching comparison data...');
      const res = await api.get('/partner/statistics/comparison');
      console.log('Comparison API response:', res.data);
      
      const data = res.data || {};
      const normalize = (arr) => {
        console.log('Normalizing array:', arr);
        if (!Array.isArray(arr)) {
          console.log('Data is not an array, returning empty array');
          return [];
        }
        return arr.map((x, index) => {
          const normalized = {
            partner_id: x.partner_id || x.PartnerID || x.partnerId || `partner_${index}`,
            partner_code: x.partner_code || x.partnerCode || '',
            brand_name: x.brand_name || x.brandName || `Partner ${index + 1}`,
            count: Number(x.count || x.Count || 0),
          };
          console.log(`Normalized item ${index}:`, normalized);
          return normalized;
        });
      };
      
      const normalizedUsed = normalize(data.used);
      const normalizedPurchased = normalize(data.purchased);
      
      console.log('Normalized used data:', normalizedUsed);
      console.log('Normalized purchased data:', normalizedPurchased);
      
      setComparison({
        used: normalizedUsed,
        purchased: normalizedPurchased,
      });
    } catch (error) {
      console.error('Error fetching comparison data:', error);
      setError(`Failed to load comparison data: ${error.message}`);
    }
  }, []);

  useEffect(() => { fetchStats(); fetchComparison(); }, [fetchStats, fetchComparison]);

  return (
    <div>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-graph1 icon-gradient bg-mean-fruit"></i>
            </div>
                          <div>
                {t('chat.my_statistics')}
                <div className="page-title-subheading">{t('chat.my_statistics_subtitle')}</div>
              </div>
          </div>
        </div>
      </div>

      {!!error && (
        <div className="alert alert-danger" role="alert">{error}</div>
      )}



      <Row>
        <Col md="3">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('chat.total_coupons')}</CardTitle>
            </CardHeader>
            <CardBody>
              <h3>{stats.total}</h3>
            </CardBody>
          </Card>
        </Col>
        <Col md="3">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('chat.used')}</CardTitle>
            </CardHeader>
            <CardBody>
              <h3><Badge color="success">{stats.used}</Badge></h3>
            </CardBody>
          </Card>
        </Col>
        <Col md="3">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('chat.new_coupons')}</CardTitle>
            </CardHeader>
            <CardBody>
              <h3><Badge color="secondary">{stats.new}</Badge></h3>
            </CardBody>
          </Card>
        </Col>
        <Col md="3">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('chat.purchased_coupons')}</CardTitle>
            </CardHeader>
            <CardBody>
              <h3><Badge color="primary">{stats.purchased}</Badge></h3>
            </CardBody>
          </Card>
        </Col>
      </Row>

      <Row>
        <Col md="6">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('activation_comparison')}</CardTitle>
            </CardHeader>
            <CardBody>
              {console.log('Rendering activation comparison chart, data:', comparison.used)}
              {comparison.used.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={comparison.used}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="brand_name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="count" fill="#28a745" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <div className="text-center text-muted py-4">
                  <i className="pe-7s-graph1" style={{ fontSize: '48px', opacity: 0.3 }}></i>
                  <div className="mt-2">{t('no_comparison_data')}</div>
                  <small className="text-muted">Debug: {comparison.used.length} items</small>
                </div>
              )}
            </CardBody>
          </Card>
        </Col>

        <Col md="6">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('purchase_comparison')}</CardTitle>
            </CardHeader>
            <CardBody>
              {console.log('Rendering purchase comparison chart, data:', comparison.purchased)}
              {comparison.purchased.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={comparison.purchased}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="brand_name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="count" fill="#007bff" />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <div className="text-center text-muted py-4">
                  <i className="pe-7s-graph1" style={{ fontSize: '48px', opacity: 0.3 }}></i>
                  <div className="mt-2">{t('no_comparison_data')}</div>
                  <small className="text-muted">Debug: {comparison.purchased.length} items</small>
                </div>
              )}
            </CardBody>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Analytics;


