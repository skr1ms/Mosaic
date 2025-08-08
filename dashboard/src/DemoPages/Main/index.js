import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import { useNavigate, useLocation, Routes, Route } from "react-router-dom";
import cx from "classnames";
import { useResizeDetector } from "react-resize-detector";

import AppMain from "../../Layout/AppMain";
import AppHeader from "../../Layout/AppHeader";
import AppSidebar from "../../Layout/AppSidebar";
import Chat from "../../components/Chat";
import Login from "../../pages/Login";

const Main = (props) => {
  const [userRole, setUserRole] = useState('admin');
  const [userId, setUserId] = useState('');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();

  const {
    colorScheme,
    enableFixedHeader,
    enableFixedSidebar,
    enableFixedFooter,
    enableClosedSidebar,
    enableMobileMenu,
    enablePageTabsAlt,
  } = props;

  const { width, ref } = useResizeDetector();

  useEffect(() => {
    // Проверяем аутентификацию при загрузке
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('userRole') || 'admin';
    const id = localStorage.getItem('userId') || '';

    if (!token || !token.startsWith('Bearer ')) {
      // Если нет токена, показываем страницу логина
      setIsAuthenticated(false);
      setIsLoading(false);
      return;
    }

    setUserRole(role);
    setUserId(id);
    setIsAuthenticated(true);
    setIsLoading(false);
  }, []);

  // Если загрузка, показываем спиннер
  if (isLoading) {
    return (
      <div className="app-container">
        <div className="d-flex justify-content-center align-items-center min-vh-100">
          <div className="text-center">
            <div className="spinner-border text-primary" role="status">
              <span className="visually-hidden">Загрузка...</span>
            </div>
            <h6 className="mt-3">Загрузка приложения...</h6>
          </div>
        </div>
      </div>
    );
  }

  // Если не аутентифицирован, показываем страницу логина
  if (!isAuthenticated) {
    return <Login />;
  }

  // Если аутентифицирован, показываем dashboard
  return (
    <Fragment>
      <div ref={ref}>
        <div
          className={cx(
            "app-container app-theme-" + colorScheme,
            { "fixed-header": enableFixedHeader },
            { "fixed-sidebar": enableFixedSidebar || width < 992 },
            { "fixed-footer": enableFixedFooter },
            { "closed-sidebar": enableClosedSidebar || width < 992 },
            {
              "closed-sidebar-mobile": width < 992,
            },
            { "sidebar-mobile-open": enableMobileMenu },
            { "body-tabs-shadow-btn": enablePageTabsAlt }
          )}>
          <AppHeader />
          <div className="app-main">
            <AppSidebar />
            <div className="app-main__outer">
              <div className="app-main__inner">
                <AppMain />
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Чат - теперь работает автономно */}
      <Chat />
    </Fragment>
  );
};

const mapStateToProp = (state) => ({
  colorScheme: state.ThemeOptions.colorScheme,
  enableFixedHeader: state.ThemeOptions.enableFixedHeader,
  enableMobileMenu: state.ThemeOptions.enableMobileMenu,
  enableFixedFooter: state.ThemeOptions.enableFixedFooter,
  enableFixedSidebar: state.ThemeOptions.enableFixedSidebar,
  enableClosedSidebar: state.ThemeOptions.enableClosedSidebar,
  enablePageTabsAlt: state.ThemeOptions.enablePageTabsAlt,
});

export default connect(mapStateToProp)(Main);
