import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert
} from 'reactstrap';

const CreatePartner = () => {
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
        allowSales: true
    });

    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitMessage, setSubmitMessage] = useState('');

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
        if (!formData.password) newErrors.password = 'Пароль обязателен';
        if (formData.password !== formData.confirmPassword) {
            newErrors.confirmPassword = 'Пароли не совпадают';
        }
        if (!formData.domain.trim()) newErrors.domain = 'Доменное имя обязательно';
        if (!formData.brandName.trim()) newErrors.brandName = 'Название бренда обязательно';
        if (!formData.email.trim()) newErrors.email = 'Email обязателен';

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (formData.email && !emailRegex.test(formData.email)) {
            newErrors.email = 'Некорректный формат email';
        }

        const domainRegex = /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$/;
        if (formData.domain && !domainRegex.test(formData.domain)) {
            newErrors.domain = 'Некорректный формат домена';
        }

        if (formData.password && formData.password.length < 6) {
            newErrors.password = 'Пароль должен содержать минимум 6 символов';
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
            console.log('Создание партнера:', formData);
            
            await new Promise(resolve => setTimeout(resolve, 2000));
            
            setSubmitMessage('Партнер успешно создан!');
            
            setTimeout(() => {
                window.location.href = '/#/partners/list';
            }, 2000);
            
        } catch (error) {
            setSubmitMessage('Ошибка при создании партнера: ' + error.message);
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
                        <div>Добавить партнера
                            <div className="page-title-subheading">
                                Создание нового партнера в системе White Label
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/partners/list'}
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

                                <Row>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="password">Пароль *</Label>
                                            <Input
                                                type="password"
                                                id="password"
                                                name="password"
                                                value={formData.password}
                                                onChange={handleInputChange}
                                                invalid={!!errors.password}
                                                placeholder="Минимум 6 символов"
                                            />
                                            <FormFeedback>{errors.password}</FormFeedback>
                                        </FormGroup>
                                    </Col>
                                    <Col md="6">
                                        <FormGroup>
                                            <Label for="confirmPassword">Подтверждение пароля *</Label>
                                            <Input
                                                type="password"
                                                id="confirmPassword"
                                                name="confirmPassword"
                                                value={formData.confirmPassword}
                                                onChange={handleInputChange}
                                                invalid={!!errors.confirmPassword}
                                                placeholder="Повторите пароль"
                                            />
                                            <FormFeedback>{errors.confirmPassword}</FormFeedback>
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
                                    <small className="form-text text-muted">
                                        Домен для White Label версии сайта
                                    </small>
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
                                <small className="form-text text-muted">
                                    Если отключено, пользователи не смогут покупать купоны напрямую через сайт партнера
                                </small>
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
                                                <i className="fa fa-spinner fa-spin"></i> Создание...
                                            </>
                                        ) : (
                                            <>
                                                <i className="pe-7s-check"></i> Создать партнера
                                            </>
                                        )}
                                    </Button>
                                    
                                    <Button 
                                        type="button" 
                                        color="secondary" 
                                        size="lg" 
                                        block
                                        onClick={() => window.location.href = '/#/partners/list'}
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

export default CreatePartner; 