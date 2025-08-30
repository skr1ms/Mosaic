import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Row, Col, Card, CardBody, CardTitle, CardHeader,
    Button, Form, FormGroup, Label, Input, FormFeedback,
    Alert, Table, Badge
} from 'reactstrap';
import api from '../../../api/api';

const ArticleGrid = ({ partnerId, isReadOnly = false }) => {
    const { t } = useTranslation();
    const [articleGrid, setArticleGrid] = useState({});
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    // Константы для размеров и стилей
    const sizes = ['20x20', '30x40', '40x40', '40x50', '40x60', '50x70'];
    const styles = ['grayscale', 'skin_tones', 'pop_art', 'max_colors'];
    const marketplaces = ['ozon', 'wildberries'];

    useEffect(() => {
        if (partnerId) {
            loadArticleGrid();
        }
    }, [partnerId]);

    const loadArticleGrid = async () => {
        try {
            setLoading(true);
            const response = await api.get(`/admin/partners/${partnerId}/articles/grid`);
            setArticleGrid(response.data);
        } catch (err) {
            console.error('Failed to load article grid:', err);
            setError(t('partners.failed_to_load_article_grid'));
        } finally {
            setLoading(false);
        }
    };

    const updateSKU = async (marketplace, style, size, sku) => {
        try {
            setSaving(true);
            await api.put(`/admin/partners/${partnerId}/articles/sku`, {
                marketplace,
                style,
                size,
                sku
            });

            // Обновляем локальное состояние
            setArticleGrid(prev => ({
                ...prev,
                [marketplace]: {
                    ...prev[marketplace],
                    [style]: {
                        ...prev[marketplace]?.[style],
                        [size]: sku
                    }
                }
            }));

            setSuccess(t('partners.sku_updated_successfully'));
            setTimeout(() => setSuccess(''), 3000);
        } catch (err) {
            console.error('Failed to update SKU:', err);
            setError(t('partners.failed_to_update_sku'));
            setTimeout(() => setError(''), 5000);
        } finally {
            setSaving(false);
        }
    };

    const handleSKUChange = (marketplace, style, size, value) => {
        // Обновляем локальное состояние немедленно для UI
        setArticleGrid(prev => ({
            ...prev,
            [marketplace]: {
                ...prev[marketplace],
                [style]: {
                    ...prev[marketplace]?.[style],
                    [size]: value
                }
            }
        }));
    };

    const handleSKUBlur = (marketplace, style, size, value) => {
        // Сохраняем при потере фокуса
        updateSKU(marketplace, style, size, value);
    };

    const getMarketplaceDisplayName = (marketplace) => {
        return marketplace === 'ozon' ? 'OZON' : 'Wildberries';
    };

    const getStyleDisplayName = (style) => {
        const styleNames = {
            'grayscale': t('partners.style_grayscale'),
            'skin_tones': t('partners.style_skin_tones'),
            'pop_art': t('partners.style_pop_art'),
            'max_colors': t('partners.style_max_colors')
        };
        return styleNames[style] || style;
    };

    if (loading) {
        return (
            <Card className="main-card mb-3">
                <CardBody className="text-center">
                    <i className="fa fa-spinner fa-spin fa-2x"></i>
                    <p className="mt-2">{t('common.loading')}...</p>
                </CardBody>
            </Card>
        );
    }

    return (
        <Card className="main-card mb-3">
            <CardHeader>
                <CardTitle>{t('partners.article_grid')}</CardTitle>
                <small className="text-muted">{t('partners.article_grid_description')}</small>
            </CardHeader>
            <CardBody>
                {error && (
                    <Alert color="danger" className="mb-3">
                        {error}
                    </Alert>
                )}
                {success && (
                    <Alert color="success" className="mb-3">
                        {success}
                    </Alert>
                )}

                {marketplaces.map(marketplace => (
                    <div key={marketplace} className="mb-4">
                        <h5 className="mb-3">
                            <Badge color={marketplace === 'ozon' ? 'warning' : 'info'} className="me-2">
                                {getMarketplaceDisplayName(marketplace)}
                            </Badge>
                            {t('partners.article_grid_for_marketplace')}
                        </h5>
                        
                        <div className="table-responsive">
                            <Table bordered hover size="sm">
                                <thead>
                                    <tr>
                                        <th className="bg-light" style={{ minWidth: '120px' }}>
                                            {t('partners.style')} / {t('partners.size')}
                                        </th>
                                        {sizes.map(size => (
                                            <th key={size} className="text-center bg-light">
                                                {size}
                                            </th>
                                        ))}
                                    </tr>
                                </thead>
                                <tbody>
                                    {styles.map(style => (
                                        <tr key={style}>
                                            <td className="bg-light fw-bold">
                                                {getStyleDisplayName(style)}
                                            </td>
                                            {sizes.map(size => (
                                                <td key={size} className="p-1">
                                                    <Input
                                                        type="text"
                                                        value={articleGrid[marketplace]?.[style]?.[size] || ''}
                                                        onChange={(e) => handleSKUChange(marketplace, style, size, e.target.value)}
                                                        onBlur={(e) => handleSKUBlur(marketplace, style, size, e.target.value)}
                                                        placeholder={t('partners.enter_sku')}
                                                        disabled={isReadOnly || saving}
                                                        size="sm"
                                                        className="text-center"
                                                    />
                                                </td>
                                            ))}
                                        </tr>
                                    ))}
                                </tbody>
                            </Table>
                        </div>
                    </div>
                ))}

                <div className="mt-3">
                    <small className="text-muted">
                        <i className="fa fa-info-circle me-1"></i>
                        {t('partners.article_grid_help')}
                    </small>
                </div>
            </CardBody>
        </Card>
    );
};

export default ArticleGrid;
