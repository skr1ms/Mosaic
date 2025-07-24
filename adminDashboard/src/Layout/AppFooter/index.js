import React, { Fragment } from "react";
import { connect } from "react-redux";

const AppFooter = ({ enableFixedFooter }) => {
  return (
    <Fragment>
      <div className="app-wrapper-footer">
        <div className="app-footer">
          <div className="app-footer__inner">
            <div className="app-footer-left">
              {/* Убрали ссылки футера */}
            </div>
            <div className="app-footer-right">
              {/* Убрали ссылки футера */}
            </div>
          </div>
        </div>
      </div>
    </Fragment>
  );
};

const mapStateToProps = (state) => ({
  enableFixedFooter: state.ThemeOptions.enableFixedFooter,
});

export default connect(mapStateToProps)(AppFooter);
