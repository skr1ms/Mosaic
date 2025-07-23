import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert, Badge, Progress
} from 'reactstrap';

const CreateCoupons = () => {
    const [formData, setFormData] = useState({
        quantity: 1,
        partnerId: '',
        size: '',
        style: ''
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [generationProgress, setGenerationProgress] = useState(0);
    const [generatedCoupons, setGeneratedCoupons] = useState([]);

    const mockPartners = [
        { id: '0000', name: 'Собственные купоны' },
        { id: '1001', name: 'Мозаика Арт' },
        { id: '1002', name: 'Алмазная студия' },
        { id: '1003', name: 'Творческая мастерская' }
    ];

    const sizeOptions = [
        { value: '21x30', label: '21×30 см' },
        { value: '30x40', label: '30×40 см' },
        { value: '40x40', label: '40×40 см' },
        { value: '40x50', label: '40×50 см' },
        { value: '40x60', label: '40×60 см' },
        { value: '50x70', label: '50×70 см' }
    ];

    const styleOptions = [
        { value: 'grayscale', label: 'Оттенки серого' },
        { value: 'skin_tones', label: 'Оттенки телесного' },
        { value: 'pop_art', label: 'Поп-арт' },
        { value: 'max_colors', label: 'Максимум цветов' }
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
            newErrors.quantity = 'Количество должно быть больше 0';
        }
        if (formData.quantity > 10000) {
            newErrors.quantity = 'Максимальное количество: 10000';
        }
        if (!formData.partnerId) {
            newErrors.partnerId = 'Выберите партнера';
        }
        if (!formData.size) {
            newErrors.size = 'Выберите размер';
        }
        if (!formData.style) {
            newErrors.style = 'Выберите стиль обработки';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const generateCouponCode = (partnerId, index) => {
        const partnerCode = partnerId.padStart(4, '0');
        const randomPart = Math.floor(Math.random() * 100000000).toString().padStart(8, '0');
        const fullCode = partnerCode + randomPart;
        
        return `${fullCode.slice(0, 4)}-${fullCode.slice(4, 8)}-${fullCode.slice(8, 12)}`;
    };

    const simulateGeneration = async (quantity) => {
        const coupons = [];
        const batchSize = Math.min(10, quantity);
        const totalBatches = Math.ceil(quantity / batchSize);
        
        for (let batch = 0; batch < totalBatches; batch++) {
            const batchStart = batch * batchSize;
            const batchEnd = Math.min(batchStart + batchSize, quantity);
            
            for (let i = batchStart; i < batchEnd; i++) {
                const couponCode = generateCouponCode(formData.partnerId, i);
                coupons.push({
                    id: i + 1,
                    code: couponCode,
                    partnerId: formData.partnerId,
                    size: formData.size,
                    style: formData.style,
                    status: 'new',
                    createdAt: new Date().toISOString()
                });
            }
            
            const progress = Math.round(((batch + 1) / totalBatches) * 100);
            setGenerationProgress(progress);
            
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        
        return coupons;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        if (!validateForm()) {
            return;
        }

        setIsSubmitting(true);
        setSubmitMessage('');
        setGenerationProgress(0);
        setGeneratedCoupons([]);

        try {
            const coupons = await simulateGeneration(parseInt(formData.quantity));
            setGeneratedCoupons(coupons);
            
            setSubmitMessage(`Успешно создано ${coupons.length} купонов!`);
            
        } catch (error) {
            setSubmitMessage('Ошибка при создании купонов: ' + error.message);
        } finally {
            setIsSubmitting(false);
            setGenerationProgress(100);
        }
    };

    const downloadCoupons = () => {
        const couponCodes = generatedCoupons.map(coupon => coupon.code).join('\n');
        const blob = new Blob([couponCodes], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `coupons_${formData.partnerId}_${Date.now()}.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const resetForm = () => {
        setFormData({
            quantity: 1,
            partnerId: '',
            size: '',
            style: ''
        });
        setGeneratedCoupons([]);
        setSubmitMessage('');
        setGenerationProgress(0);
        setErrors({});
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
                        <div>Создание купонов
                            <div className="page-title-subheading">
                                Генерация новых купонов для партнеров
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {submitMessage && (
                <Alert color={submitMessage.includes('успешно') ? 'success' : 'danger'}>
                    {submitMessage}
                </Alert>
            )}

            <Row>
                <Col lg="8">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Параметры создания купонов</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Form onSubmit={handleSubmit}>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="quantity">Количество купонов *</Label>
                                            <Input
                                                type="number"
                                                id="quantity"
                                                name="quantity"
                                                value={formData.quantity}
                                                onChange={handleInputChange}
                                                invalid={!!errors.quantity}
                                                min="1"
                                                max="10000"
                                                placeholder="Введите количество"
                                                disabled={isSubmitting}
                                            />
                                            <FormFeedback>{errors.quantity}</FormFeedback>
                                            <small className="form-text text-muted">
                                                От 1 до 10000 купонов за раз
                                            </small>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="partnerId">Партнер *</Label>
                                            <Input
                                                type="select"
                                                id="partnerId"
                                                name="partnerId"
                                                value={formData.partnerId}
                                                onChange={handleInputChange}
                                                invalid={!!errors.partnerId}
                                                disabled={isSubmitting}
                                            >
                                                <option value="">Выберите партнера</option>
                                                {mockPartners.map(partner => (
                                                    <option key={partner.id} value={partner.id}>
                                                        {partner.name} ({partner.id})
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
                                            <Label for="size">Размер мозаики *</Label>
                                            <Input
                                                type="select"
                                                id="size"
                                                name="size"
                                                value={formData.size}
                                                onChange={handleInputChange}
                                                invalid={!!errors.size}
                                                disabled={isSubmitting}
                                            >
                                                <option value="">Выберите размер</option>
                                                {sizeOptions.map(size => (
                                                    <option key={size.value} value={size.value}>
                                                        {size.label}
                                                    </option>
                                                ))}
                                            </Input>
                                            <FormFeedback>{errors.size}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="style">Стиль обработки *</Label>
                                            <Input
                                                type="select"
                                                id="style"
                                                name="style"
                                                value={formData.style}
                                                onChange={handleInputChange}
                                                invalid={!!errors.style}
                                                disabled={isSubmitting}
                                            >
                                                <option value="">Выберите стиль</option>
                                                {styleOptions.map(style => (
                                                    <option key={style.value} value={style.value}>
                                                        {style.label}
                                                    </option>
                                                ))}
                                            </Input>
                                            <FormFeedback>{errors.style}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                {isSubmitting && (
                                    <FormGroup>
                                        <Label>Прогресс генерации</Label>
                                        <Progress
                                            value={generationProgress}
                                            color="success"
                                            className="mb-2"
                                        >
                                            {generationProgress}%
                                        </Progress>
                                        <small className="text-muted">
                                            Генерация купонов в процессе...
                                        </small>
                                    </FormGroup>
                                )}

                                <div className="d-flex gap-2">
                                    <Button 
                                        type="submit" 
                                        color="success" 
                                        size="lg"
                                        disabled={isSubmitting}
                                    >
                                        {isSubmitting ? (
                                            <>
                                                <i className="fa fa-spinner fa-spin"></i> Создание...
                                            </>
                                        ) : (
                                            <>
                                                <i className="pe-7s-check"></i> Создать купоны
                                            </>
                                        )}
                                    </Button>
                                    
                                    <Button 
                                        type="button" 
                                        color="secondary" 
                                        size="lg"
                                        onClick={resetForm}
                                        disabled={isSubmitting}
                                    >
                                        <i className="pe-7s-refresh"></i> Сброс
                                    </Button>
                                </div>
                            </Form>
                        </CardBody>
                    </Card>
                </Col>

                <Col lg="4">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Информация о генерации</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <div className="widget-chart-box">
                                <div className="widget-content">
                                    <div className="widget-subheading mb-2">
                                        Формат номера купона
                                    </div>
                                    <div className="text-center mb-3">
                                        <Badge color="info" className="badge-pill">
                                            XXXX-XXXX-XXXX
                                        </Badge>
                                    </div>
                                    <small className="text-muted">
                                        <strong>Первые 4 цифры:</strong> код партнера<br/>
                                        <strong>Последние 8 цифр:</strong> уникальный номер
                                    </small>
                                </div>
                            </div>

                            {formData.partnerId && (
                                <div className="mt-3">
                                    <div className="widget-subheading mb-2">
                                        Предварительный пример
                                    </div>
                                    <div className="text-center">
                                        <Badge color="secondary" className="badge-pill">
                                            {formData.partnerId.padStart(4, '0')}-XXXX-XXXX
                                        </Badge>
                                    </div>
                                </div>
                            )}
                        </CardBody>
                    </Card>

                    {generatedCoupons.length > 0 && (
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Результат генерации</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <div className="widget-chart-box">
                                    <div className="widget-content">
                                        <div className="widget-numbers text-success mb-2">
                                            <b>{generatedCoupons.length}</b>
                                        </div>
                                        <div className="widget-subheading mb-3">
                                            купонов создано
                                        </div>
                                        
                                        <div className="mb-3">
                                            <small className="text-muted">
                                                <strong>Диапазон номеров:</strong><br/>
                                                {generatedCoupons[0]?.code} - {generatedCoupons[generatedCoupons.length - 1]?.code}
                                            </small>
                                        </div>

                                        <Button 
                                            color="primary" 
                                            size="sm" 
                                            block
                                            onClick={downloadCoupons}
                                        >
                                            <i className="pe-7s-download"></i> Скачать список
                                        </Button>
                                    </div>
                                </div>
                            </CardBody>
                        </Card>
                    )}
                </Col>
            </Row>

            {generatedCoupons.length > 0 && (
                <Row>
                    <Col lg="12">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Созданные купоны (показано первых 10)</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <div className="table-responsive">
                                    <table className="table table-hover">
                                        <thead>
                                            <tr>
                                                <th>#</th>
                                                <th>Номер купона</th>
                                                <th>Размер</th>
                                                <th>Стиль</th>
                                                <th>Статус</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {generatedCoupons.slice(0, 10).map((coupon, index) => (
                                                <tr key={coupon.id}>
                                                    <td>{index + 1}</td>
                                                    <td>
                                                        <Badge color="secondary">{coupon.code}</Badge>
                                                    </td>
                                                    <td>{sizeOptions.find(s => s.value === coupon.size)?.label}</td>
                                                    <td>{styleOptions.find(s => s.value === coupon.style)?.label}</td>
                                                    <td>
                                                        <Badge color="success">Новый</Badge>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                                
                                {generatedCoupons.length > 10 && (
                                    <div className="text-center text-muted">
                                        ... и ещё {generatedCoupons.length - 10} купонов
                                    </div>
                                )}
                            </CardBody>
                        </Card>
                    </Col>
                </Row>
            )}
        </Fragment>
    );
};

export default CreateCoupons; 