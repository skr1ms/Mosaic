import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText,
    Table, Badge, Dropdown, DropdownToggle, DropdownMenu, DropdownItem
} from 'reactstrap';

const PartnersList = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [isFilterOpen, setIsFilterOpen] = useState(false);

    const mockPartners = [
        {
            id: 1,
            partnerCode: '1001',
            brandName: 'Мозаика Арт',
            domain: 'mosaic-art.ru',
            email: 'admin@mosaic-art.ru',
            phone: '+7 (495) 123-45-67',
            status: 'active',
            createdAt: '2023-12-01',
            couponsCreated: 150,
            couponsActivated: 89
        },
        {
            id: 2,
            partnerCode: '1002',
            brandName: 'Алмазная студия',
            domain: 'diamond-studio.com',
            email: 'info@diamond-studio.com',
            phone: '+7 (812) 987-65-43',
            status: 'active',
            createdAt: '2023-11-15',
            couponsCreated: 75,
            couponsActivated: 42
        },
        {
            id: 3,
            partnerCode: '1003',
            brandName: 'Творческая мастерская',
            domain: 'creative-workshop.ru',
            email: 'contact@creative-workshop.ru',
            phone: '+7 (499) 555-12-34',
            status: 'blocked',
            createdAt: '2023-10-22',
            couponsCreated: 200,
            couponsActivated: 156
        }
    ];

    const filteredPartners = mockPartners.filter(partner => {
        const matchesSearch = partner.brandName.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            partner.domain.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            partner.email.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || partner.status === statusFilter;
        
        return matchesSearch && matchesStatus;
    });

    const getStatusBadge = (status) => {
        return status === 'active' ? 
            <Badge color="success">Активный</Badge> : 
            <Badge color="danger">Заблокирован</Badge>;
    };

    const handleEdit = (partnerId) => {
        window.location.href = `/#/partners/edit/${partnerId}`;
    };

    const handleBlock = (partnerId) => {
        console.log('Заблокировать партнера:', partnerId);
    };

    const handleDelete = (partnerId) => {
        if (window.confirm('Вы уверены, что хотите удалить этого партнера? Это действие необратимо.')) {
            console.log('Удалить партнера:', partnerId);
        }
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-users icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Управление партнерами
                            <div className="page-title-subheading">
                                Список и управление всеми партнерами системы
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="primary" 
                            size="lg"
                            onClick={() => window.location.href = '/#/partners/create'}
                        >
                            <i className="pe-7s-plus"></i> Добавить партнера
                        </Button>
                    </div>
                </div>
            </div>

            <Row>
                <Col lg="12">
                    <Card className="main-card mb-3">
                        <CardHeader>
                            <CardTitle>Фильтры и поиск</CardTitle>
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
                                <Col md="4">
                                    <Dropdown 
                                        isOpen={isFilterOpen} 
                                        toggle={() => setIsFilterOpen(!isFilterOpen)}
                                    >
                                        <DropdownToggle caret color="info">
                                            Статус: {statusFilter === 'all' ? 'Все' : 
                                                    statusFilter === 'active' ? 'Активные' : 'Заблокированные'}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStatusFilter('all')}>
                                                Все
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('active')}>
                                                Активные
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('blocked')}>
                                                Заблокированные
                                            </DropdownItem>
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
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
                            <CardTitle>Список партнеров</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Код</th>
                                        <th>Название бренда</th>
                                        <th>Домен</th>
                                        <th>Email</th>
                                        <th>Телефон</th>
                                        <th>Статус</th>
                                        <th>Купоны</th>
                                        <th>Дата создания</th>
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
                                            <td>
                                                <a href={`https://${partner.domain}`} target="_blank" rel="noopener noreferrer">
                                                    {partner.domain}
                                                </a>
                                            </td>
                                            <td>{partner.email}</td>
                                            <td>{partner.phone}</td>
                                            <td>{getStatusBadge(partner.status)}</td>
                                            <td>
                                                <small>
                                                    Создано: <strong>{partner.couponsCreated}</strong><br/>
                                                    Активировано: <strong>{partner.couponsActivated}</strong>
                                                </small>
                                            </td>
                                            <td>{partner.createdAt}</td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="primary"
                                                        onClick={() => handleEdit(partner.id)}
                                                    >
                                                        <i className="pe-7s-edit"></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color={partner.status === 'active' ? 'warning' : 'success'}
                                                        onClick={() => handleBlock(partner.id)}
                                                    >
                                                        <i className={partner.status === 'active' ? 'pe-7s-lock' : 'pe-7s-unlock'}></i>
                                                    </Button>
                                                    <Button 
                                                        size="sm" 
                                                        color="danger"
                                                        onClick={() => handleDelete(partner.id)}
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
                                    <p className="text-muted">Партнеры не найдены</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        </Fragment>
    );
};

export default PartnersList; 