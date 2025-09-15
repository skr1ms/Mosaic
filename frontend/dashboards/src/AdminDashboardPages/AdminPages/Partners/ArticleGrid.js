import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
    Card, CardBody, CardTitle, CardHeader,
    Input, Alert, Table, Badge
} from 'reactstrap';
import api from '../../../api/api';

const ArticleGrid = ({ partnerId, isReadOnly = false }) => {
    const { t } = useTranslation();
    const [articleGrid, setArticleGrid] = useState({});
    const [loading, setLoading] = useState(true);
    const [saving] = useState(false);
    const [error, setError] = useState('');
    const [focusedCell, setFocusedCell] = useState(null);

    // Константы для размеров и стилей
    const sizes = ['21x30', '30x40', '40x40', '40x50', '40x60', '50x70'];
    const styles = ['grayscale', 'skin_tones', 'pop_art', 'max_colors'];
    const marketplaces = ['ozon', 'wildberries'];

    const loadArticleGrid = useCallback(async () => {
        try {
            setLoading(true);
            const response = await api.get(`/admin/partners/${partnerId}/articles/grid`);
            let loadedGrid = response.data;
            
            
            const localChanges = localStorage.getItem(`partner_articles_${partnerId}`);
            if (localChanges) {
                try {
                    const changes = JSON.parse(localChanges);
                    console.log('Found localStorage changes:', changes);
                    
                    
                    Object.entries(changes).forEach(([cellKey, sku]) => {
                        const [marketplace, style, size] = cellKey.split('-');
                        if (!loadedGrid[marketplace]) loadedGrid[marketplace] = {};
                        if (!loadedGrid[marketplace][style]) loadedGrid[marketplace][style] = {};
                        loadedGrid[marketplace][style][size] = sku;
                    });
                } catch (parseError) {
                    console.error('Error parsing localStorage changes:', parseError);
                }
            }
            
            setArticleGrid(loadedGrid);
        } catch (err) {
            console.error('Failed to load article grid:', err);
            setError(t('partners.failed_to_load_article_grid'));
        } finally {
            setLoading(false);
        }
    }, [partnerId, t]);

    useEffect(() => {
        if (partnerId && !isReadOnly) {
            loadArticleGrid();
        } else {
            
            setLoading(false);
        }
        
    }, [partnerId, isReadOnly, loadArticleGrid]);

    const handleSKUChange = (marketplace, style, size, value) => {
        
        const newGrid = {
            ...articleGrid,
            [marketplace]: {
                ...articleGrid[marketplace],
                [style]: {
                    ...articleGrid[marketplace]?.[style],
                    [size]: value
                }
            }
        };

        setArticleGrid(newGrid);

        
        if (isCreationMode) {
            localStorage.setItem('create_partner_articles', JSON.stringify(newGrid));
        } else if (partnerId) {
            
            const currentArticles = JSON.parse(localStorage.getItem(`partner_articles_${partnerId}`) || '{}');
            const cellKey = `${marketplace}-${style}-${size}`;
            
            currentArticles[cellKey] = value ? value.trim() : '';
            localStorage.setItem(`partner_articles_${partnerId}`, JSON.stringify(currentArticles));
            console.log('Saved to localStorage:', `partner_articles_${partnerId}`, currentArticles);
        }
    };

    const handleSKUBlur = (marketplace, style, size, value) => {
        handleSKUChange(marketplace, style, size, value);
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

    
    const getNextCell = (currentMarketplace, currentStyle, currentSize, direction) => {
        const marketplaceIndex = marketplaces.findIndex(m => m === currentMarketplace);
        const styleIndex = styles.findIndex(s => s === currentStyle);
        const sizeIndex = sizes.findIndex(s => s === currentSize);

        let newMarketplace = currentMarketplace;
        let newStyle = currentStyle;
        let newSize = currentSize;

        switch (direction) {
            case 'ArrowUp':
                if (styleIndex > 0) {
                    newStyle = styles[styleIndex - 1];
                } else if (marketplaceIndex > 0) {
                    newMarketplace = marketplaces[marketplaceIndex - 1];
                    newStyle = styles[styles.length - 1];
                }
                break;
            case 'ArrowDown':
                if (styleIndex < styles.length - 1) {
                    newStyle = styles[styleIndex + 1];
                } else if (marketplaceIndex < marketplaces.length - 1) {
                    newMarketplace = marketplaces[marketplaceIndex + 1];
                    newStyle = styles[0];
                }
                break;
            case 'ArrowLeft':
                if (sizeIndex > 0) {
                    newSize = sizes[sizeIndex - 1];
                } else if (styleIndex > 0) {
                    newStyle = styles[styleIndex - 1];
                    newSize = sizes[sizes.length - 1];
                } else if (marketplaceIndex > 0) {
                    newMarketplace = marketplaces[marketplaceIndex - 1];
                    newStyle = styles[styles.length - 1];
                    newSize = sizes[sizes.length - 1];
                }
                break;
            case 'ArrowRight':
                if (sizeIndex < sizes.length - 1) {
                    newSize = sizes[sizeIndex + 1];
                } else if (styleIndex < styles.length - 1) {
                    newStyle = styles[styleIndex + 1];
                    newSize = sizes[0];
                } else if (marketplaceIndex < marketplaces.length - 1) {
                    newMarketplace = marketplaces[marketplaceIndex + 1];
                    newStyle = styles[0];
                    newSize = sizes[0];
                }
                break;
            default:
                break;
        }

        return { marketplace: newMarketplace, style: newStyle, size: newSize };
    };

    const handleKeyNavigation = (e, currentMarketplace, currentStyle, currentSize) => {
        if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Tab'].includes(e.key)) {
            e.preventDefault();

            let direction = e.key;
            if (e.key === 'Tab') {
                direction = e.shiftKey ? 'ArrowLeft' : 'ArrowRight';
            }

            const nextCell = getNextCell(currentMarketplace, currentStyle, currentSize, direction);
            const nextCellId = `${nextCell.marketplace}-${nextCell.style}-${nextCell.size}`;

            
            setFocusedCell(nextCellId);
            requestAnimationFrame(() => {
                const nextInput = document.querySelector(`input[data-cell="${nextCellId}"]`);
                if (nextInput) {
                    nextInput.focus();
                    nextInput.select();
                }
            });
        }
    };

    if (loading && !isReadOnly) {
        return (
            <Card className="main-card mb-3">
                <CardBody className="text-center">
                    <i className="fa fa-spinner fa-spin fa-2x"></i>
                    <p className="mt-2">{t('common.loading')}...</p>
                </CardBody>
            </Card>
        );
    }

    
    const isCreationMode = isReadOnly && !partnerId;

    return (
        <Card className="main-card mb-3">
            <CardHeader>
                <CardTitle className="text-xl">{t('partners.article_grid_title')}</CardTitle>
            </CardHeader>
            <CardBody>
                {isCreationMode && (
                    <Alert color="info" className="mb-3">
                        <i className="fa fa-info-circle me-2"></i>
                        {t('partners.article_grid_creation_note')}
                    </Alert>
                )}
                {error && (
                    <Alert color="danger" className="mb-3">
                        {error}
                    </Alert>
                )}


                {marketplaces.map(marketplace => (
                    <div key={marketplace} className="mb-4">
                        <h5 className="mb-3">
                            {t('partners.article_grid_for_marketplace')}
                            <Badge color={marketplace === 'ozon' ? 'warning' : 'info'} className="ms-2">
                                {getMarketplaceDisplayName(marketplace)}
                            </Badge>
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
                                                        onKeyDown={(e) => handleKeyNavigation(e, marketplace, style, size)}
                                                        onFocus={() => setFocusedCell(`${marketplace}-${style}-${size}`)}
                                                        placeholder={t('partners.enter_sku')}
                                                        disabled={(!isCreationMode && isReadOnly) || saving}
                                                        bsSize="sm"
                                                        className={`text-center ${focusedCell === `${marketplace}-${style}-${size}`
                                                                ? 'border-primary bg-light'
                                                                : (articleGrid[marketplace]?.[style]?.[size] || '').trim() !== ''
                                                                    ? 'border-success bg-success text-white'
                                                                    : ''
                                                            }`}
                                                        data-cell={`${marketplace}-${style}-${size}`}
                                                        tabIndex={0}
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
