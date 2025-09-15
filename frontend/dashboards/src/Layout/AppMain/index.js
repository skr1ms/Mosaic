import { Routes, Route } from "react-router-dom";
import React, { Suspense, lazy, Fragment } from "react";
import { useTranslation } from "react-i18next";
import Loader from "react-loaders";
import { ToastContainer } from "react-toastify";

import ProtectedRoute from "../../components/ProtectedRoute";

const AdminPages = lazy(() => import("../../AdminDashboardPages/AdminPages"));

const AppMain = () => {
    const { t } = useTranslation();
    return (
        <Fragment>
            <Routes>
                {}
                <Route path="/*" element={
                    <ProtectedRoute>
                        <Suspense fallback={
                            <div className="loader-container">
                                <div className="loader-container-inner">
                                    <div className="text-center">
                                        <Loader type="ball-pulse-rise"/>
                                    </div>
                                    <h6 className="mt-5">
                                        {t('chat.loading_application')}
                                    </h6>
                                </div>
                            </div>
                        }>
                            <AdminPages />
                        </Suspense>
                    </ProtectedRoute>
                } />
            </Routes>
            <ToastContainer/>
        </Fragment>
    )
};

export default AppMain;
