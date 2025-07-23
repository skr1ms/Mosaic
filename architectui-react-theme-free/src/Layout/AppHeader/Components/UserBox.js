import React, { Fragment } from "react";

import PerfectScrollbar from "react-perfect-scrollbar";

import {
  DropdownToggle,
  DropdownMenu,
  Nav,
  Col,
  Row,
  Button,
  NavItem,
  NavLink,
  UncontrolledButtonDropdown,
} from "reactstrap";

import city3 from "../../../assets/utils/images/dropdown-header/city3.jpg";

// CSS для силуэта аватара
const avatarStyle = {
  width: '42px',
  height: '42px',
  borderRadius: '50%',
  backgroundColor: '#6c757d',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  color: '#fff',
  fontSize: '18px'
};

class UserBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      active: false,
    };
  }

  render() {
    return (
      <Fragment>
        <div className="header-btn-lg pe-0">
          <div className="widget-content p-0">
            <div className="widget-content-wrapper">
              <div className="widget-content-left">
                <UncontrolledButtonDropdown>
                  <DropdownToggle color="link" className="p-0">
                    <div style={avatarStyle}>
                      <i className="pe-7s-user"></i>
                    </div>
                  </DropdownToggle>
                  <DropdownMenu className="rm-pointers dropdown-menu-lg">
                    <div className="dropdown-menu-header">
                      <div className="dropdown-menu-header-inner bg-info">
                        <div className="menu-header-image opacity-2"
                          style={{
                            backgroundImage: "url(" + city3 + ")",
                          }}/>
                        <div className="menu-header-content text-start">
                          <div className="widget-content p-0">
                            <div className="widget-content-wrapper">
                              <div className="widget-content-left me-3">
                                <div style={avatarStyle}>
                                  <i className="pe-7s-user"></i>
                                </div>
                              </div>
                              <div className="widget-content-left">
                                <div className="widget-heading">
                                  Администратор
                                </div>
                                <div className="widget-subheading opacity-8">
                                  Алмазная мозаика
                                </div>
                              </div>
                              <div className="widget-content-right me-2">
                                <Button className="btn-pill btn-shadow btn-shine" color="focus">
                                  Выход
                                </Button>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="scroll-area-xs"
                      style={{
                        height: "150px",
                      }}>
                      <PerfectScrollbar>
                        <Nav vertical>
                          <NavItem className="nav-item-header">
                            Активность
                          </NavItem>
                          <NavItem>
                            <NavLink href="#">
                              Уведомления
                              <div className="ms-auto badge rounded-pill bg-info">
                                3
                              </div>
                            </NavLink>
                          </NavItem>
                          <NavItem>
                            <NavLink href="#">Смена пароля</NavLink>
                          </NavItem>
                          <NavItem className="nav-item-header">
                            Мой аккаунт
                          </NavItem>
                          <NavItem>
                            <NavLink href="#">
                              Настройки
                              <div className="ms-auto badge bg-success">
                                Новое
                              </div>
                            </NavLink>
                          </NavItem>
                          <NavItem>
                            <NavLink href="#">Логи</NavLink>
                          </NavItem>
                        </Nav>
                      </PerfectScrollbar>
                    </div>
                    <Nav vertical>
                      <NavItem className="nav-item-divider mb-0" />
                    </Nav>
                    <div className="grid-menu grid-menu-2col">
                      <Row className="g-0">
                        <Col sm="6">
                          <Button className="btn-icon-vertical btn-transition btn-transition-alt pt-2 pb-2"
                            outline color="warning">
                            <i className="pe-7s-users icon-gradient bg-amy-crisp btn-icon-wrapper mb-2"> {" "} </i>
                            Партнеры
                          </Button>
                        </Col>
                        <Col sm="6">
                          <Button className="btn-icon-vertical btn-transition btn-transition-alt pt-2 pb-2"
                            outline color="danger">
                            <i className="pe-7s-ticket icon-gradient bg-love-kiss btn-icon-wrapper mb-2"> {" "} </i>
                            <b>Купоны</b>
                          </Button>
                        </Col>
                      </Row>
                    </div>
                    <Nav vertical>
                      <NavItem className="nav-item-divider" />
                      <NavItem className="nav-item-btn text-center">
                        <Button size="sm" className="btn-wide" color="primary">
                          Статистика
                        </Button>
                      </NavItem>
                    </Nav>
                  </DropdownMenu>
                </UncontrolledButtonDropdown>
              </div>
            </div>
          </div>
        </div>
      </Fragment>
    );
  }
}

export default UserBox;
