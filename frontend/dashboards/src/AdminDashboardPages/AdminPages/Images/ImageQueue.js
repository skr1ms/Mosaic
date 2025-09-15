import React, { Fragment, useEffect, useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Row, Col, Card, CardBody, CardHeader, CardTitle,
  Button, Input, InputGroup, InputGroupText, Table, Badge,
  Modal, ModalHeader, ModalBody, ModalFooter
} from 'reactstrap';
import api from '../../../api/api';

const ImageQueue = () => {
  const { t } = useTranslation();
  const [status, setStatus] = useState('all');
  const [search, setSearch] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [images, setImages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [page] = useState(1);
  const [limit] = useState(20);
  const [selected, setSelected] = useState(null);
  const [details, setDetails] = useState(null);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [detailsStatus, setDetailsStatus] = useState(null);

  // Функция для применения маски даты
  const applyDateMask = (raw) => {
    const digits = (raw || '').replace(/\D/g, '').slice(0, 8); // dd mm yyyy → 8 цифр
    let out = '';
    for (let i = 0; i < digits.length; i += 1) {
      out += digits[i];
      if (i === 1 || i === 3) out += '.'; // после DD и MM
    }
    return out;
  };

  const statusOptions = [
    { value: 'all', label: t('images.all_statuses') },
    { value: 'uploaded', label: 'uploaded' },
    { value: 'edited', label: 'edited' },
    { value: 'processing', label: 'processing' },
    { value: 'processed', label: 'processed' },
    { value: 'completed', label: 'completed' },
    { value: 'failed', label: 'failed' },
  ];

  const loadImages = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const params = new URLSearchParams();
      if (status !== 'all') params.append('status', status);
      
      
      if (dateFrom) {
        const [day, month, year] = dateFrom.split('.');
        if (day && month && year) {
          const apiDate = `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
          params.append('date_from', apiDate);
        }
      }
      if (dateTo) {
        const [day, month, year] = dateTo.split('.');
        if (day && month && year) {
          const apiDate = `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
          params.append('date_to', apiDate);
        }
      }
      
      const res = await api.get(`/admin/queue?${params.toString()}`);
      setImages(res.data || []);
    } catch (e) {
      setError(e?.response?.data?.error || e.message || t('images.failed_to_load_tasks'));
    } finally {
      setLoading(false);
    }
  }, [status, dateFrom, dateTo, t]);

  useEffect(() => { loadImages(); }, [status, page, limit, dateFrom, dateTo, loadImages]);

  const openDetails = async (img) => {
    setSelected(img);
    try {
      const res = await api.get(`/admin/queue/${img.id}`);
      setDetails(res.data || {});
      console.log('Task details:', res.data); 
    } catch (_) {
      setDetails(img);
    }
    try {
      const st = await fetchStatus(img.id);
      setDetailsStatus(st);
      console.log('Task status:', st); 
    } catch (_) {
      setDetailsStatus(null);
    }
    setDetailsOpen(true);
  };

  const computeProgress = (st) => {
    switch ((st || '').toLowerCase()) {
      case 'uploaded': return 20;
      case 'edited': return 40;
      case 'processing': return 60;
      case 'processed': return 80;
      case 'completed': return 100;
      case 'failed': default: return 0;
    }
  };

  const fetchStatus = async (id) => {
    try {
      
      const res = await api.get(`/admin/queue/${id}`);
      const taskData = res.data;
      
      
      return {
        status: taskData.status,
        preview_url: taskData.preview_url,
        schema_url: taskData.schema_url,
        original_url: taskData.original_url,
        processed_url: taskData.processed_url,
        edited_url: taskData.edited_url
      };
    } catch (e) {
      return null;
    }
  };

  const retryTask = async (id) => {
    try {
      await api.post(`/admin/images/${id}/retry`);
      await loadImages();
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('images.failed_to_retry_task'));
    }
  };

  const deleteTask = async (id) => {
    if (!window.confirm(t('images.confirm_delete_task'))) return;
    try {
      await api.delete(`/admin/images/${id}`);
      await loadImages();
    } catch (e) {
      alert(e?.response?.data?.error || e.message || t('images.failed_to_delete_task'));
    }
  };

  const filtered = useMemo(() => {
    const term = search.toLowerCase();
    return (images || []).filter((it) => {
      
      const matchesText = !term || 
        (it.user_email || '').toLowerCase().includes(term) || 
        (it.coupon_id || '').toLowerCase().includes(term) ||
        (it.partner_code || '').toLowerCase().includes(term);
      
      return matchesText;
    });
  }, [images, search]);

  const statusBadge = (st) => {
    const map = { uploaded: 'secondary', edited: 'secondary', processing: 'info', processed: 'primary', completed: 'success', failed: 'danger' };
    return <Badge color={map[st] || 'light'}>{st || '-'}</Badge>;
  };

  
  const formatDateTime = (dateString) => {
    if (!dateString) return '-';
    try {
      const date = new Date(dateString);
      
      const mskDate = new Date(date.getTime() + (3 * 60 * 60 * 1000));
      return mskDate.toLocaleString('ru-RU', {
        timeZone: 'Europe/Moscow',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      });
    } catch {
      return dateString;
    }
  };

  return (
    <Fragment>
      <div className="app-page-title">
        <div className="page-title-wrapper">
          <div className="page-title-heading">
            <div className="page-title-icon">
              <i className="pe-7s-photo icon-gradient bg-mean-fruit" />
            </div>
            <div>
              {t('images.image_processing_title')}
              <div className="page-title-subheading">{t('images.image_processing_subtitle')}</div>
            </div>
          </div>
        </div>
      </div>

      <Row>
        <Col lg="12">
          <Card className="main-card mb-3">
            <CardHeader>
              <CardTitle>{t('images.queue')}</CardTitle>
            </CardHeader>
            <CardBody>
              <Row className="mb-3">
                <Col md="3">
                  <Input
                    type="select"
                    value={status}
                    onChange={(e) => setStatus(e.target.value)}
                  >
                    {statusOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </Input>
                </Col>
                <Col md="6">
                  <InputGroup>
                    <InputGroupText>
                      <i className="pe-7s-search" />
                    </InputGroupText>
                    <Input
                      placeholder={t('images.search_by_partner_id')}
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                    />
                  </InputGroup>
                </Col>
                <Col md="3">
                  <Button
                    color="secondary"
                    size="sm"
                    onClick={() => {
                      setDateFrom('');
                      setDateTo('');
                      setSearch('');
                      setStatus('all');
                    }}
                  >
                    {t('images.reset_filters')}
                  </Button>
                </Col>
              </Row>
              
              <Row className="mb-3">
                <Col md="3">
                  <label className="form-label">{t('images.date_from')}</label>
                  <Input
                    type="text"
                    value={dateFrom}
                    onChange={(e) => setDateFrom(applyDateMask(e.target.value))}
                    placeholder={t('images.date_from_placeholder')}
                  />
                </Col>
                <Col md="3">
                  <label className="form-label">{t('images.date_to')}</label>
                  <Input
                    type="text"
                    value={dateTo}
                    onChange={(e) => setDateTo(applyDateMask(e.target.value))}
                    placeholder={t('images.date_to_placeholder')}
                  />
                </Col>
              </Row>

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
                        <th>{t('images.code')}</th>
                        <th>ID</th>
                        <th>Coupon</th>
                        <th>Email</th>
                        <th>{t('images.created_time')}</th>
                        <th>{t('images.status')}</th>
                        <th>{t('images.progress')}</th>
                        <th>{t('common.actions')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filtered.map((img) => (
                        <tr key={img.id}>
                          <td><Badge color="secondary">{img.partner_code || '-'}</Badge></td>
                          <td>{img.id}</td>
                          <td><code>{img.coupon_id}</code></td>
                          <td>{img.user_email || '-'}</td>
                          <td><small>{formatDateTime(img.created_at || img.createdAt)}</small></td>
                          <td>{statusBadge(img.status)}</td>
                          <td>
                            <div className="progress" style={{height: 8}}>
                              <div className="progress-bar" role="progressbar" style={{width: `${computeProgress(img.status)}%`}} />
                            </div>
                          </td>
                          <td>
                            <div className="btn-group">
                              <Button
                                size="sm"
                                color="info"
                                onClick={() => openDetails(img)}
                              >
                                {t('common.view')}
                              </Button>
                              <Button
                                size="sm"
                                color="secondary"
                                onClick={async () => {
                                  try {
                                    
                                    const response = await api.get(`/public/schemas/${img.id}/download`, {
                                      responseType: 'blob'
                                    });
                                    
                                    
                                    const url = window.URL.createObjectURL(new Blob([response.data]));
                                    const link = document.createElement('a');
                                    link.href = url;
                                    link.setAttribute('download', `schema_${img.id}.zip`);
                                    document.body.appendChild(link);
                                    link.click();
                                    link.remove();
                                    window.URL.revokeObjectURL(url);
                                  } catch (error) {
                                    console.error('Error downloading schema:', error);
                                    alert(t('images.failed_to_download_schema') || 'Failed to download schema');
                                  }
                                }}
                              >
                                {t('images.install_schema')}
                              </Button>
                              {img.status === 'failed' && (
                                <Button
                                  size="sm"
                                  color="warning"
                                  onClick={() => retryTask(img.id)}
                                >
                                  {t('images.retry_task')}
                                </Button>
                              )}
                              <Button
                                size="sm"
                                color="danger"
                                onClick={() => deleteTask(img.id)}
                              >
                                {t('images.delete_task')}
                              </Button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </Table>
                </div>
              )}
            </CardBody>
          </Card>
        </Col>
      </Row>

      <Modal isOpen={detailsOpen} toggle={() => setDetailsOpen(false)} size="lg">
        <ModalHeader toggle={() => setDetailsOpen(false)}>
          {t('images.image_processing_title')}: {selected?.id}
        </ModalHeader>
        <ModalBody>
          {details && (
            <div>
              <p><strong>ID:</strong> {details.id}</p>
              <p><strong>Coupon ID:</strong> {details.coupon_id}</p>
              <p><strong>Email:</strong> {details.user_email}</p>
              <p><strong>Status:</strong> {statusBadge(details.status)}</p>
              <p><strong>Created At:</strong> {formatDateTime(details.created_at)}</p>
              {details.updated_at && (
                <p><strong>Updated At:</strong> {formatDateTime(details.updated_at)}</p>
              )}
              {details.error_message && (
                <div className="alert alert-danger">
                  <strong>Error:</strong> {details.error_message}
                </div>
              )}
            </div>
          )}
          
          {detailsStatus && (
            <div className="mt-4">
              {detailsStatus.original_url && (
                <p><strong>Original:</strong> <a href={detailsStatus.original_url} target="_blank" rel="noopener noreferrer">View</a></p>
              )}
              {detailsStatus.edited_url && (
                <p><strong>Edited:</strong> <a href={detailsStatus.edited_url} target="_blank" rel="noopener noreferrer">View</a></p>
              )}
              {detailsStatus.processed_url && (
                <p><strong>Processed:</strong> <a href={detailsStatus.processed_url} target="_blank" rel="noopener noreferrer">View</a></p>
              )}
              {detailsStatus.preview_url && (
                <p><strong>Preview:</strong> <a href={detailsStatus.preview_url} target="_blank" rel="noopener noreferrer">View</a></p>
              )}
              {detailsStatus.schema_url && (
                <p><strong>Schema:</strong> <a href={detailsStatus.schema_url} target="_blank" rel="noopener noreferrer">Download</a></p>
              )}
            </div>
          )}
        </ModalBody>
        <ModalFooter>
          <Button
            color="primary"
            disabled={!selected?.id || selected?.status !== 'completed'}
            onClick={async () => {
              if (!selected?.id) return;
              try {
                
                const response = await api.get(`/public/schemas/${selected.id}/download`, {
                  responseType: 'blob'
                });
                
                
                const url = window.URL.createObjectURL(new Blob([response.data]));
                const link = document.createElement('a');
                link.href = url;
                link.setAttribute('download', `schema_${selected.id}.zip`);
                document.body.appendChild(link);
                link.click();
                link.remove();
                window.URL.revokeObjectURL(url);
              } catch (error) {
                console.error('Error downloading schema:', error);
                alert(t('images.failed_to_download_schema') || 'Failed to download schema');
              }
            }}
          >
            {t('coupons.download_materials')}
          </Button>
          <Button color="secondary" onClick={() => setDetailsOpen(false)}>
            {t('common.close')}
          </Button>
        </ModalFooter>
      </Modal>
    </Fragment>
  );
};

export default ImageQueue;