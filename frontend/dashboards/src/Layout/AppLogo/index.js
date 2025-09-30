import React, { Fragment } from "react";
import { connect } from "react-redux";

import AppMobileMenu from "../AppMobileMenu";

import {
  setEnableClosedSidebar,
  setEnableMobileMenu,
  setEnableMobileMenuSmall,
} from "../../reducers/ThemeOptions";

class HeaderLogo extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      active: false,
      mobile: false,
      activeSecondaryMenuMobile: false,
      panelTitle: 'Admin Panel'
    };
  }

  state = {
    openLeft: false,
    openRight: false,
    relativeWidth: false,
    width: 280,
    noTouchOpen: false,
    noTouchClose: false,
    panelTitle: 'Admin Panel',
  };

  getPanelTitleFromStorage = () => {
    try {
      const role = (typeof window !== 'undefined' && window.localStorage)
        ? (localStorage.getItem('userRole') || 'admin')
        : 'admin';
      return role === 'partner' ? 'Partner Panel' : 'Admin Panel';
    } catch (_) {
      return 'Admin Panel';
    }
  };

  updatePanelTitle = () => {
    this.setState({ panelTitle: this.getPanelTitleFromStorage() });
  };

  componentDidMount() {
    
    this.updatePanelTitle();
    
    window.addEventListener('auth-updated', this.updatePanelTitle);
    window.addEventListener('storage', this.updatePanelTitle);
  }

  componentWillUnmount() {
    window.removeEventListener('auth-updated', this.updatePanelTitle);
    window.removeEventListener('storage', this.updatePanelTitle);
  }

  render() {
    const panelTitle = this.state.panelTitle;
    return (
      <Fragment>
        <div className="app-header__logo">
          <div className="logo-src">{panelTitle}</div>
          {}
        </div>
        <AppMobileMenu />
      </Fragment>
    );
  }
}

const mapStateToProps = (state) => ({
  enableClosedSidebar: state.ThemeOptions.enableClosedSidebar,
  enableMobileMenu: state.ThemeOptions.enableMobileMenu,
  enableMobileMenuSmall: state.ThemeOptions.enableMobileMenuSmall,
});

const mapDispatchToProps = (dispatch) => ({
  setEnableClosedSidebar: (enable) => dispatch(setEnableClosedSidebar(enable)),
  setEnableMobileMenu: (enable) => dispatch(setEnableMobileMenu(enable)),
  setEnableMobileMenuSmall: (enable) =>
    dispatch(setEnableMobileMenuSmall(enable)),
});

export default connect(mapStateToProps, mapDispatchToProps)(HeaderLogo);
