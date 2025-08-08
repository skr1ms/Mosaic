import { Routes, Route, Navigate } from "react-router-dom";
import React, { Suspense, lazy, Fragment } from "react";
import Loader from "react-loaders";
import { ToastContainer } from "react-toastify";

import ProtectedRoute from "../../components/ProtectedRoute";

const AdminPages = lazy(() => import("../../DemoPages/AdminPages"));

const AppMain = () => {
    return (
        <Fragment>
            <Routes>
                {/* Dashboard - доступен всем авторизованным пользователям */}
                <Route path="/dashboard" element={
                    <ProtectedRoute>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка панели управления...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

                {/* Админские маршруты - только для админов */}
                <Route path="/partners/*" element={
                    <ProtectedRoute allowedRoles={['admin']}>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка управления партнерами...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

                <Route path="/coupons/*" element={
                    <ProtectedRoute allowedRoles={['admin']}>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка управления купонами...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

                {/* Аналитика - доступна всем (админам и партнерам) */}
                <Route path="/analytics" element={
                    <ProtectedRoute>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка аналитики...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

                <Route path="/analytics/*" element={
                    <ProtectedRoute>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка аналитики...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

                {/* Default redirect - перенаправляем на dashboard */}
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Routes>
            <ToastContainer/>
        </Fragment>
    )
};

export default AppMain;
