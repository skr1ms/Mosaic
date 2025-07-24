import React, { Fragment, useState } from 'react';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Input, InputGroup, InputGroupText, Table, Badge,
    Dropdown, DropdownToggle, DropdownMenu, DropdownItem,
    Modal, ModalHeader, ModalBody, ModalFooter, UncontrolledTooltip
} from 'reactstrap';

const ManageCoupons = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [partnerFilter, setPartnerFilter] = useState('all');
    const [sizeFilter, setSizeFilter] = useState('all');
    const [styleFilter, setStyleFilter] = useState('all');
    const [isStatusFilterOpen, setIsStatusFilterOpen] = useState(false);
    const [isPartnerFilterOpen, setIsPartnerFilterOpen] = useState(false);
    const [isSizeFilterOpen, setIsSizeFilterOpen] = useState(false);
    const [isStyleFilterOpen, setIsStyleFilterOpen] = useState(false);
    const [selectedCoupon, setSelectedCoupon] = useState(null);
    const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);

    const mockPartners = [
        { id: '0000', name: 'Собственные' },
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

    const mockCoupons = [
        {
            id: 1,
            code: '1001-2345-6789',
            partnerId: '1001',
            partnerName: 'Мозаика Арт',
            size: '30x40',
            style: 'max_colors',
            status: 'new',
            createdAt: '2023-12-01T10:30:00',
            activatedAt: null,
            originalImageUrl: null,
            previewUrl: null,
            schemaUrl: null,
            userEmail: null
        },
        {
            id: 2,
            code: '1001-9876-5432',
            partnerId: '1001',
            partnerName: 'Мозаика Арт',
            size: '40x50',
            style: 'pop_art',
            status: 'used',
            createdAt: '2023-11-15T14:20:00',
            activatedAt: '2023-12-10T16:45:00',
            originalImageUrl: '/uploads/original_123.jpg',
            previewUrl: '/uploads/preview_123.jpg',
            schemaUrl: '/uploads/schema_123.pdf',
            userEmail: 'user@example.com'
        },
        {
            id: 3,
            code: '0000-1111-2222',
            partnerId: '0000',
            partnerName: 'Собственный',
            size: '21x30',
            style: 'grayscale',
            status: 'used',
            createdAt: '2023-10-22T09:15:00',
            activatedAt: '2023-11-05T12:30:00',
            originalImageUrl: '/uploads/original_456.jpg',
            previewUrl: '/uploads/preview_456.jpg',
            schemaUrl: '/uploads/schema_456.pdf',
            userEmail: 'test@test.com'
        }
    ];

    const filteredCoupons = mockCoupons.filter(coupon => {
        const matchesSearch = coupon.code.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            coupon.userEmail?.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || coupon.status === statusFilter;
        const matchesPartner = partnerFilter === 'all' || coupon.partnerId === partnerFilter;
        const matchesSize = sizeFilter === 'all' || coupon.size === sizeFilter;
        const matchesStyle = styleFilter === 'all' || coupon.style === styleFilter;
        
        return matchesSearch && matchesStatus && matchesPartner && matchesSize && matchesStyle;
    });

    const getStatusBadge = (status) => {
        return status === 'new' ? 
            <Badge color="success">Новый</Badge> : 
            <Badge color="primary">Погашенный</Badge>;
    };

    const formatDate = (dateString) => {
        if (!dateString) return '-';
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU') + ' ' + date.toLocaleTimeString('ru-RU', {hour: '2-digit', minute: '2-digit'});
    };

    const handleViewDetails = (coupon) => {
        setSelectedCoupon(coupon);
        setIsDetailModalOpen(true);
    };

    const handleReset = (couponId) => {
        if (window.confirm('Вы уверены, что хотите сбросить этот купон? Все данные активации будут удалены.')) {
            console.log('Сброс купона:', couponId);
        }
    };

    const handleDelete = (couponId) => {
        if (window.confirm('Вы уверены, что хотите удалить этот купон? Это действие необратимо.')) {
            console.log('Удаление купона:', couponId);
        }
    };

    const handleDownloadMaterials = (coupon) => {
        console.log('Скачивание материалов для купона:', coupon.code);
    };

    const exportCoupons = () => {
        const csvContent = [
            'Номер купона,Партнер,Размер,Стиль,Статус,Дата создания,Дата активации',
            ...filteredCoupons.map(coupon => 
                `${coupon.code},${coupon.partnerName},${sizeOptions.find(s => s.value === coupon.size)?.label},${styleOptions.find(s => s.value === coupon.style)?.label},${coupon.status === 'new' ? 'Новый' : 'Погашенный'},${formatDate(coupon.createdAt)},${formatDate(coupon.activatedAt)}`
            )
        ].join('\n');

        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `coupons_export_${Date.now()}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    return (
        <Fragment>
            <div className="app-page-title">
                <div className="page-title-wrapper">
                    <div className="page-title-heading">
                        <div className="page-title-icon">
                            <i className="pe-7s-search icon-gradient bg-mean-fruit">
                            </i>
                        </div>
                        <div>Управление купонами
                            <div className="page-title-subheading">
                                Поиск, просмотр и управление всеми купонами в системе
                            </div>
                        </div>
                    </div>
                    <div className="page-title-actions">
                        <Button 
                            color="success" 
                            size="lg"
                            onClick={exportCoupons}
                        >
                            <i className="pe-7s-download"></i> Экспорт списка
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
                            <Row className="mb-3">
                                <Col md="4">
                                    <InputGroup>
                                        <InputGroupText>
                                            <i className="pe-7s-search"></i>
                                        </InputGroupText>
                                        <Input
                                            type="text"
                                            placeholder="Поиск по номеру купона или email..."
                                            value={searchTerm}
                                            onChange={(e) => setSearchTerm(e.target.value)}
                                        />
                                    </InputGroup>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isStatusFilterOpen} 
                                        toggle={() => setIsStatusFilterOpen(!isStatusFilterOpen)}
                                    >
                                        <DropdownToggle caret color="info" block>
                                            {statusFilter === 'all' ? 'Все статусы' : 
                                             statusFilter === 'new' ? 'Новые' : 'Погашенные'}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStatusFilter('all')}>
                                                Все статусы
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('new')}>
                                                Новые
                                            </DropdownItem>
                                            <DropdownItem onClick={() => setStatusFilter('used')}>
                                                Погашенные
                                            </DropdownItem>
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isPartnerFilterOpen} 
                                        toggle={() => setIsPartnerFilterOpen(!isPartnerFilterOpen)}
                                    >
                                        <DropdownToggle caret color="secondary" block>
                                            {partnerFilter === 'all' ? 'Все партнеры' : 
                                             mockPartners.find(p => p.id === partnerFilter)?.name}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setPartnerFilter('all')}>
                                                Все партнеры
                                            </DropdownItem>
                                            {mockPartners.map(partner => (
                                                <DropdownItem 
                                                    key={partner.id}
                                                    onClick={() => setPartnerFilter(partner.id)}
                                                >
                                                    {partner.name}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isSizeFilterOpen} 
                                        toggle={() => setIsSizeFilterOpen(!isSizeFilterOpen)}
                                    >
                                        <DropdownToggle caret color="warning" block>
                                            {sizeFilter === 'all' ? 'Все размеры' : 
                                             sizeOptions.find(s => s.value === sizeFilter)?.label}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setSizeFilter('all')}>
                                                Все размеры
                                            </DropdownItem>
                                            {sizeOptions.map(size => (
                                                <DropdownItem 
                                                    key={size.value}
                                                    onClick={() => setSizeFilter(size.value)}
                                                >
                                                    {size.label}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                                <Col md="2">
                                    <Dropdown 
                                        isOpen={isStyleFilterOpen} 
                                        toggle={() => setIsStyleFilterOpen(!isStyleFilterOpen)}
                                    >
                                        <DropdownToggle caret color="primary" block>
                                            {styleFilter === 'all' ? 'Все стили' : 
                                             styleOptions.find(s => s.value === styleFilter)?.label}
                                        </DropdownToggle>
                                        <DropdownMenu>
                                            <DropdownItem onClick={() => setStyleFilter('all')}>
                                                Все стили
                                            </DropdownItem>
                                            {styleOptions.map(style => (
                                                <DropdownItem 
                                                    key={style.value}
                                                    onClick={() => setStyleFilter(style.value)}
                                                >
                                                    {style.label}
                                                </DropdownItem>
                                            ))}
                                        </DropdownMenu>
                                    </Dropdown>
                                </Col>
                            </Row>
                            <Row>
                                <Col md="12">
                                    <div className="d-flex justify-content-between align-items-center">
                                        <small className="text-muted">
                                            Найдено купонов: <strong>{filteredCoupons.length}</strong>
                                        </small>
                                        <Button 
                                            color="light" 
                                            size="sm"
                                            onClick={() => {
                                                setSearchTerm('');
                                                setStatusFilter('all');
                                                setPartnerFilter('all');
                                                setSizeFilter('all');
                                                setStyleFilter('all');
                                            }}
                                        >
                                            <i className="pe-7s-refresh"></i> Сбросить фильтры
                                        </Button>
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
                            <CardTitle>Список купонов</CardTitle>
                        </CardHeader>
                        <CardBody>
                            <Table responsive hover>
                                <thead>
                                    <tr>
                                        <th>Номер купона</th>
                                        <th>Партнер</th>
                                        <th>Размер</th>
                                        <th>Стиль</th>
                                        <th>Статус</th>
                                        <th>Дата создания</th>
                                        <th>Дата активации</th>
                                        <th>Действия</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {filteredCoupons.map((coupon) => (
                                        <tr key={coupon.id}>
                                            <td>
                                                <Badge color="secondary" className="badge-pill">
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
                                            <td>{sizeOptions.find(s => s.value === coupon.size)?.label}</td>
                                            <td>{styleOptions.find(s => s.value === coupon.style)?.label}</td>
                                            <td>{getStatusBadge(coupon.status)}</td>
                                            <td>
                                                <small>{formatDate(coupon.createdAt)}</small>
                                            </td>
                                            <td>
                                                <small>{formatDate(coupon.activatedAt)}</small>
                                            </td>
                                            <td>
                                                <div className="btn-group" role="group">
                                                    <Button 
                                                        size="sm" 
                                                        color="info"
                                                        onClick={() => handleViewDetails(coupon)}
                                                        id={`details-${coupon.id}`}
                                                    >
                                                        <i className="pe-7s-look"></i>
                                                    </Button>
                                                    <UncontrolledTooltip target={`details-${coupon.id}`}>
                                                        Просмотр деталей
                                                    </UncontrolledTooltip>

                                                    {coupon.status === 'used' && (
                                                        <>
                                                            <Button 
                                                                size="sm" 
                                                                color="success"
                                                                onClick={() => handleDownloadMaterials(coupon)}
                                                                id={`download-${coupon.id}`}
                                                            >
                                                                <i className="pe-7s-download"></i>
                                                            </Button>
                                                            <UncontrolledTooltip target={`download-${coupon.id}`}>
                                                                Скачать материалы
                                                            </UncontrolledTooltip>
                                                        </>
                                                    )}

                                                    {coupon.status === 'used' && (
                                                        <>
                                                            <Button 
                                                                size="sm" 
                                                                color="warning"
                                                                onClick={() => handleReset(coupon.id)}
                                                                id={`reset-${coupon.id}`}
                                                            >
                                                                <i className="pe-7s-refresh"></i>
                                                            </Button>
                                                            <UncontrolledTooltip target={`reset-${coupon.id}`}>
                                                                Сбросить купон
                                                            </UncontrolledTooltip>
                                                        </>
                                                    )}

                                                    <Button 
                                                        size="sm" 
                                                        color="danger"
                                                        onClick={() => handleDelete(coupon.id)}
                                                        id={`delete-${coupon.id}`}
                                                    >
                                                        <i className="pe-7s-trash"></i>
                                                    </Button>
                                                    <UncontrolledTooltip target={`delete-${coupon.id}`}>
                                                        Удалить купон
                                                    </UncontrolledTooltip>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>

                            {filteredCoupons.length === 0 && (
                                <div className="text-center py-4">
                                    <p className="text-muted">Купоны не найдены</p>
                                </div>
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>

            <Modal isOpen={isDetailModalOpen} toggle={() => setIsDetailModalOpen(false)} size="lg">
                <ModalHeader toggle={() => setIsDetailModalOpen(false)}>
                    Детали купона: {selectedCoupon?.code}
                </ModalHeader>
                <ModalBody>
                    {selectedCoupon && (
                        <Row>
                            <Col md="6">
                                <h6>Основная информация</h6>
                                <table className="table table-sm">
                                    <tbody>
                                        <tr>
                                            <td><strong>Номер купона:</strong></td>
                                            <td>{selectedCoupon.code}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>Партнер:</strong></td>
                                            <td>{selectedCoupon.partnerName} ({selectedCoupon.partnerId})</td>
                                        </tr>
                                        <tr>
                                            <td><strong>Размер:</strong></td>
                                            <td>{sizeOptions.find(s => s.value === selectedCoupon.size)?.label}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>Стиль:</strong></td>
                                            <td>{styleOptions.find(s => s.value === selectedCoupon.style)?.label}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>Статус:</strong></td>
                                            <td>{getStatusBadge(selectedCoupon.status)}</td>
                                        </tr>
                                        <tr>
                                            <td><strong>Дата создания:</strong></td>
                                            <td>{formatDate(selectedCoupon.createdAt)}</td>
                                        </tr>
                                        {selectedCoupon.status === 'used' && (
                                            <tr>
                                                <td><strong>Дата активации:</strong></td>
                                                <td>{formatDate(selectedCoupon.activatedAt)}</td>
                                            </tr>
                                        )}
                                    </tbody>
                                </table>
                            </Col>
                            <Col md="6">
                                {selectedCoupon.status === 'used' ? (
                                    <>
                                        <h6>Информация об активации</h6>
                                        <table className="table table-sm">
                                            <tbody>
                                                <tr>
                                                    <td><strong>Email пользователя:</strong></td>
                                                    <td>{selectedCoupon.userEmail}</td>
                                                </tr>
                                                <tr>
                                                    <td><strong>Оригинальное изображение:</strong></td>
                                                    <td>
                                                        {selectedCoupon.originalImageUrl ? (
                                                            <a href={selectedCoupon.originalImageUrl} target="_blank" rel="noopener noreferrer">
                                                                Просмотреть
                                                            </a>
                                                        ) : 'Недоступно'}
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <td><strong>Превью мозаики:</strong></td>
                                                    <td>
                                                        {selectedCoupon.previewUrl ? (
                                                            <a href={selectedCoupon.previewUrl} target="_blank" rel="noopener noreferrer">
                                                                Просмотреть
                                                            </a>
                                                        ) : 'Недоступно'}
                                                    </td>
                                                </tr>
                                                <tr>
                                                    <td><strong>Схема мозаики:</strong></td>
                                                    <td>
                                                        {selectedCoupon.schemaUrl ? (
                                                            <a href={selectedCoupon.schemaUrl} target="_blank" rel="noopener noreferrer">
                                                                Скачать PDF
                                                            </a>
                                                        ) : 'Недоступно'}
                                                    </td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </>
                                ) : (
                                    <div className="text-center text-muted py-4">
                                        <i className="pe-7s-info" style={{fontSize: '2em'}}></i>
                                        <p>Купон ещё не был активирован</p>
                                    </div>
                                )}
                            </Col>
                        </Row>
                    )}
                </ModalBody>
                <ModalFooter>
                    <Button color="secondary" onClick={() => setIsDetailModalOpen(false)}>
                        Закрыть
                    </Button>
                </ModalFooter>
            </Modal>
        </Fragment>
    );
};

export default ManageCoupons; 