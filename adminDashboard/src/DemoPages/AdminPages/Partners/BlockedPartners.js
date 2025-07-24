import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText,
    Table, Badge
} from 'reactstrap';

const BlockedPartners = () => {
    const [searchTerm, setSearchTerm] = useState('');

    const mockBlockedPartners = [
        {
            id: 3,
            partnerCode: '1003',
            brandName: 'Творческая мастерская',
            domain: 'creative-workshop.ru',
            email: 'contact@creative-workshop.ru',
            phone: '+7 (499) 555-12-34',
            status: 'blocked',
            createdAt: '2023-10-22',
            blockedAt: '2023-12-15',
            blockedReason: 'Нарушение условий договора',
            couponsCreated: 200,
            couponsActivated: 156
        }
    ];

    const filteredPartners = mockBlockedPartners.filter(partner => 
        partner.brandName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        partner.domain.toLowerCase().includes(searchTerm.toLowerCase()) ||
        partner.email.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const handleUnblock = (partnerId) => {
        console.log('Разблокировать партнера:', partnerId);
        // Здесь будет API вызов для разблокировки
    };

    const handleDelete = (partnerId) => {
        if (window.confirm('Вы уверены, что хотите удалить этого партнера? Это действие необратимо.')) {
            console.log('Удалить партнера:', partnerId);
            // Здесь будет API вызов для удаления
        }
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-lock icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Заблокированные партнеры
                            <div className="page-title-subheading">
                                Список всех заблокированных партнеров системы
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="secondary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/partners/list'}
                        >
                            <i className="pe-7s-back"></i> Все партнеры
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Поиск</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Row>
                                <Col md="6">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder="Поиск по названию, домену, email..."
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="6">
                                    <div className="text-right">
                                        <small className="text-muted">
                                            Найдено: {filteredPartners.length}
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
                            <CardTitle>Заблокированные партнеры</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Код</th>
                                        <th>Название бренда</th>
                                        <th>Домен</th>
                                        <th>Email</th>
                                        <th>Дата блокировки</th>
                                        <th>Причина</th>
                                        <th>Купоны</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredPartners.map((partner) => (
                                        <tr key={partner.id}>
                                            <td>
                                                <Badge color="secondary">{partner.partnerCode}</Badge>
                                            </td>
                                            <td>
                                                <strong>{partner.brandName}</strong>
                                            </td>
                                            <td>{partner.domain}</td>
                                            <td>{partner.email}</td>
                                            <td>{partner.blockedAt}</td>
                                            <td>
                                                <small className="text-danger">
                                                    {partner.blockedReason}
                                                </small>
                                            </td>
                                            <td>
                                                <small>
                                                    Создано: <strong>{partner.couponsCreated}</strong><br/>
                                                    Активировано: <strong>{partner.couponsActivated}</strong>
                                                </small>
                                            </td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="success"
                                                        onClick={() => handleUnblock(partner.id)}
                                                        title="Разблокировать"
                                                    >
                                                        <i className="pe-7s-unlock"></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color="danger"
                                                        onClick={() => handleDelete(partner.id)}
                                                        title="Удалить"
                                                    >
                                                        <i className="pe-7s-trash"></i>
                                                    </Button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>

                            {filteredPartners.length === 0 && (
                                <div className="text-center py-4">
                                    <i className="pe-7s-check text-success" style={{fontSize: '3em'}}></i>
                                    <p className="text-muted mt-3">Заблокированных партнеров не найдено</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default BlockedPartners; 