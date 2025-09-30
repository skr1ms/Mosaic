import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import { useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import cx from "classnames";
import { useResizeDetector } from "react-resize-detector";

import AppMain from "../../Layout/AppMain";
import AppHeader from "../../Layout/AppHeader";
import AppSidebar from "../../Layout/AppSidebar";
import Chat from "../../components/Chat";
import Login from "../../pages/Login";
import { checkTokenOnInit, clearAuth } from "../../api/api";

const Main = (props) => {
  const { t } = useTranslation();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
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
    const evaluateAuth = () => {
      setIsLoading(true);
      try {
        
        const isValid = checkTokenOnInit();
        setIsAuthenticated(isValid);
      } catch (error) {
        console.error('Auth evaluation error:', error);
        clearAuth();
        setIsAuthenticated(false);
      } finally {
        setIsLoading(false);
      }
    };

    
    evaluateAuth();

    
    window.addEventListener('auth-updated', evaluateAuth);
    return () => window.removeEventListener('auth-updated', evaluateAuth);
  }, [location]);

  
  if (isLoading) {
    return (
      <div className="app-container">
        <div className="d-flex justify-content-center align-items-center min-vh-100">
          <div className="text-center">
            <div className="spinner-border text-primary" role="status">
              <span className="visually-hidden">{t('chat.loading')}</span>
            </div>
            <h6 className="mt-3">{t('chat.loading_application')}</h6>
          </div>
        </div>
      </div>
    );
  }

  
  if (!isAuthenticated) {
    return <Login />;
  }

  
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
      
      {}
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
