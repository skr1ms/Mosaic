import React, { Fragment, Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

// Компонент для обработки ошибок загрузки
const ErrorBoundary = ({ children }) => {
    const [hasError, setHasError] = React.useState(false);
    
    React.useEffect(() => {
        const handleError = (error) => {
            console.error('Lazy loading error:', error);
            setHasError(true);
        };
        
        window.addEventListener('error', handleError);
        return () => window.removeEventListener('error', handleError);
    }, []);
    
    if (hasError) {
        return (
            <div className="alert alert-danger m-3">
                <h5>Ошибка загрузки страницы</h5>
                <p>Произошла ошибка при загрузке компонента. Попробуйте обновить страницу.</p>
                <button 
                    className="btn btn-primary" 
                    onClick={() => window.location.reload()}
                >
                    Обновить страницу
                </button>
            </div>
        );
    }
    
    return children;
};

const Dashboard = lazy(() => import('./Dashboard/Dashboard'));

const PartnersList = lazy(() => import('./Partners/PartnersList'));
const CreatePartner = lazy(() => import('./Partners/CreatePartner'));
const EditPartner = lazy(() => import('./Partners/EditPartner'));
const PartnerDetails = lazy(() => import('./Partners/PartnerDetails'));
const BlockedPartners = lazy(() => import('./Partners/BlockedPartners'));

const CreateCoupons = lazy(() => import('./Coupons/CreateCoupons'));
const ManageCoupons = lazy(() => import('./Coupons/ManageCoupons'));
const ActivatedCoupons = lazy(() => import('./Coupons/ActivatedCoupons'));

const Analytics = lazy(() => import('./Analytics/Analytics'));
const PartnerAnalytics = lazy(() => import('./Analytics/PartnerAnalytics'));
const SystemStatus = lazy(() => import('./System/SystemStatus'));
const AdminManagement = lazy(() => import('./System/AdminManagement'));
const ImageQueue = lazy(() => import('./Images/ImageQueue'));
const SupportChats = lazy(() => import('./Support/SupportChats'));

// Партнерские страницы больше не импортируются из отдельного каталога.
// Добавляем ленивые импорты существующих партнёрских страниц
const PartnerCoupons = lazy(() => import('../../PartnerDashboardPages/PartnerPages/Coupons/List'));
const PartnerCouponsExport = lazy(() => import('../../PartnerDashboardPages/PartnerPages/Coupons/Export'));
const PartnerProfile = lazy(() => import('../../PartnerDashboardPages/PartnerPages/Profile/Profile'));
const PartnerAnalyticsPage = lazy(() => import('../../PartnerDashboardPages/PartnerPages/Analytics/Analytics'));

const AdminPages = () => {
    const { t } = useTranslation();
    // Получаем роль пользователя
    const userRole = localStorage.getItem('userRole') || 'admin';

    return (
        <Fragment>
            <ErrorBoundary>
                <Suspense fallback={
                    <div className="loader-container">
                        <div className="loader-container-inner">
                            <h6 className="mt-3">
                                {t('chat.loading')}
                            </h6>
                        </div>
                    </div>
                }>
                    <Routes>
                        {/* Главная панель - доступна всем */}
                        <Route path="/dashboard" element={<Dashboard />} />

                        {/* Админские маршруты - только для админов и main_admin */}
                        {(userRole === 'admin' || userRole === 'main_admin') && (
                            <>
                                <Route path="/partners/list" element={<PartnersList />} />
                                <Route path="/partners/create" element={<CreatePartner />} />
                                <Route path="/partners/edit/:id" element={<EditPartner />} />
                                <Route path="/partners/view/:id" element={<PartnerDetails />} />
                                <Route path="/partners/blocked" element={<BlockedPartners />} />

                                <Route path="/coupons/create" element={<CreateCoupons />} />
                                <Route path="/coupons/manage" element={<ManageCoupons />} />
                                <Route path="/coupons/activated" element={<ActivatedCoupons />} />
                                <Route path="/images/queue" element={<ImageQueue />} />
                                <Route path="/support/chats" element={<SupportChats />} />
                            </>
                        )}

                        {/* Управление администраторами - только для main_admin */}
                        {userRole === 'main_admin' && (
                            <Route path="/system/admins" element={<AdminManagement />} />
                        )}

                        {/* Общая аналитика (админская) */}
                        {(userRole === 'admin' || userRole === 'main_admin') && (
                            <Route path="/analytics" element={<Analytics />} />
                        )}
                        
                        {/* Статистика по партнерам - только для админов и main_admin */}
                        {(userRole === 'admin' || userRole === 'main_admin') && (
                            <Route path="/analytics/partners" element={<PartnerAnalytics />} />
                        )}

                        {/* Партнерские маршруты */}
                        {userRole === 'partner' && (
                            <>
                                <Route path="/partner/coupons" element={<PartnerCoupons />} />
                                <Route path="/partner/coupons/export" element={<PartnerCouponsExport />} />
                                <Route path="/partner/profile" element={<PartnerProfile />} />
                                <Route path="/partner/analytics" element={<PartnerAnalyticsPage />} />
                            </>
                        )}
                        {(userRole === 'admin' || userRole === 'main_admin') && (
                            <Route path="/system/status" element={<SystemStatus />} />
                        )}
                        
                        {/* Default redirect */}
                        <Route path="/" element={<Navigate to="/dashboard" replace />} />
                        {/* Fallback внутри админки */}
                        <Route path="*" element={<Dashboard />} />
                    </Routes>
                </Suspense>
            </ErrorBoundary>
        </Fragment>
    );
};

export default AdminPages; 