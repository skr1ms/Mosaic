import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText, Table, Badge
} from 'reactstrap';

const ActivatedCoupons = () => {
    const [searchTerm, setSearchTerm] = useState('');

    const mockActivatedCoupons = [
        {
            id: 1,
            code: '1001-2345-6789',
            partnerId: '1001',
            partnerName: 'Мозаика Арт',
            size: '30x40',
            style: 'max_colors',
            createdAt: '2023-11-15T14:20:00',
            activatedAt: '2023-12-10T16:45:00',
            userEmail: 'user@example.com',
            hasFiles: true
        },
        {
            id: 2,
            code: '0000-1111-2222',
            partnerId: '0000',
            partnerName: 'Собственный',
            size: '21x30',
            style: 'grayscale',
            createdAt: '2023-10-22T09:15:00',
            activatedAt: '2023-11-05T12:30:00',
            userEmail: 'test@test.com',
            hasFiles: true
        }
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

    const filteredCoupons = mockActivatedCoupons.filter(coupon =>
        coupon.code.toLowerCase().includes(searchTerm.toLowerCase()) ||
        coupon.userEmail?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        coupon.partnerName.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU') + ' ' + date.toLocaleTimeString('ru-RU', {hour: '2-digit', minute: '2-digit'});
    };

    const handleDownloadMaterials = (coupon) => {
        console.log('Скачивание материалов для купона:', coupon.code);
        // Здесь будет логика скачивания
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-check icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Активированные купоны
                            <div className="page-title-subheading">
                                Список всех погашенных купонов с материалами пользователей
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/coupons/manage'}
                        >
                            <i className="pe-7s-back"></i> Все купоны
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Поиск активированных купонов</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row>
                                <Col md="8">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder="Поиск по номеру купона, email пользователя, партнеру..."
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="4">
                                    <div className="text-right">
                                        <small className="text-muted">
                                            Найдено: <strong>{filteredCoupons.length}</strong> активированных купонов
                                        </small>
                                    </div>
                                </Col>
                            </Row>
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Список активированных купонов</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Номер купона</th>
                                        <th>Партнер</th>
                                        <th>Размер / Стиль</th>
                                        <th>Email пользователя</th>
                                        <th>Дата создания</th>
                                        <th>Дата активации</th>
                                        <th>Материалы</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredCoupons.map((coupon) => (
                                        <tr key={coupon.id}>
                                            <td>
                                                <Badge color="primary" className="badge-pill">
                                                    {coupon.code}
                                                </Badge>
                                            </td>
                                            <td>
                                                <strong>{coupon.partnerName}</strong>
                                                <br/>
                                                <small className="text-muted">
                                                    Код: {coupon.partnerId}
                                                </small>
                                            </td>
                                            <td>
                                                <div>
                                                    <Badge color="info">
                                                        {sizeOptions.find(s => s.value === coupon.size)?.label}
                                                    </Badge>
                                                </div>
                                                <small className="text-muted">
                                                    {styleOptions.find(s => s.value === coupon.style)?.label}
                                                </small>
                                            </td>
                                            <td>
                                                <strong>{coupon.userEmail}</strong>
                                            </td>
                                            <td>
                                                <small>{formatDate(coupon.createdAt)}</small>
                                            </td>
                                            <td>
                                                <small className="text-success">
                                                    <strong>{formatDate(coupon.activatedAt)}</strong>
                                                </small>
                                            </td>
                                            <td>
                                                {coupon.hasFiles ? (
                                                    <Badge color="success">
                                                        <i className="pe-7s-check"></i> Есть
                                                    </Badge>
                                                ) : (
                                                    <Badge color="warning">
                                                        <i className="pe-7s-close"></i> Нет
                                                    </Badge>
                                                )}
                                            </td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="info"
                                                        title="Просмотр деталей"
                                                    >
                                                        <i className="pe-7s-look"></i>
                                                    </Button>
                                                    {coupon.hasFiles && (
                                                        <Button 
                                                            size="sm" 
                                                            color="success"
                                                            onClick={() => handleDownloadMaterials(coupon)}
                                                            title="Скачать материалы"
                                                        >
                                                            <i className="pe-7s-download"></i>
                                                        </Button>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>

                            {filteredCoupons.length === 0 && (
                                <div className="text-center py-4">
                                    <i className="pe-7s-info text-muted" style={{fontSize: '3em'}}></i>
                                    <p className="text-muted mt-3">Активированные купоны не найдены</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default ActivatedCoupons; 