import React, { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api/api';
import LanguageSelector from '../components/LanguageSelector';


function clearAuth() {
  localStorage.removeItem('token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('userRole');
  localStorage.removeItem('userId');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('userName');
  window.dispatchEvent(new Event('auth-updated'));
}

const Login = () => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({ login: '', password: '' });
  const [fieldErrors, setFieldErrors] = useState({ login: '', password: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  // Removed unused navigate
  // Если пришли с параметром force=1 — чистим локальные токены, чтобы всегда показать форму
  useEffect(() => {
    try {
      const params = new URLSearchParams(window.location.search);
      const force = params.get('force');
      if (force === '1') {
        clearAuth();
      }
    } catch (_) {}
  }, []);
  

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value
    });
    if (fieldErrors[name]) {
      setFieldErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    // Простая валидация полей
    const nextErrors = { login: '', password: '' };
    if (!formData.login.trim()) nextErrors.login = t('login.enter_login');
    if (!formData.password.trim()) nextErrors.password = t('login.enter_password');
    setFieldErrors(nextErrors);
    if (nextErrors.login || nextErrors.password) return;

    setLoading(true);

    try {
      const response = await api.post('/auth/login', {
        login: formData.login,
        password: formData.password
      });

      const { access_token, refresh_token, user } = response.data || {};

      
      if (access_token) {
        localStorage.setItem('token', `Bearer ${access_token}`);
      }
      if (refresh_token) {
        localStorage.setItem('refresh_token', refresh_token);
      }
      if (user?.role) {
        localStorage.setItem('userRole', user.role);
      }
      if (user?.id) {
        localStorage.setItem('userId', String(user.id));
      }
      if (user?.login) {
        localStorage.setItem('userEmail', user.login);
        localStorage.setItem('userName', user.login);
      }
      if (user?.email) {
        localStorage.setItem('userEmail', user.email);
      }

      const sessionId = Date.now() + '_' + Math.random().toString(36).substring(2);
      localStorage.setItem('auth_session_id', sessionId);
      sessionStorage.setItem('auth_session_id', sessionId);

      
      window.dispatchEvent(new Event('auth-updated'));

      
      setTimeout(() => {
        window.location.reload();
      }, 100);
    } catch (error) {
      let errorMessage = t('login.invalid_credentials');
      const server = error?.response?.data;
      if (server?.error) errorMessage = server.error;
      else if (server?.message) errorMessage = server.message;
      else if (error.message) errorMessage = error.message;
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  if (showForgotPassword) {
    return <ForgotPasswordForm onBack={() => setShowForgotPassword(false)} />;
  }

  return (
    <div className="app-container app-theme-white">
      {}
      <div style={{ position: 'absolute', top: '20px', right: '20px', zIndex: 1000 }}>
        <LanguageSelector />
      </div>
      
      <div className="d-flex justify-content-center align-items-center min-vh-100 bg-light">
        <div className="card shadow-lg" style={{ width: '500px', maxWidth: '90vw' }}>
          <div className="card-header text-center bg-primary text-white">
            <h4 className="mb-0">
              <i className="pe-7s-user mr-2"></i>
              {t('login.system_login_title')}
            </h4>
          </div>
          <div className="card-body p-4">
            {error && (
              <div className="alert alert-danger" role="alert">
                <i className="pe-7s-attention mr-2"></i>
                {error}
              </div>
            )}
            
            <form onSubmit={handleSubmit} noValidate>
              <div className="mb-3">
                <label htmlFor="login" className="form-label d-flex align-items-center" style={{gap:'8px'}}>
                  <i className="pe-7s-mail" style={{ position:'relative', left:'-2px' }}></i>
                  <span>{t('login.login_label')}</span>
                </label>
                <input
                  type="text"
                  className={`form-control ${fieldErrors.login ? 'is-invalid' : ''}`}
                  id="login"
                  name="login"
                  value={formData.login}
                  onChange={handleChange}
                  placeholder={t('login.login_placeholder')}
                />
                {fieldErrors.login && (
                  <div className="invalid-feedback" style={{ display: 'block' }}>{fieldErrors.login}</div>
                )}
              </div>

              <div className="mb-3">
                <label htmlFor="password" className="form-label d-flex align-items-center" style={{gap:'8px'}}>
                  <i className="pe-7s-lock" style={{ position:'relative', left:'-2px' }}></i>
                  <span>{t('login.password_label')}</span>
                </label>
                <div className="input-group">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    className={`form-control ${fieldErrors.password ? 'is-invalid' : ''}`}
                    id="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    placeholder={t('login.password_placeholder')}
                  />
                  <button
                    type="button"
                    className="btn btn-outline-secondary"
                    onClick={() => setShowPassword(prev => !prev)}
                    title={showPassword ? t('login.hide_password') : t('login.show_password')}
                  >
                    <i className={showPassword ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                  </button>
                </div>
                {fieldErrors.password && (
                  <div className="invalid-feedback" style={{ display: 'block' }}>{fieldErrors.password}</div>
                )}
              </div>

              <div className="mb-4 text-end">
                <button
                  type="button"
                  className="btn btn-link p-0"
                  onClick={() => setShowForgotPassword(true)}
                >
                  {t('login.forgot_password')}
                </button>
              </div>

              <button
                type="submit"
                className="btn btn-primary w-100 btn-lg"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                    {t('login.logging_in')}
                  </>
                ) : (
                  <>
                    <i className="pe-7s-right-arrow" style={{ marginRight: '6px', transform: 'translateX(-2px)' }}></i>
                    {t('login.login_button')}
                  </>
                )}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};


const ForgotPasswordForm = ({ onBack }) => {
  const { t } = useTranslation();
  const [formData, setFormData] = useState({ login: '', email: '' });
  const [fieldErrors, setFieldErrors] = useState({ login: '', email: '' });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState(''); // 'success' | 'error'
  const [siteKey, setSiteKey] = useState('');
  const [widgetId, setWidgetId] = useState(null);
  const recaptchaRef = useRef(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value
    });
    if (fieldErrors[name]) setFieldErrors(prev => ({ ...prev, [name]: '' }));
  };

  // Загружаем site key и скрипт v2
  useEffect(() => {
    let cancelled = false;
    
    const loadReCaptcha = async () => {
      try {
        let key = '';
        try {
          const resp = await api.get('/config/recaptcha');
          key = resp?.data?.site_key || '';
        } catch (_) {
          // В development используем переменную из Docker Compose
          key = process.env.REACT_APP_RECAPTCHA_SITE_KEY || '';
        }
        
        if (cancelled) return;
        setSiteKey(key);
        if (!key) {
          console.error('reCAPTCHA site key not found');
          return;
        }

        
        window.onRecaptchaLoad = () => {
          if (cancelled || !recaptchaRef.current || widgetId !== null) return;
          
          try {
            const id = window.grecaptcha.render(recaptchaRef.current, { 
              sitekey: key,
              callback: (response) => {
                console.log('reCAPTCHA completed:', response ? 'success' : 'failed');
              },
              'error-callback': () => {
                console.error('reCAPTCHA error occurred');
                setMessageType('error');
                setMessage(t('login.recaptcha_error_refresh'));
              },
              'expired-callback': () => {
                console.warn('reCAPTCHA expired');
                setMessageType('error');
                setMessage(t('login.recaptcha_expired'));
              }
            });
            setWidgetId(id);
            console.log('reCAPTCHA widget rendered with ID:', id);
          } catch (error) {
            console.error('Error rendering reCAPTCHA:', error);
          }
        };

        
        if (!document.querySelector('script[src*="google.com/recaptcha/api.js"]')) {
          const script = document.createElement('script');
          script.src = 'https://www.google.com/recaptcha/api.js?onload=onRecaptchaLoad&render=explicit';
          script.async = true;
          script.defer = true;
          document.head.appendChild(script);
          
          script.onerror = () => {
            console.error('Failed to load reCAPTCHA script');
            if (!cancelled) {
              setMessageType('error');
              setMessage(t('login.recaptcha_load_failed'));
            }
          };
        } else if (window.grecaptcha && window.grecaptcha.render) {
          
          window.onRecaptchaLoad();
        }
      } catch (error) {
        console.error('Error loading reCAPTCHA:', error);
        if (!cancelled) {
          setMessageType('error');
          setMessage(t('login.recaptcha_init_error'));
        }
      }
    };

    loadReCaptcha();
    return () => { cancelled = true; };
  }, [t, widgetId]);

  const handleSubmit = async (e) => {
    e.preventDefault();

    const nextErrors = { login: '', email: '' };
    if (!formData.login.trim()) nextErrors.login = t('login.enter_login');
    if (!formData.email.trim()) nextErrors.email = t('login.email_placeholder');
    setFieldErrors(nextErrors);
    if (nextErrors.login || nextErrors.email) {
      return;
    }

    if (!window.grecaptcha || widgetId === null) {
      setMessageType('error');
      setMessage(t('login.recaptcha_not_loaded'));
      return;
    }

    const token = window.grecaptcha.getResponse(widgetId);
    if (!token) {
      setMessageType('error');
      setMessage(t('login.recaptcha_required'));
      return;
    }

    setLoading(true);
    setMessage('');
    setMessageType('');

    try {
      await api.post('/auth/forgot-password', {
        login: formData.login,
        email: formData.email,
        captcha: token,
      });
      setMessageType('success');
      setMessage(t('login.email_sent_success'));
      
      try { window.grecaptcha.reset(widgetId); } catch (_) {}
    } catch (error) {
      setMessageType('error');
      let msg = error?.response?.data?.error || error.message || t('login.email_send_failed');
      const raw = (error?.response?.data?.error || error?.response?.data?.message || '').toLowerCase();
      if (raw.includes('not found') || raw.includes('не найден')) {
        msg = t('login.user_not_found');
      }
      setMessage(msg);
      
      try { 
        if (window.grecaptcha && widgetId !== null) {
          window.grecaptcha.reset(widgetId); 
        }
      } catch (_) {}
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-container app-theme-white">
      {}
      <div style={{ position: 'absolute', top: '20px', right: '20px', zIndex: 1000 }}>
        <LanguageSelector />
      </div>
      
      <div className="d-flex justify-content-center align-items-center min-vh-100 bg-light">
        <div className="card shadow-lg" style={{ width: '550px', maxWidth: '90vw' }}>
          <div className="card-header text-center bg-primary text-white">
            <h4 className="mb-0">
              <i className="pe-7s-lock mr-2"></i>
              {t('login.password_recovery_title')}
            </h4>
          </div>
          <div className="card-body p-4">
            {message && (
              <div className={`alert ${messageType === 'error' ? 'alert-danger' : 'alert-success'}`} role="alert">
                <i className={`pe-7s-${messageType === 'error' ? 'attention' : 'check'} mr-2`}></i>
                {message}
              </div>
            )}
            
            <form onSubmit={handleSubmit}>
              <div className="mb-3">
                <label htmlFor="login" className="form-label d-flex align-items-center" style={{gap:'8px'}}>
                  <i className="pe-7s-user" style={{ position:'relative', left:'-2px' }}></i>
                  <span>{t('login.login_label')}</span>
                </label>
                <input
                  type="text"
                  className={`form-control ${fieldErrors.login ? 'is-invalid' : ''}`}
                  id="login"
                  name="login"
                  value={formData.login}
                  onChange={handleChange}
                  placeholder={t('login.login_placeholder')}
                />
                {fieldErrors.login && (
                  <div className="invalid-feedback" style={{ display: 'block' }}>{fieldErrors.login}</div>
                )}
              </div>

              <div className="mb-3">
                <label htmlFor="email" className="form-label d-flex align-items-center" style={{gap:'8px'}}>
                  <i className="pe-7s-mail" style={{ position:'relative', left:'-2px' }}></i>
                  <span>{t('login.email_label')}</span>
                </label>
                <input
                  type="email"
                  className={`form-control ${fieldErrors.email ? 'is-invalid' : ''}`}
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleChange}
                  placeholder={t('login.email_placeholder')}
                />
                {fieldErrors.email && (
                  <div className="invalid-feedback" style={{ display: 'block' }}>{fieldErrors.email}</div>
                )}
              </div>

              {}
              <div className="mb-3">
                {siteKey ? (
                  <div ref={recaptchaRef} className="g-recaptcha" data-sitekey={siteKey}></div>
                ) : (
                  <div className="text-muted small">{t('login.loading_recaptcha')}</div>
                )}
              </div>

              <button
                type="submit"
                className="btn btn-primary w-100 btn-lg mb-3"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                    {t('login.sending')}
                  </>
                ) : (
                  <>
                    <i className="pe-7s-mail" style={{ marginRight: '6px', transform: 'translateX(-2px)' }}></i>
                    {t('login.send_email_button')}
                  </>
                )}
              </button>

              <button
                type="button"
                className="btn btn-outline-secondary w-100"
                onClick={onBack}
              >
                <i className="pe-7s-back" style={{ marginRight: '6px', transform: 'translateX(-2px)' }}></i>
                {t('login.back_to_login')}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
