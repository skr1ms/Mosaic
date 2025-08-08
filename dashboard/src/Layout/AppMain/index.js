import { Routes, Route, Navigate } from "react-router-dom";
import React, { Suspense, lazy, Fragment } from "react";
import Loader from "react-loaders";
import { ToastContainer } from "react-toastify";

import ProtectedRoute from "../../components/ProtectedRoute";

const UserPages = lazy(() => import("../../DemoPages/UserPages"));
const Applications = lazy(() => import("../../DemoPages/Applications"));
const Dashboards = lazy(() => import("../../DemoPages/Dashboards"));
const AdminPages = lazy(() => import("../../DemoPages/AdminPages"));
const Login = lazy(() => import("../../pages/Login"));

const Widgets = lazy(() => import("../../DemoPages/Widgets"));
const Elements = lazy(() => import("../../DemoPages/Elements"));
const Components = lazy(() => import("../../DemoPages/Components"));
const Charts = lazy(() => import("../../DemoPages/Charts"));
const Forms = lazy(() => import("../../DemoPages/Forms"));
const Tables = lazy(() => import("../../DemoPages/Tables"));

const AppMain = () => {
    return (
        <Fragment>
            <Routes>
                {/* Login */}
                <Route path="/login" element={
                    <Suspense fallback={
                        <div className="loader-container">
                            <div className="loader-container-inner">
                                <div className="text-center">
                                    <Loader type="ball-pulse-rise"/>
                                </div>
                                <h6 className="mt-5">
                                    Загрузка страницы входа...
                                </h6>
                            </div>
                        </div>
                    }>
                        <Login />
                    </Suspense>
                } />

                {/* Protected Admin Pages */}
                <Route path="/dashboard" element={
                    <ProtectedRoute>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        Загрузка админ панели...
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />

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

                {/* Default redirect */}
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Routes>
            <ToastContainer/>
        </Fragment>
    )
};

export default AppMain;
