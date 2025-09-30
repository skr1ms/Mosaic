import React, { Fragment, useState, useEffect } from "react";
import { useTranslation } from "react-i18next";

import {
  DropdownToggle,
  DropdownMenu,
  DropdownItem,
  Button,
  UncontrolledButtonDropdown,
  Modal,
  ModalHeader,
  ModalBody,
  ModalFooter,
} from "reactstrap";

import api from "../../../api/api";
import LanguageSelector from "../../../components/LanguageSelector";


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
  const { t, i18n } = useTranslation();
  const [userInfo, setUserInfo] = useState({
    role: 'admin',
    email: '',
    name: t('common.name')
  });
  const [showChangePassword, setShowChangePassword] = useState(false);
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  });
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');
  const [showPasswordFields, setShowPasswordFields] = useState({
    currentPassword: false,
    newPassword: false,
    confirmPassword: false
  });
  
  // Состояние для изменения email
  const [showChangeEmail, setShowChangeEmail] = useState(false);
  const [emailForm, setEmailForm] = useState({
    password: '',
    newEmail: ''
  });
  const [emailLoading, setEmailLoading] = useState(false);
  const [emailError, setEmailError] = useState('');
  const [emailSuccess, setEmailSuccess] = useState('');
  const [showEmailPassword, setShowEmailPassword] = useState(false);

  // {t('userbox.create_administrator')} (только для роли admin)
  const [showCreateAdmin, setShowCreateAdmin] = useState(false);
  const [adminForm, setAdminForm] = useState({ login: '', email: '', password: '' });
  const [adminErrors, setAdminErrors] = useState({ login: '', email: '', password: '' });
  const [adminLoading, setAdminLoading] = useState(false);
  const [adminSuccess, setAdminSuccess] = useState('');
  const [adminError, setAdminError] = useState('');
  const [showAdminPwd, setShowAdminPwd] = useState(false);
  

  const getRoleDisplayName = React.useCallback((role) => {
    switch (role) {
      case 'admin':
      case 'main_admin':
        return t('user.profile');
      case 'partner':
        return t('navigation.partners');
      default:
        return t('common.name');
    }
  }, [t]);

  useEffect(() => {
    
    const role = localStorage.getItem('userRole') || 'admin';
    const email = localStorage.getItem('userEmail') || '';
    const name = localStorage.getItem('userName') || (role === 'admin' || role === 'main_admin' ? t('user.profile') : t('navigation.partners'));
    
    setUserInfo({ role, email, name });
  }, [t, i18n.language, getRoleDisplayName]);

  const handleLogout = () => {
    
    localStorage.removeItem('token');
    localStorage.removeItem('userRole');
    localStorage.removeItem('userId');
    localStorage.removeItem('userEmail');
    localStorage.removeItem('userName');
    
    
    window.location.hash = '#/';
  };

  const togglePasswordVisibility = (field) => {
    setShowPasswordFields(prev => ({
      ...prev,
      [field]: !prev[field]
    }));
  };

  

  const handlePasswordFormChange = (e) => {
    setPasswordForm({
      ...passwordForm,
      [e.target.name]: e.target.value
    });
  };

  const handleChangePassword = async (e) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    // Валидация
    if (!passwordForm.currentPassword) {
      setPasswordError(t('user.enter_password'));
      return;
    }
    if (!passwordForm.newPassword || passwordForm.newPassword.length < 6) {
      setPasswordError(t('user.password_requirements'));
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      setPasswordError(t('user.passwords_not_match'));
      return;
    }

    setPasswordLoading(true);

    try {
      await api.post('/auth/change-password', {
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword
      });

      setPasswordSuccess(t('user.password_changed'));
      setPasswordForm({
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
      });

      // Закрываем модал через некоторое время
      setTimeout(() => {
        setShowChangePassword(false);
        setPasswordSuccess('');
      }, 2000);

    } catch (error) {
      setPasswordError(error?.response?.data?.error || t('notifications.operation_failed'));
    } finally {
      setPasswordLoading(false);
    }
  };

  
  const handleEmailFormChange = (e) => {
    const { name, value } = e.target;
    setEmailForm(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleChangeEmail = async (e) => {
    e.preventDefault();
    setEmailLoading(true);
    setEmailError('');
    setEmailSuccess('');

    // Валидация
    if (!emailForm.password) {
      setEmailError(t('user.enter_password'));
      setEmailLoading(false);
      return;
    }

    if (!emailForm.newEmail) {
      setEmailError(t('user.enter_email'));
      setEmailLoading(false);
      return;
    }

    
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(emailForm.newEmail)) {
      setEmailError(t('user.invalid_email'));
      setEmailLoading(false);
      return;
    }

    try {
      await api.post('/auth/change-email', {
        password: emailForm.password,
        new_email: emailForm.newEmail
      });

      setEmailSuccess(t('user.email_changed'));
      
      
      setUserInfo(prev => ({
        ...prev,
        email: emailForm.newEmail
      }));

      
      localStorage.setItem('userEmail', emailForm.newEmail);

      
      setEmailForm({
        password: '',
        newEmail: ''
      });

      // Закрываем модал через некоторое время
      setTimeout(() => {
        setShowChangeEmail(false);
        setEmailSuccess('');
      }, 2000);

    } catch (error) {
      const status = error?.response?.status;
      const apiError = error?.response?.data?.error || error?.response?.data?.detail;
      if (status === 429) {
        setEmailError(t('user.too_many_requests'));
      } else if (status === 401) {
        setEmailError(t('user.wrong_password'));
      } else if (status === 409 || (apiError && /in use|already/i.test(apiError))) {
        setEmailError(t('user.email_in_use'));
      } else {
        setEmailError(apiError || t('notifications.operation_failed'));
      }
    } finally {
      setEmailLoading(false);
    }
  };

  return (
    <Fragment>
      <div className="header-btn-lg pe-0">
        <div className="widget-content p-0">
          <div className="widget-content-wrapper">
            <div className="widget-content-left d-flex align-items-center">
              <LanguageSelector />
              <UncontrolledButtonDropdown>
                <DropdownToggle color="link" className="p-0">
                  <div style={avatarStyle}>
                    <i className="pe-7s-user"></i>
                  </div>
                </DropdownToggle>
                <DropdownMenu className="rm-pointers dropdown-menu-lg" style={{ padding: 0, margin: 0, borderRadius: '10px', overflow: 'hidden' }}>
                  <div className="dropdown-menu-header" style={{ margin: 0, padding: 0 }}>
                    <div className="dropdown-menu-header-inner" style={{ backgroundColor: '#ffffff', color: '#000' }}>
                      <div className="menu-header-content text-start">
                        <div className="widget-content p-0">
                          <div className="widget-content-wrapper">
                            <div className="widget-content-left me-3">
                              <div style={avatarStyle}>
                                <i className="pe-7s-user"></i>
                              </div>
                            </div>
                            <div className="widget-content-left">
                               <div className="widget-heading" style={{ color: '#000' }}>
                                {userInfo.name}
                              </div>
                               <div className="widget-subheading opacity-8" style={{ color: '#000' }}>
                                {getRoleDisplayName(userInfo.role)}
                              </div>
                              {userInfo.email && (
                                <div className="widget-subheading opacity-6" style={{ color: '#000' }}>
                                  {userInfo.email}
                                </div>
                              )}
                            </div>
                            {}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div className="scroll-area-xs" style={{ height: 'auto' }}>
                    <div className="scrollbar-container">
                      {(userInfo.role === 'admin' || userInfo.role === 'main_admin' || userInfo.role === 'partner') && (
                        <DropdownItem onClick={() => setShowChangePassword(true)}>
                          <i className="pe-7s-key me-2" style={{ marginLeft: 4 }}></i>
                          <span style={{ marginLeft: 4 }}>{t('user.change_password')}</span>
                        </DropdownItem>
                      )}
                      {(userInfo.role === 'admin' || userInfo.role === 'main_admin' || userInfo.role === 'partner') && (
                        <DropdownItem onClick={() => setShowChangeEmail(true)}>
                          <i className="pe-7s-mail me-2"></i>
                          {t('user.change_email')}
                        </DropdownItem>
                      )}
                      {userInfo.role === 'main_admin' && (
                        <DropdownItem onClick={() => setShowCreateAdmin(true)}>
                          <i className="pe-7s-add-user me-2"></i>
                          {t('user.create_admin')}
                        </DropdownItem>
                      )}
                      <DropdownItem divider />
                      <DropdownItem onClick={handleLogout}>
                        <i className="pe-7s-power me-2"></i>
                        {t('navigation.logout')}
                      </DropdownItem>
                    </div>
                  </div>
                </DropdownMenu>
              </UncontrolledButtonDropdown>
            </div>
          </div>
        </div>
      </div>

      {}
      <Modal isOpen={showChangePassword} toggle={() => setShowChangePassword(false)}>
        <ModalHeader toggle={() => setShowChangePassword(false)}>
          <i className="pe-7s-key me-2"></i>
          {t('user.change_password')}
        </ModalHeader>
        <ModalBody>
          {passwordError && (
            <div className="alert alert-danger">
              <i className="pe-7s-attention me-2"></i>
              {passwordError}
            </div>
          )}
          {passwordSuccess && (
            <div className="alert alert-success">
              <i className="pe-7s-check me-2"></i>
              {passwordSuccess}
            </div>
          )}
          <form onSubmit={handleChangePassword}>
              <div className="mb-3">
              <label className="form-label">{t('user.current_password')}</label>
              <div className="input-group">
                <input
                  type={showPasswordFields.currentPassword ? "text" : "password"}
                  className="form-control"
                  name="currentPassword"
                  value={passwordForm.currentPassword}
                  onChange={handlePasswordFormChange}
                  required
                />
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={() => togglePasswordVisibility('currentPassword')}
                  title={showPasswordFields.currentPassword ? t('user.hide_password') : t('user.show_password')}
                >
                  <i className={showPasswordFields.currentPassword ? "pe-7s-look" : "pe-7s-close-circle"}></i>
                </button>
              </div>
            </div>
              <div className="mb-3">
              <label className="form-label">{t('user.new_password')}</label>
              <div className="input-group">
                <input
                  type={showPasswordFields.newPassword ? "text" : "password"}
                  className="form-control"
                  name="newPassword"
                  value={passwordForm.newPassword}
                  onChange={handlePasswordFormChange}
                  required
                />
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={() => togglePasswordVisibility('newPassword')}
                  title={showPasswordFields.newPassword ? t('user.hide_password') : t('user.show_password')}
                >
                  <i className={showPasswordFields.newPassword ? "pe-7s-look" : "pe-7s-close-circle"}></i>
                </button>
              </div>
            </div>
              <div className="mb-3">
              <label className="form-label">{t('user.confirm_password')}</label>
              <div className="input-group">
                <input
                  type={showPasswordFields.confirmPassword ? "text" : "password"}
                  className="form-control"
                  name="confirmPassword"
                  value={passwordForm.confirmPassword}
                  onChange={handlePasswordFormChange}
                  required
                />
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={() => togglePasswordVisibility('confirmPassword')}
                  title={showPasswordFields.confirmPassword ? t('user.hide_password') : t('user.show_password')}
                >
                  <i className={showPasswordFields.confirmPassword ? "pe-7s-look" : "pe-7s-close-circle"}></i>
                </button>
              </div>
            </div>
          </form>
        </ModalBody>
        <ModalFooter>
          <Button 
            color="primary" 
            onClick={handleChangePassword}
            disabled={passwordLoading}
          >
            {passwordLoading ? (
              <>
                <span className="spinner-border spinner-border-sm me-2"></span>
                {t('user.saving')}
              </>
            ) : (
              <>
                <i className="pe-7s-check me-2"></i>
                {t('user.change_password')}
              </>
            )}
          </Button>
          <Button color="secondary" onClick={() => setShowChangePassword(false)}>
            {t('common.cancel')}
          </Button>
        </ModalFooter>
      </Modal>

              {}
      <Modal isOpen={showCreateAdmin} toggle={() => setShowCreateAdmin(false)}>
        <ModalHeader toggle={() => setShowCreateAdmin(false)}>
          <i className="pe-7s-add-user me-2"></i>
          {t('user.create_admin')}
        </ModalHeader>
        <ModalBody>
          {adminError && (
            <div className="alert alert-danger" role="alert">{adminError}</div>
          )}
          {adminSuccess && (
            <div className="alert alert-success" role="alert">{adminSuccess}</div>
          )}
          <form onSubmit={(e)=>{ e.preventDefault(); }} noValidate>
            <div className="mb-3">
              <label className="form-label">{t('common.name')}</label>
              <input
                type="text"
                className={`form-control ${adminErrors.login ? 'is-invalid' : ''}`}
                value={adminForm.login}
                onChange={(e)=>{ setAdminForm({...adminForm, login: e.target.value}); if (adminErrors.login) setAdminErrors({...adminErrors, login: ''}); }}
                placeholder={t('user.enter_login')}
                disabled={adminLoading}
              />
              {adminErrors.login && <div className="invalid-feedback" style={{display:'block'}}>{adminErrors.login}</div>}
            </div>
            <div className="mb-3">
              <label className="form-label">{t('common.email')}</label>
              <input
                type="email"
                className={`form-control ${adminErrors.email ? 'is-invalid' : ''}`}
                value={adminForm.email}
                onChange={(e)=>{ setAdminForm({...adminForm, email: e.target.value}); if (adminErrors.email) setAdminErrors({...adminErrors, email: ''}); }}
                placeholder="admin@example.com"
                disabled={adminLoading}
              />
              {adminErrors.email && <div className="invalid-feedback" style={{display:'block'}}>{adminErrors.email}</div>}
            </div>
            <div className="mb-3">
              <label className="form-label">{t('user.current_password')}</label>
              <div className="input-group">
                <input
                  type={showAdminPwd ? 'text' : 'password'}
                  className={`form-control ${adminErrors.password ? 'is-invalid' : ''}`}
                  value={adminForm.password}
                  onChange={(e)=>{ setAdminForm({...adminForm, password: e.target.value}); if (adminErrors.password) setAdminErrors({...adminErrors, password: ''}); }}
                  placeholder={t('user.minimum_8_chars')}
                  disabled={adminLoading}
                />
                <button type="button" className="btn btn-outline-secondary" onClick={()=> setShowAdminPwd(v=>!v)} disabled={adminLoading} title={showAdminPwd ? t('user.hide_password') : t('user.show_password')}>
                  <i className={showAdminPwd ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                </button>
              </div>
              {adminErrors.password && <div className="invalid-feedback" style={{display:'block'}}>{adminErrors.password}</div>}
            </div>
        </form>
        </ModalBody>
        <ModalFooter>
          <Button
            color="primary"
            disabled={adminLoading}
            onClick={async ()=>{
              setAdminError(''); setAdminSuccess('');
              // Простая валидация
              const errs = { login: '', email: '', password: '' };
              if (!adminForm.login.trim()) errs.login = t('user.enter_login');
              if (!adminForm.email.trim()) errs.email = t('user.enter_email');
              const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
              if (adminForm.email && !emailRe.test(adminForm.email)) errs.email = t('user.invalid_email');
              if (!adminForm.password || adminForm.password.length < 8) errs.password = t('user.minimum_8_chars');
              setAdminErrors(errs);
              if (errs.login || errs.email || errs.password) return;
              try {
                setAdminLoading(true);
                await api.post('/admin/main-admin/admins', { login: adminForm.login, email: adminForm.email, password: adminForm.password });
                setAdminSuccess(t('user.admin_created'));
                setAdminForm({ login: '', email: '', password: '' });
                setTimeout(()=>{ setShowCreateAdmin(false); setAdminSuccess(''); }, 1500);
              } catch (e) {
                const s = e?.response?.data?.error || e?.response?.data?.message || e.message;
                setAdminError(s || t('notifications.operation_failed'));
              } finally {
                setAdminLoading(false);
              }
            }}
          >
            {adminLoading ? (<><span className="spinner-border spinner-border-sm me-2"></span>{t('user.creation')}</>) : (<><i className="pe-7s-check me-2"></i>{t('common.create')}</>)}
          </Button>
          <Button color="secondary" onClick={()=> setShowCreateAdmin(false)} disabled={adminLoading}>{t('common.cancel')}</Button>
        </ModalFooter>
      </Modal>

      {}
      <Modal isOpen={showChangeEmail} toggle={() => setShowChangeEmail(false)}>
        <ModalHeader toggle={() => setShowChangeEmail(false)}>
          {t('user.change_email')}
        </ModalHeader>
        <ModalBody>
          {emailError && (
            <div className="alert alert-danger" role="alert">
              {emailError}
            </div>
          )}
          {emailSuccess && (
            <div className="alert alert-success" role="alert">
              {emailSuccess}
            </div>
          )}
          <form onSubmit={handleChangeEmail}>
            <div className="mb-3">
              <label className="form-label">{t('common.email')}</label>
              <input
                type="text"
                className="form-control"
                value={userInfo.email || ''}
                readOnly
              />
            </div>
            <div className="mb-3">
              <label className="form-label">{t('user.new_email')}</label>
              <input
                type="text"
                className="form-control"
                name="newEmail"
                value={emailForm.newEmail}
                onChange={handleEmailFormChange}
                placeholder={t('user.enter_email')}
                disabled={emailLoading}
              />
            </div>
            <div className="mb-3">
              <label className="form-label">{t('user.current_password')}</label>
              <div className="input-group">
                <input
                  type={showEmailPassword ? 'text' : 'password'}
                  className="form-control"
                  name="password"
                  value={emailForm.password}
                  onChange={handleEmailFormChange}
                  placeholder={t('user.enter_password')}
                  disabled={emailLoading}
                />
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={() => setShowEmailPassword((v) => !v)}
                  title={showEmailPassword ? t('user.hide_password') : t('user.show_password')}
                  disabled={emailLoading}
                >
                  <i className={showEmailPassword ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                </button>
              </div>
            </div>
          </form>
        </ModalBody>
        <ModalFooter>
          <Button 
            color="primary" 
            onClick={handleChangeEmail}
            disabled={emailLoading}
          >
            {emailLoading ? (
              <>
                <span className="spinner-border spinner-border-sm me-2"></span>
                {t('user.saving')}
              </>
            ) : (
              <>
                <i className="pe-7s-mail me-2"></i>
                {t('user.change_email')}
              </>
            )}
          </Button>
          <Button color="secondary" onClick={() => setShowChangeEmail(false)}>
            {t('common.cancel')}
          </Button>
        </ModalFooter>
      </Modal>
    </Fragment>
  );
};

export default UserBox;
