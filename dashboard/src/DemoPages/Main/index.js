import React, { Fragment, useState, useEffect } from "react";
import { connect } from "react-redux";
import cx from "classnames";
import { useResizeDetector } from "react-resize-detector";

import AppMain from "../../Layout/AppMain";
import AppHeader from "../../Layout/AppHeader";
import AppSidebar from "../../Layout/AppSidebar";
import Chat from "../../components/Chat";

const Main = (props) => {
  const [userRole, setUserRole] = useState('admin');
  const [userId, setUserId] = useState('');

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
    // Получаем информацию о пользователе из localStorage
    const role = localStorage.getItem('userRole') || 'admin';
    const id = localStorage.getItem('userId') || '';
    setUserRole(role);
    setUserId(id);
  }, []);

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
