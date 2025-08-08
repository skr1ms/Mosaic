import React, { Fragment, useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

import {
  DropdownToggle,
  DropdownMenu,
  Button,
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

const UserBox = () => {
  const [userInfo, setUserInfo] = useState({
    role: 'admin',
    email: '',
    name: 'Пользователь'
  });
  const navigate = useNavigate();

  useEffect(() => {
    // Получаем информацию о пользователе из localStorage
    const role = localStorage.getItem('userRole') || 'admin';
    const email = localStorage.getItem('userEmail') || '';
    const name = localStorage.getItem('userName') || (role === 'admin' ? 'Администратор' : 'Партнер');
    
    setUserInfo({ role, email, name });
  }, []);

  const handleLogout = () => {
    // Очищаем localStorage
    localStorage.removeItem('token');
    localStorage.removeItem('userRole');
    localStorage.removeItem('userId');
    localStorage.removeItem('userEmail');
    localStorage.removeItem('userName');
    
    // Перенаправляем на главную страницу (которая покажет логин)
    navigate('/', { replace: true });
  };

  const getRoleDisplayName = (role) => {
    switch (role) {
      case 'admin':
        return 'Администратор';
      case 'partner':
        return 'Партнер';
      default:
        return 'Пользователь';
    }
  };

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
                <DropdownMenu className="rm-pointers dropdown-menu-lg" style={{ padding: 0, margin: 0, borderRadius: '10px', overflow: 'hidden' }}>
                  <div className="dropdown-menu-header" style={{ margin: 0, padding: 0 }}>
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
                                {userInfo.name}
                              </div>
                              <div className="widget-subheading opacity-8">
                                {getRoleDisplayName(userInfo.role)}
                              </div>
                              {userInfo.email && (
                                <div className="widget-subheading opacity-6">
                                  {userInfo.email}
                                </div>
                              )}
                            </div>
                            <div className="widget-content-right me-2">
                              <Button 
                                className="btn-pill btn-shadow btn-shine" 
                                color="focus"
                                onClick={handleLogout}
                              >
                                <i className="pe-7s-power mr-2"></i>
                                Выход
                              </Button>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </DropdownMenu>
              </UncontrolledButtonDropdown>
            </div>
          </div>
        </div>
      </div>
    </Fragment>
  );
};

export default UserBox;
