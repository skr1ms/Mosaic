import React, { Fragment, useState, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert
} from 'reactstrap';
import api from '../../../api/api';
import ArticleGrid from './ArticleGrid';

const CreatePartner = () => {
    const { t } = useTranslation();
    const fileInputRef = useRef(null);
    const [formData, setFormData] = useState({
        login: '',
        password: '',
        confirmPassword: '',
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
        color1: '#3B82F6',
        color2: '#10B981',
        color3: '#F59E0B',
        brandColorsRaw: '#3B82F6, #10B981, #F59E0B'
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [showPwd, setShowPwd] = useState(false);
    const [showConfirmPwd, setShowConfirmPwd] = useState(false);

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
        // Простая проверка HEX-цветов (необязательные)
        const hexRe = /^#(?:[0-9a-fA-F]{3}){1,2}$/;
        [formData.color1, formData.color2, formData.color3]
            .filter(Boolean)
            .forEach((c, idx) => {
                if (!hexRe.test(c.trim())) {
                    newErrors[`color${idx + 1}`] = t('partners.invalid_hex_color');
                }
            });

        if (!formData.login.trim()) newErrors.login = t('partners.login_required');
        if (!formData.password) newErrors.password = t('partners.password_required');
        if (formData.password !== formData.confirmPassword) {
            newErrors.confirmPassword = t('partners.passwords_dont_match');
        }
        if (!formData.domain.trim()) newErrors.domain = t('partners.domain_required');
        if (!formData.brandName.trim()) newErrors.brandName = t('partners.brand_name_required');
        if (!formData.email.trim()) newErrors.email = t('partners.email_required');

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (formData.email && !emailRegex.test(formData.email)) {
            newErrors.email = t('partners.invalid_email_format');
        }

        const domainRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z0-9]{1,}$/;
        if (formData.domain && !domainRegex.test(formData.domain)) {
            newErrors.domain = t('partners.invalid_domain_format');
        }


        if (formData.password) {
            const pwd = formData.password;
            const hasMin = pwd.length >= 8;
            const hasUpper = /[A-ZА-Я]/.test(pwd);
            const hasDigit = /\d/.test(pwd);
            const hasSpecial = /[^\w\s]/.test(pwd);
            if (!hasMin || !hasUpper || !hasDigit || !hasSpecial) {
                newErrors.password = t('partners.password_validation_error');
            }
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!validateForm()) {
            return;
        }

        setIsSubmitting(true);
        setSubmitMessage('');

        try {
            // Сначала создаем партнера JSON'ом
            const payload = {
                login: formData.login,
                password: formData.password,
                domain: formData.domain,
                brand_name: formData.brandName,
                ozon_link: formData.ozonLink,
                wildberries_link: formData.wildberriesLink,
                ozon_link_template: formData.ozonLinkTemplate,
                wildberries_link_template: formData.wildberriesLinkTemplate,
                email: formData.email,
                address: formData.address,
                phone: formData.phone,
                telegram: formData.telegram,
                whatsapp: formData.whatsapp,
                allow_sales: formData.allowSales,
                brand_colors: [formData.color1, formData.color2, formData.color3].filter(v => v && v.trim())
            };

            const res = await api.post('/admin/partners', payload);


            if (formData.logoFile && res?.data?.id) {
                const fd = new FormData();
                fd.append('logo', formData.logoFile);
                await api.post(`/admin/partners/${res.data.id}/logo`, fd, {
                    headers: { 'Content-Type': 'multipart/form-data' }
                });
            }


            const localArticles = localStorage.getItem('create_partner_articles')
            if (localArticles && res?.data?.id) {
                try {
                    const articles = JSON.parse(localArticles)
                    const articlePromises = []


                    Object.entries(articles).forEach(([marketplace, styles]) => {
                        Object.entries(styles).forEach(([style, sizes]) => {
                            Object.entries(sizes).forEach(([size, sku]) => {

                                articlePromises.push(
                                    api.put(`/admin/partners/${res.data.id}/articles/sku`, {
                                        marketplace,
                                        style,
                                        size,
                                        sku: sku ? sku.trim() : ''
                                    })
                                )
                            })
                        })
                    })

                    if (articlePromises.length > 0) {
                        await Promise.all(articlePromises)
                        // Очищаем localStorage после успешного сохранения
                        localStorage.removeItem('create_partner_articles')
                    }
                } catch (articleError) {
                    console.error('Error saving articles during partner creation:', articleError)

                }
            }

            setSubmitMessage(t('partners.partner_created_successfully'));

            setTimeout(() => {
                window.location.href = '/#/partners/list';
            }, 1500);

        } catch (error) {
            setSubmitMessage(t('partners.failed_to_create_partner') + ': ' + (error?.response?.data?.error || error.message));
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-user icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('partners.add_partner_title')}
                            <div className="page-title-subheading">
                                {t('partners.add_partner_subtitle')}
                            </div>
                        </div>
                    </div>

                </div>
            </div>



            <Form onSubmit={handleSubmit}>
                <Row className="justify-content-center">
                    <Col lg="10" xl="8">
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

                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="password">{t('partners.password')}</Label>
                                            <div className="input-group">
                                                <Input
                                                    type={showPwd ? 'text' : 'password'}
                                                    id="password"
                                                    name="password"
                                                    value={formData.password}
                                                    onChange={handleInputChange}
                                                    invalid={!!errors.password}
                                                    placeholder={t('partners.password_requirements')}
                                                />
                                                <button
                                                    type="button"
                                                    className="btn btn-outline-secondary"
                                                    onClick={() => setShowPwd(v => !v)}
                                                    aria-label={showPwd ? t('partners.hide_password') : t('partners.show_password')}
                                                >
                                                    <i className={showPwd ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                                                </button>
                                            </div>
                                            <FormFeedback>{errors.password}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="confirmPassword">{t('partners.confirm_password')}</Label>
                                            <div className="input-group">
                                                <Input
                                                    type={showConfirmPwd ? 'text' : 'password'}
                                                    id="confirmPassword"
                                                    name="confirmPassword"
                                                    value={formData.confirmPassword}
                                                    onChange={handleInputChange}
                                                    invalid={!!errors.confirmPassword}
                                                    placeholder={t('partners.password_requirements')}
                                                />
                                                <button
                                                    type="button"
                                                    className="btn btn-outline-secondary"
                                                    onClick={() => setShowConfirmPwd(v => !v)}
                                                    aria-label={showConfirmPwd ? t('partners.hide_password') : t('partners.show_password')}
                                                >
                                                    <i className={showConfirmPwd ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                                                </button>
                                            </div>
                                            <FormFeedback>{errors.confirmPassword}</FormFeedback>
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
                                    <small className="form-text text-info">
                                        💡 {t('partners.default_colors_note')}
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
                                        placeholder={t('partners.address_placeholder')}
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

                        <ArticleGrid partnerId={null} isReadOnly={false} />

                        { }
                        {submitMessage && (
                            <Alert color={submitMessage.includes(t('partners.partner_created_successfully')) ? 'success' : 'danger'} className="mt-4">
                                {submitMessage}
                            </Alert>
                        )}

                        { }
                        <div className="d-flex justify-content-between mt-4 mb-4">
                            <Button
                                color="secondary"
                                size="lg"
                                onClick={() => window.location.href = '/#/partners/list'}
                                className="px-5 py-2"
                            >
                                <i className="pe-7s-back me-2"></i> {t('partners.back_to_list')}
                            </Button>

                            <Button
                                type="submit"
                                color="primary"
                                size="lg"
                                disabled={isSubmitting}
                                className="px-5 py-2"
                            >
                                {isSubmitting ? (
                                    <>
                                        <i className="fa fa-spinner fa-spin me-2"></i> {t('common.saving')}...
                                    </>
                                ) : (
                                    <>
                                        <i className="pe-7s-check me-2"></i> {t('common.create')}
                                    </>
                                )}
                            </Button>
                        </div>
                    </Col>
                </Row>
            </Form>
        </Fragment>
    );
};

export default CreatePartner; 