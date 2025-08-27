import React, { Fragment, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { connect } from "react-redux";
import { useTranslation } from "react-i18next";
import { setEnableMobileMenu } from "../../reducers/ThemeOptions";
import {
  MainNav,
  AdminNav,
  PartnerNav,
  SystemNav,
} from "./NavItems";

const SubMenu = ({ item, toggleMobileSidebar }) => {
  const { t } = useTranslation();
  const [isSubMenuOpen, setIsSubMenuOpen] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();

  const toggleSubMenu = (e) => {
    if (item.content && item.content.length > 0) {
      e.preventDefault();
      e.stopPropagation();
      setIsSubMenuOpen(!isSubMenuOpen);
    } else if (item.to && !item.external) {
      toggleMobileSidebar();
    }
  };

  const hasSubmenu = item.content && item.content.length > 0;
  
  // Determine if the parent or any child is active
  const isActive = location.pathname === item.to || 
    (hasSubmenu && item.content.some(child => child.to === location.pathname));

  const LinkComponent = item.external ? 'a' : Link;
  const linkProps = item.external 
    ? { href: item.to, target: "_blank", rel: "noopener noreferrer" }
    : { to: item.to || "#" };

  return (
    <li className={`metismenu-item ${isActive ? "active" : ""}`}>
      <LinkComponent
        {...linkProps}
        className={`metismenu-link ${isActive ? "active" : ""}`}
        onClick={toggleSubMenu}
        style={{ display: 'flex', alignItems: 'center' }}
      >
        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <i className={`metismenu-icon ${item.icon}`} />
          {item.translate ? t(item.label) : item.label}
        </span>
        {hasSubmenu && (
          <i
            className={`metismenu-state-icon pe-7s-angle-${isSubMenuOpen ? 'up' : 'down'}`}
            style={{ marginLeft: 'auto' }}
          />
        )}
      </LinkComponent>
      {hasSubmenu && (
        <ul className={`metismenu-container ${isSubMenuOpen ? "visible" : ""}`}>
          {item.content.map((child, i) => (
            <li key={i} className="metismenu-item">
              <Link
                to={child.to}
                className={`metismenu-link ${
                  location.pathname === child.to ? "active" : ""
                }`}
                onClick={(e) => {
                  // не даём событию всплыть к родителю, который вызывает preventDefault
                  e.stopPropagation();
                  toggleMobileSidebar();
                }}
              >
                {child.translate ? t(child.label) : child.label}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </li>
  );
};

const Nav = ({ enableMobileMenu, setEnableMobileMenu }) => {
  const { t } = useTranslation();
  const toggleMobileSidebar = () => {
    if (enableMobileMenu) {
      setEnableMobileMenu(false);
    }
  };

  const renderMenu = (items) =>
    items.map((item, i) => (
      <SubMenu key={i} item={item} toggleMobileSidebar={toggleMobileSidebar} />
    ));

  // Получаем роль пользователя из localStorage
  const userRole = localStorage.getItem('userRole') || 'admin';

  return (
    <Fragment>
      <div className="vertical-nav-menu">
        {/* Главное меню - доступно всем */}
        <h5 className="app-sidebar__heading">{t('navigation.main').toUpperCase()}</h5>
        <ul className="metismenu-container">{renderMenu(MainNav)}</ul>

        {/* Админские разделы - только для админов и main_admin */}
        {(userRole === 'admin' || userRole === 'main_admin') && (
          <>
            <h5 className="app-sidebar__heading">{t('navigation.administration').toUpperCase()}</h5>
            <ul className="metismenu-container">{renderMenu(AdminNav)}</ul>
          </>
        )}

        {/* Партнерские разделы - только для партнеров */}
        {userRole === 'partner' && (
          <>
            <h5 className="app-sidebar__heading">{t('navigation.management').toUpperCase()}</h5>
            <ul className="metismenu-container">{renderMenu(PartnerNav)}</ul>
          </>
        )}

        {/* Системные операции - только для админов и main_admin */}
        {(userRole === 'admin' || userRole === 'main_admin') && (
          <>
            <h5 className="app-sidebar__heading">{t('navigation.system').toUpperCase()}</h5>
            <ul className="metismenu-container">
              {renderMenu(SystemNav.filter(item => {
                // Управление администраторами только для main_admin
                if (item.to === '/system/admins') {
                  return userRole === 'main_admin';
                }
                // S3 MinIO доступен для admin и main_admin
                if (item.label === 'system.s3_minio') {
                  return userRole === 'admin' || userRole === 'main_admin';
                }
                return true;
              }))}
            </ul>
          </>
        )}
      </div>
    </Fragment>
  );
};

const mapStateToProps = (state) => ({
  enableMobileMenu: state.ThemeOptions.enableMobileMenu,
});

const mapDispatchToProps = (dispatch) => ({
  setEnableMobileMenu: (enable) => dispatch(setEnableMobileMenu(enable)),
});

export default connect(mapStateToProps, mapDispatchToProps)(Nav);
