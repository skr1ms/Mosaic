import React, { Fragment, useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert
} from 'reactstrap';
import api from '../../../api/api';

const EditPartner = () => {
    const { t } = useTranslation();
    const { id } = useParams();
    const navigate = useNavigate();
    const fileInputRef = useRef(null);
    
    const [formData, setFormData] = useState({
        login: '',
        domain: '',
        brandName: '',
        logoFile: null,
        ozonLink: '',
        wildberriesLink: '',
        email: '',
        address: '',
        phone: '',
        telegram: '',
        whatsapp: '',
        allowSales: true,
        color1: '',
        color2: '',
        color3: '',
        brandColorsRaw: ''
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const load = async () => {
            try {
                const res = await api.get(`/admin/partners/${id}`);
                const p = res.data || {};
                const brandColors = p.brand_colors || [];
                // Если у партнера нет цветов, показываем дефолтные
                const defaultColors = ['#3B82F6', '#10B981', '#F59E0B'];
                const effectiveColors = brandColors.length > 0 ? brandColors : defaultColors;
                
                setFormData({
                    login: p.login || '',
                    domain: p.domain || '',
                    brandName: p.brand_name || '',
                    logoFile: null,
                    ozonLink: p.ozon_link || '',
                    wildberriesLink: p.wildberries_link || '',
                    email: p.email || '',
                    address: p.address || '',
                    phone: p.phone || '',
                    telegram: p.telegram || '',
                    whatsapp: p.whatsapp || '',
                    allowSales: Boolean(p.allow_sales),
                    color1: effectiveColors[0] || '',
                    color2: effectiveColors[1] || '',
                    color3: effectiveColors[2] || '',
                    brandColorsRaw: brandColors.length > 0 ? brandColors.join(', ') : defaultColors.join(', ')
                });
            } catch (_) {}
            setLoading(false);
        };
        load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [id]);

    const handleInputChange = (e) => {
        const { name, value, type, checked } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value
        }));
        
        if (errors[name]) {
            setErrors(prev => ({
                ...prev,
                [name]: ''
            }));
        }
    };

    const handleFileChange = (e) => {
        const file = e.target.files[0];
        setFormData(prev => ({
            ...prev,
            logoFile: file
        }));
    };

    const handleColorsChange = (e) => {
        const raw = e.target.value || '';
        const tokens = raw
            .split(/[\s,;]+/)
            .map(s => s.trim())
            .filter(Boolean)
            .slice(0, 3);
        setFormData(prev => ({
            ...prev,
            brandColorsRaw: raw,
            color1: tokens[0] || '',
            color2: tokens[1] || '',
            color3: tokens[2] || ''
        }));
        if (errors.color1 || errors.color2 || errors.color3) {
            setErrors(prev => ({ ...prev, color1: '', color2: '', color3: '' }));
        }
    };

    const validateForm = () => {
        const newErrors = {};
        const hexRe = /^#(?:[0-9a-fA-F]{3}){1,2}$/;
        [formData.color1, formData.color2, formData.color3]
            .filter(Boolean)
            .forEach((c, idx) => {
                if (!hexRe.test(c.trim())) {
                    newErrors[`color${idx + 1}`] = t('partners.invalid_hex_color');
                }
            });

        if (!formData.login.trim()) newErrors.login = t('partners.login_required');
        if (!formData.domain.trim()) newErrors.domain = t('partners.domain_required');
        if (!formData.brandName.trim()) newErrors.brandName = t('partners.brand_name_required');
        if (!formData.email.trim()) newErrors.email = t('partners.email_required');

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (formData.email && !emailRegex.test(formData.email)) {
            newErrors.email = t('partners.invalid_email_format');
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        if (!validateForm()) return;
        
        setIsSubmitting(true);
        setSubmitMessage('');
        
        try {
            // 1) Обновляем партнера JSON'ом (PUT)
            // Определяем, изменил ли пользователь цвета
            const originalBrandColors = formData.brandColorsRaw.split(/[\s,;]+/).map(s => s.trim()).filter(Boolean);
            const hasOriginalColors = originalBrandColors.length > 0;
            
            // Если у партнера были оригинальные цвета, отправляем их
            // Если не было, но пользователь ввел новые - отправляем новые
            // Если не было и пользователь не ввел - отправляем пустой массив
            let brandColorsToSend;
            if (hasOriginalColors) {
                brandColorsToSend = [formData.color1, formData.color2, formData.color3].filter(v => v && v.trim());
            } else {
                // Проверяем, ввел ли пользователь новые цвета
                const newColors = [formData.color1, formData.color2, formData.color3].filter(v => v && v.trim());
                brandColorsToSend = newColors.length > 0 ? newColors : [];
            }
            
            const payload = {
                login: formData.login,
                domain: formData.domain,
                brand_name: formData.brandName,
                ozon_link: formData.ozonLink,
                wildberries_link: formData.wildberriesLink,
                email: formData.email,
                address: formData.address,
                phone: formData.phone,
                telegram: formData.telegram,
                whatsapp: formData.whatsapp,
                allow_sales: formData.allowSales,
                brand_colors: brandColorsToSend
            };

            await api.put(`/admin/partners/${id}`, payload);

            // 2) Логотип отправляем отдельным запросом, если выбран
            if (formData.logoFile) {
                const fd = new FormData();
                fd.append('logo', formData.logoFile);
                await api.post(`/admin/partners/${id}/logo`, fd, {
                    headers: { 'Content-Type': 'multipart/form-data' }
                });
            }
            setSubmitMessage(t('partners.changes_saved_successfully'));
            
            setTimeout(() => {
                navigate('/partners/list');
            }, 1500);
            
        } catch (error) {
            setSubmitMessage(t('partners.failed_to_save_changes'));
        } finally {
            setIsSubmitting(false);
        }
    };

    if (loading) {
        return (
            <Fragment>
                <div className="app-page-title">
                    <div className="page-title-wrapper">
                        <div className="page-title-heading">
                            <div className="page-title-icon">
                                <i className="pe-7s-user icon-gradient bg-mean-fruit"></i>
                            </div>
                            <div>{t('partners.edit_partner_title')}
                                <div className="page-title-subheading">
                                    {t('partners.loading_partner_data')}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="text-center py-5">
                    <i className="fa fa-spinner fa-spin fa-3x"></i>
                </div>
            </Fragment>
        );
    }

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-user icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('partners.edit_partner_title')}
                            <div className="page-title-subheading">
                                {t('partners.edit_partner_subtitle').replace('{id}', id)}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => navigate('/partners/list')}
                        >
                            <i className="pe-7s-back"></i> {t('partners.back_to_list')}
                        </Button>
                    </div>
                </div>
            </div>

            {submitMessage && (
                <Alert color={submitMessage.includes(t('partners.changes_saved_successfully')) ? 'success' : 'danger'}>
                    {submitMessage}
                </Alert>
            )}

            <Form onSubmit={handleSubmit}>
                <Row>
                    <Col lg="8">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('partners.main_information')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="login">{t('partners.login_for_entry')}</Label>
                                            <Input
                                                type="text"
                                                id="login"
                                                name="login"
                                                value={formData.login}
                                                onChange={handleInputChange}
                                                invalid={!!errors.login}
                                                placeholder={t('partners.enter_login')}
                                            />
                                            <FormFeedback>{errors.login}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="brandName">{t('partners.brand_name')}</Label>
                                            <Input
                                                type="text"
                                                id="brandName"
                                                name="brandName"
                                                value={formData.brandName}
                                                onChange={handleInputChange}
                                                invalid={!!errors.brandName}
                                                placeholder={t('partners.enter_brand_name')}
                                            />
                                            <FormFeedback>{errors.brandName}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <FormGroup>
                                    <Label for="domain">{t('partners.domain_name')}</Label>
                                    <Input
                                        type="text"
                                        id="domain"
                                        name="domain"
                                        value={formData.domain}
                                        onChange={handleInputChange}
                                        invalid={!!errors.domain}
                                        placeholder={t('partners.domain_example')}
                                    />
                                    <FormFeedback>{errors.domain}</FormFeedback>
                                    <small className="form-text text-muted">
                                        {t('partners.domain_white_label')}
                                    </small>
                                </FormGroup>

                                <FormGroup>
                                    <Label for="logoFile">{t('partners.partner_logo')}</Label>
                                    <input
                                        ref={fileInputRef}
                                        type="file"
                                        id="logoFile"
                                        name="logoFile"
                                        accept="image/*"
                                        onChange={handleFileChange}
                                        style={{ display: 'none' }}
                                    />
                                    <div className="d-flex align-items-center">
                                        <Button type="button" color="secondary" onClick={() => fileInputRef.current && fileInputRef.current.click()}>
                                            {t('common.select_file')}
                                        </Button>
                                        <small className="text-muted ml-2">
                                            {formData.logoFile ? formData.logoFile.name : t('common.file_not_selected')}
                                        </small>
                                    </div>
                                    <small className="form-text text-muted">
                                        {t('partners.logo_recommended_size')}
                                    </small>
                                </FormGroup>

                                <FormGroup>
                                    <Label for="brandColorsRaw">{t('partners.brand_colors_label')}</Label>
                                    <Input
                                        type="text"
                                        id="brandColorsRaw"
                                        name="brandColorsRaw"
                                        value={formData.brandColorsRaw}
                                        onChange={handleColorsChange}
                                        invalid={!!(errors.color1 || errors.color2 || errors.color3)}
                                        placeholder={t('partners.brand_colors_placeholder')}
                                    />
                                    {(errors.color1 || errors.color2 || errors.color3) && (
                                        <FormFeedback>{t('partners.invalid_hex_color')}</FormFeedback>
                                    )}
                                    <small className="form-text text-muted">
                                        {t('partners.brand_colors_help')}
                                    </small>
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('partners.contact_information')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="email">{t('common.email')} *</Label>
                                            <Input
                                                type="email"
                                                id="email"
                                                name="email"
                                                value={formData.email}
                                                onChange={handleInputChange}
                                                invalid={!!errors.email}
                                                placeholder="admin@example.com"
                                            />
                                            <FormFeedback>{errors.email}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="phone">{t('partners.phone')}</Label>
                                            <Input
                                                type="tel"
                                                id="phone"
                                                name="phone"
                                                value={formData.phone}
                                                onChange={handleInputChange}
                                                placeholder={t('partners.phone_placeholder')}
                                                title={t('partners.phone_help')}
                                            />
                                            <small className="form-text text-muted">{t('partners.phone_help')}</small>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="telegram">{t('partners.telegram')}</Label>
                                            <Input
                                                type="text"
                                                id="telegram"
                                                name="telegram"
                                                value={formData.telegram}
                                                onChange={handleInputChange}
                                                placeholder="@username"
                                            />
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="whatsapp">{t('partners.whatsapp')}</Label>
                                            <Input
                                                type="text"
                                                id="whatsapp"
                                                name="whatsapp"
                                                value={formData.whatsapp}
                                                onChange={handleInputChange}
                                                placeholder={t('partners.phone_placeholder')}
                                                title={t('partners.phone_help')}
                                            />
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <FormGroup>
                                    <Label for="address">{t('partners.address')}</Label>
                                    <Input
                                        type="textarea"
                                        id="address"
                                        name="address"
                                        value={formData.address}
                                        onChange={handleInputChange}
                                        rows="3"
                                        placeholder="Полный адрес партнера"
                                    />
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('partners.marketplace_links')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="ozonLink">{t('partners.ozon_link')}</Label>
                                            <Input
                                                type="url"
                                                id="ozonLink"
                                                name="ozonLink"
                                                value={formData.ozonLink}
                                                onChange={handleInputChange}
                                                placeholder="https://www.ozon.ru/seller/..."
                                            />
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="wildberriesLink">{t('partners.wildberries_link')}</Label>
                                            <Input
                                                type="url"
                                                id="wildberriesLink"
                                                name="wildberriesLink"
                                                value={formData.wildberriesLink}
                                                onChange={handleInputChange}
                                                placeholder="https://www.wildberries.ru/seller/..."
                                            />
                                        </FormGroup>
                                    </Col>
                                </Row>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('partners.brand_colors')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <FormGroup>
                                    <Label for="brandColorsRaw">{t('partners.brand_colors_label')}</Label>
                                    <Input
                                        type="text"
                                        id="brandColorsRaw"
                                        name="brandColorsRaw"
                                        value={formData.brandColorsRaw}
                                        onChange={handleColorsChange}
                                        invalid={!!(errors.color1 || errors.color2 || errors.color3)}
                                        placeholder={t('partners.brand_colors_placeholder')}
                                    />
                                    {(errors.color1 || errors.color2 || errors.color3) && (
                                        <FormFeedback>{t('partners.invalid_hex_color')}</FormFeedback>
                                    )}
                                    <small className="form-text text-muted">
                                        {t('partners.brand_colors_help')}
                                    </small>
                                    {formData.brandColorsRaw.includes('#3B82F6, #10B981, #F59E0B') && (
                                        <small className="form-text text-info">
                                            💡 {t('partners.default_colors_note')}
                                        </small>
                                    )}
                                    <small className="form-text text-info">
                                        💡 {t('partners.colors_editing_note')}
                                    </small>
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('partners.sales_settings')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <FormGroup check>
                                    <Label check>
                                        <Input
                                            type="checkbox"
                                            name="allowSales"
                                            checked={formData.allowSales}
                                            onChange={handleInputChange}
                                        />
                                        {t('partners.allow_sales')}
                                    </Label>
                                </FormGroup>
                            </CardBody>
                        </Card>
                    </Col>

                    <Col lg="4">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>{t('common.actions')}</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Button
                                    type="submit"
                                    color="primary"
                                    size="lg"
                                    block
                                    disabled={isSubmitting}
                                >
                                    {isSubmitting ? (
                                        <>
                                            <i className="fa fa-spinner fa-spin"></i> {t('common.saving')}...
                                        </>
                                    ) : (
                                        <>
                                            <i className="pe-7s-diskette"></i> {t('partners.save_changes')}
                                        </>
                                    )}
                                </Button>
                            </CardBody>
                        </Card>
                    </Col>
                </Row>
            </Form>
        </Fragment>
    );
};

export default EditPartner; 