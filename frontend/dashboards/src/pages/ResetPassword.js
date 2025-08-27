import React, { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import api from '../api/api';

const ResetPassword = () => {
  const location = useLocation();
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fieldErrors, setFieldErrors] = useState({});
  const [formMessage, setFormMessage] = useState('');
  const [formMessageType, setFormMessageType] = useState('');

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const t = params.get('token') || '';
    setToken(t);
  }, [location.search]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormMessage('');
    setFormMessageType('');
    const errors = {};
    if (!token || token.length < 20) errors.token = 'Некорректный токен';
    if (!password || password.length < 6) errors.password = 'Минимум 6 символов';
    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }
    setLoading(true);
    try {
      await api.post('/auth/reset-password', { token, new_password: password });
      setFormMessageType('success');
      setFormMessage('Пароль изменён. Теперь вы можете войти.');
    } catch (e1) {
      setFormMessageType('error');
      setFormMessage(e1?.response?.data?.error || e1.message || 'Не удалось изменить пароль');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-container app-theme-white">
      <div className="d-flex justify-content-center align-items-center min-vh-100 bg-light">
        <div className="card shadow-lg" style={{ width: '420px', maxWidth: '92vw' }}>
          <div className="card-header text-center bg-primary text-white">
            <h4 className="mb-0">
              <i className="pe-7s-key mr-2"></i>
              Сброс пароля
            </h4>
          </div>
          <div className="card-body p-4">
            <form onSubmit={handleSubmit} noValidate>
              <div className="mb-3">
                <label className="form-label">Токен</label>
                <input className="form-control" value={token} onChange={(e)=> { setToken(e.target.value); if (fieldErrors.token) setFieldErrors({ ...fieldErrors, token: '' }); }} />
                {fieldErrors.token && <div className="form-text text-danger small mt-1">{fieldErrors.token}</div>}
              </div>
              <div className="mb-4">
                <label className="form-label">Новый пароль</label>
                <div className="input-group">
                  <input type={showPassword ? 'text' : 'password'} className="form-control" value={password} onChange={(e)=> { setPassword(e.target.value); if (fieldErrors.password) setFieldErrors({ ...fieldErrors, password: '' }); }} />
                  <button type="button" className="btn btn-outline-secondary" onClick={()=> setShowPassword(p=>!p)} title={showPassword ? 'Скрыть пароль' : 'Показать пароль'}>
                    <i className={showPassword ? 'pe-7s-look' : 'pe-7s-close-circle'}></i>
                  </button>
                </div>
                {fieldErrors.password && <div className="form-text text-danger small mt-1">{fieldErrors.password}</div>}
              </div>
              <button type="submit" className="btn btn-primary w-100" disabled={loading}>
                {loading ? 'Сохраняем...' : 'Сбросить пароль'}
              </button>

              {formMessage && (
                <div className={`small mt-2 ${formMessageType === 'error' ? 'text-danger' : 'text-success'}`}>
                  {formMessage}
                </div>
              )}
            </form>

            <div className="mt-3 text-center">
              <button type="button" className="btn btn-link" onClick={() => { window.location.hash = '#/'; }}>Вернуться к авторизации</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResetPassword;
