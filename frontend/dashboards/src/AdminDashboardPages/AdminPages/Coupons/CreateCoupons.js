import React, { Fragment, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert, Badge, Progress
} from 'reactstrap';
import api from '../../../api/api';

const CreateCoupons = () => {
    const { t } = useTranslation();
    const [formData, setFormData] = useState({
        quantity: 1,
        partnerId: '',
        size: '',
        style: ''
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [submitOk, setSubmitOk] = useState(false);
    const [generationProgress, setGenerationProgress] = useState(0);
    const [generatedCoupons, setGeneratedCoupons] = useState([]);
    const [partners, setPartners] = useState([]);
    const [ownPartnerId, setOwnPartnerId] = useState(null); // реальный UUID партнёра с кодом 0000

    useEffect(() => {
        const loadPartners = async () => {
            try {
                const res = await api.get('/admin/partners');
                const list = res.data?.partners || [];
                
                const own = list.find(p => p?.partner_code === '0000');
                setOwnPartnerId(own?.id || null);
                const filtered = list.filter(p => p?.partner_code !== '0000');
                
                setPartners([
                    { id: '0000', brand_name: t('coupons.own_coupons') },
                    ...filtered,
                ]);
            } catch (_) {}
        };
        loadPartners();
    }, [t]);

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

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
        
        if (errors[name]) {
            setErrors(prev => ({
                ...prev,
                [name]: ''
            }));
        }
    };

    const validateForm = () => {
        const newErrors = {};

        if (!formData.quantity || formData.quantity < 1) {
            newErrors.quantity = t('coupons.quantity_validation');
        }
        if (formData.quantity > 1000) {
            newErrors.quantity = t('coupons.max_quantity_validation');
        }
        if (!formData.partnerId) {
            newErrors.partnerId = t('coupons.partner_validation');
        }
        if (!formData.size) {
            newErrors.size = t('coupons.size_validation');
        }
        if (!formData.style) {
            newErrors.style = t('coupons.style_validation');
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
        setSubmitOk(false);
        setGenerationProgress(0);
        setGeneratedCoupons([]);

        try {
            // Симуляция прогресса генерации
            const progressInterval = setInterval(() => {
                setGenerationProgress(prev => {
                    if (prev >= 90) {
                        clearInterval(progressInterval);
                        return 90;
                    }
                    return prev + 10;
                });
            }, 200);

            const response = await api.post('/admin/coupons', {
                count: Number(formData.quantity),
                partner_id: (formData.partnerId === '0000' ? (ownPartnerId || '00000000-0000-0000-0000-000000000000') : formData.partnerId) || '00000000-0000-0000-0000-000000000000',
                size: formData.size,
                style: formData.style
            });

            clearInterval(progressInterval);
            setGenerationProgress(100);

            if (response.data?.codes) {
                
                const coupons = response.data.codes.map(code => ({ code }));
                setGeneratedCoupons(coupons);
                setSubmitOk(true);
                setSubmitMessage(t('coupons.generating_new_coupons'));
                
                
                setFormData({
                    quantity: 1,
                    partnerId: '',
                    size: '',
                    style: ''
                });
            }
        } catch (error) {
            setSubmitOk(false);
            setSubmitMessage(error?.response?.data?.error || t('coupons.error_creating_coupons'));
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
                            <i className="pe-7s-ticket icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>{t('coupons.create')}
                            <div className="page-title-subheading">
                                {t('coupons.generating_new_coupons')}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="8">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('coupons.create')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Form onSubmit={handleSubmit}>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="quantity">{t('coupons.quantity')}</Label>
                                            <Input
                                                id="quantity"
                                                name="quantity"
                                                type="number"
                                                min="1"
                                                max="1000"
                                                value={formData.quantity}
                                                onChange={handleInputChange}
                                                invalid={!!errors.quantity}
                                            />
                                            <FormFeedback>{errors.quantity}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="partnerId">{t('coupons.partner')}</Label>
                                            <Input
                                                id="partnerId"
                                                name="partnerId"
                                                type="select"
                                                value={formData.partnerId}
                                                onChange={handleInputChange}
                                                invalid={!!errors.partnerId}
                                            >
                                                <option value="">{t('common.select')}</option>
                                                {partners.map(partner => (
                                                    <option key={partner.id} value={partner.id}>
                                                        {partner.brand_name}
                                                    </option>
                                                ))}
                                            </Input>
                                            <FormFeedback>{errors.partnerId}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="size">{t('coupons.size')}</Label>
                                            <Input
                                                id="size"
                                                name="size"
                                                type="select"
                                                value={formData.size}
                                                onChange={handleInputChange}
                                                invalid={!!errors.size}
                                            >
                                                <option value="">{t('common.select')}</option>
                                                {sizeOptions.map(option => (
                                                    <option key={option.value} value={option.value}>
                                                        {option.value}
                                                    </option>
                                                ))}
                                            </Input>
                                            <FormFeedback>{errors.size}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="style">{t('coupons.style')}</Label>
                                            <Input
                                                id="style"
                                                name="style"
                                                type="select"
                                                value={formData.style}
                                                onChange={handleInputChange}
                                                invalid={!!errors.style}
                                            >
                                                <option value="">{t('common.select')}</option>
                                                {styleOptions.map(option => (
                                                    <option key={option.value} value={option.value}>
                                                        {option.label}
                                                    </option>
                                                ))}
                                            </Input>
                                            <FormFeedback>{errors.style}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <div className="text-center">
                                    <Button
                                        type="submit"
                                        color="primary"
                                        size="lg"
                                        disabled={isSubmitting}
                                    >
                                        {isSubmitting ? t('common.loading') : t('coupons.create')}
                                    </Button>
                                </div>
                            </Form>

                            {isSubmitting && (
                                <div className="mt-4">
                                    <h6>{t('coupons.generation_progress')}</h6>
                                    <Progress value={generationProgress} color="success" />
                                </div>
                            )}

                            {submitMessage && (
                                <Alert 
                                    color={submitOk ? 'success' : 'danger'} 
                                    className="mt-3"
                                >
                                    {submitMessage}
                                </Alert>
                            )}
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="4">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>{t('coupons.generated_coupons')}</CardTitle>
                        </CardHeader>
                        <CardBody>
                            {generatedCoupons.length > 0 ? (
                                <div>
                                    {generatedCoupons.map((coupon, index) => (
                                        <div key={index} className="mb-2">
                                            <Badge color="success" className="mr-2">
                                                {coupon.code}
                                            </Badge>
                                            <small className="text-muted">
                                                {t('coupons.first_4_digits')}
                                            </small>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-muted text-center">
                                    {t('tables.no_data')}
                                </p>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default CreateCoupons; 