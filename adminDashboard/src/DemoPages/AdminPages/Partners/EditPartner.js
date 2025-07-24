import React, { Fragment, useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert
} from 'reactstrap';

const EditPartner = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    
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
        allowSales: true
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');
    const [loading, setLoading] = useState(true);

    // Моковые данные партнеров для демонстрации
    const mockPartners = {
        '1001': {
            login: 'mosaic_art',
            domain: 'mosaic-art.ru',
            brandName: 'Мозаика Арт',
            ozonLink: 'https://ozon.ru/seller/mosaic-art',
            wildberriesLink: 'https://wildberries.ru/seller/mosaic-art',
            email: 'admin@mosaic-art.ru',
            address: 'г. Москва, ул. Примерная, д. 123',
            phone: '+7 (495) 123-45-67',
            telegram: 'https://t.me/mosaic_art',
            whatsapp: 'https://wa.me/79951234567',
            allowSales: true
        }
    };

    useEffect(() => {
        // Имитация загрузки данных партнера
        setTimeout(() => {
            const partnerData = mockPartners[id];
            if (partnerData) {
                setFormData(partnerData);
            }
            setLoading(false);
        }, 1000);
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

    const validateForm = () => {
        const newErrors = {};

        if (!formData.login.trim()) newErrors.login = 'Логин обязателен';
        if (!formData.domain.trim()) newErrors.domain = 'Доменное имя обязательно';
        if (!formData.brandName.trim()) newErrors.brandName = 'Название бренда обязательно';
        if (!formData.email.trim()) newErrors.email = 'Email обязателен';

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (formData.email && !emailRegex.test(formData.email)) {
            newErrors.email = 'Некорректный формат email';
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
            console.log('Обновление партнера:', formData);
            await new Promise(resolve => setTimeout(resolve, 2000));
            
            setSubmitMessage('Партнер успешно обновлен!');
            
            setTimeout(() => {
                navigate('/partners/list');
            }, 2000);
            
        } catch (error) {
            setSubmitMessage('Ошибка при обновлении партнера: ' + error.message);
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
                            <div>Редактирование партнера
                                <div className="page-title-subheading">
                                    Загрузка данных партнера...
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
                        <div>Редактировать партнера
                            <div className="page-title-subheading">
                                Изменение данных партнера #{id}
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => navigate('/partners/list')}
                        >
                            <i className="pe-7s-back"></i> Назад к списку
                        </Button>
                    </div>
                </div>
            </div>

            {submitMessage && (
                <Alert color={submitMessage.includes('успешно') ? 'success' : 'danger'}>
                    {submitMessage}
                </Alert>
            )}

            <Form onSubmit={handleSubmit}>
                <Row>
                    <Col lg="8">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Основная информация</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="login">Логин для входа *</Label>
                                            <Input
                                                type="text"
                                                id="login"
                                                name="login"
                                                value={formData.login}
                                                onChange={handleInputChange}
                                                invalid={!!errors.login}
                                                placeholder="Введите логин"
                                            />
                                            <FormFeedback>{errors.login}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="brandName">Название бренда *</Label>
                                            <Input
                                                type="text"
                                                id="brandName"
                                                name="brandName"
                                                value={formData.brandName}
                                                onChange={handleInputChange}
                                                invalid={!!errors.brandName}
                                                placeholder="Название бренда партнера"
                                            />
                                            <FormFeedback>{errors.brandName}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <FormGroup>
                                    <Label for="domain">Доменное имя *</Label>
                                    <Input
                                        type="text"
                                        id="domain"
                                        name="domain"
                                        value={formData.domain}
                                        onChange={handleInputChange}
                                        invalid={!!errors.domain}
                                        placeholder="example.com"
                                    />
                                    <FormFeedback>{errors.domain}</FormFeedback>
                                </FormGroup>

                                <FormGroup>
                                    <Label for="logoFile">Логотип партнера</Label>
                                    <Input
                                        type="file"
                                        id="logoFile"
                                        name="logoFile"
                                        accept="image/*"
                                        onChange={handleFileChange}
                                    />
                                    <small className="form-text text-muted">
                                        Рекомендуемый размер: 200x60px, форматы: JPG, PNG
                                    </small>
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Контактная информация</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="email">Email *</Label>
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
                                            <Label for="phone">Телефон</Label>
                                            <Input
                                                type="tel"
                                                id="phone"
                                                name="phone"
                                                value={formData.phone}
                                                onChange={handleInputChange}
                                                placeholder="+7 (xxx) xxx-xx-xx"
                                            />
                                        </FormGroup>
                                    </Col>
                                </Row>

                                <FormGroup>
                                    <Label for="address">Адрес офиса</Label>
                                    <Input
                                        type="textarea"
                                        id="address"
                                        name="address"
                                        value={formData.address}
                                        onChange={handleInputChange}
                                        rows="3"
                                        placeholder="Полный адрес офиса партнера"
                                    />
                                </FormGroup>

                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="telegram">Telegram</Label>
                                            <Input
                                                type="url"
                                                id="telegram"
                                                name="telegram"
                                                value={formData.telegram}
                                                onChange={handleInputChange}
                                                placeholder="https://t.me/username"
                                            />
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="whatsapp">WhatsApp</Label>
                                            <Input
                                                type="url"
                                                id="whatsapp"
                                                name="whatsapp"
                                                value={formData.whatsapp}
                                                onChange={handleInputChange}
                                                placeholder="https://wa.me/xxxxxxxxxx"
                                            />
                                        </FormGroup>
                                    </Col>
                                </Row>
                            </CardBody>
                        </Card>
                    </Col>

                    <Col lg="4">
                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Ссылки на маркетплейсы</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <FormGroup>
                                    <Label for="ozonLink">OZON</Label>
                                    <Input
                                        type="url"
                                        id="ozonLink"
                                        name="ozonLink"
                                        value={formData.ozonLink}
                                        onChange={handleInputChange}
                                        placeholder="https://ozon.ru/..."
                                    />
                                </FormGroup>

                                <FormGroup>
                                    <Label for="wildberriesLink">Wildberries</Label>
                                    <Input
                                        type="url"
                                        id="wildberriesLink"
                                        name="wildberriesLink"
                                        value={formData.wildberriesLink}
                                        onChange={handleInputChange}
                                        placeholder="https://wildberries.ru/..."
                                    />
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Настройки продаж</CardTitle>
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
                                        Разрешить покупку купонов через брендированную платформу
                                    </Label>
                                </FormGroup>
                            </CardBody>
                        </Card>

                        <Card className="main-card mb-3">
                            <CardHeader>
                                <CardTitle>Действия</CardTitle>
                            </CardHeader>
                            <CardBody>
                                <div className="d-flex flex-column gap-2">
                                    <Button 
                                        type="submit" 
                                        color="success" 
                                        size="lg" 
                                        block
                                        disabled={isSubmitting}
                                    >
                                        {isSubmitting ? (
                                            <>
                                                <i className="fa fa-spinner fa-spin"></i> Сохранение...
                                            </>
                                        ) : (
                                            <>
                                                <i className="pe-7s-check"></i> Сохранить изменения
                                            </>
                                        )}
                                    </Button>
                                    
                                    <Button 
                                        type="button" 
                                        color="secondary" 
                                        size="lg" 
                                        block
                                        onClick={() => navigate('/partners/list')}
                                        disabled={isSubmitting}
                                    >
                                        <i className="pe-7s-close"></i> Отменить
                                    </Button>
                                </div>
                            </CardBody>
                        </Card>
                    </Col>
                </Row>
            </Form>
        </Fragment>
    );
};

export default EditPartner; 