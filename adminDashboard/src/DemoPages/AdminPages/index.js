import React, { Fragment, Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';

const Dashboard = lazy(() => import('./Dashboard/Dashboard'));

const PartnersList = lazy(() => import('./Partners/PartnersList'));
const CreatePartner = lazy(() => import('./Partners/CreatePartner'));
const EditPartner = lazy(() => import('./Partners/EditPartner'));
const BlockedPartners = lazy(() => import('./Partners/BlockedPartners'));

const CreateCoupons = lazy(() => import('./Coupons/CreateCoupons'));
const ManageCoupons = lazy(() => import('./Coupons/ManageCoupons'));
const ActivatedCoupons = lazy(() => import('./Coupons/ActivatedCoupons'));

const Analytics = lazy(() => import('./Analytics/Analytics'));
const PartnerAnalytics = lazy(() => import('./Analytics/PartnerAnalytics'));

const AdminPages = () => {
    return (
        <Fragment>
            <Suspense fallback={
                <div className="loader-container">
                    <div className="loader-container-inner">
                        <h6 className="mt-3">
                            Загрузка...
                        </h6>
                    </div>
                </div>
            }>
                <Routes>
                    <Route path="/dashboard" element={<Dashboard />} />

                    <Route path="/partners/list" element={<PartnersList />} />
                    <Route path="/partners/create" element={<CreatePartner />} />
                    <Route path="/partners/edit/:id" element={<EditPartner />} />
                    <Route path="/partners/blocked" element={<BlockedPartners />} />

                    <Route path="/coupons/create" element={<CreateCoupons />} />
                    <Route path="/coupons/manage" element={<ManageCoupons />} />
                    <Route path="/coupons/activated" element={<ActivatedCoupons />} />
                    
                    <Route path="/analytics" element={<Analytics />} />
                    <Route path="/analytics/partners" element={<PartnerAnalytics />} />
                    
                    <Route path="/" element={<Navigate to="/dashboard" replace />} />
                </Routes>
            </Suspense>
        </Fragment>
    );
};

export default AdminPages; 