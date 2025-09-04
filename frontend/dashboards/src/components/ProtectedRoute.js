import React from 'react';
import { Navigate } from 'react-router-dom';

const ProtectedRoute = ({ children, allowedRoles = [] }) => {
  
  // Проверяем наличие токена
  const token = localStorage.getItem('token');
  let userRole = (localStorage.getItem('userRole') || '').toLowerCase().trim();

  // Автовосстановление роли из JWT, если нет в localStorage
  if ((!userRole || userRole.length === 0) && token && token.startsWith('Bearer ')) {
    try {
      const payloadStr = token.split(' ')[1].split('.')[1];
      const json = JSON.parse(decodeURIComponent(escape(window.atob(payloadStr))));
      if (json?.role) {
        userRole = String(json.role).toLowerCase().trim();
        localStorage.setItem('userRole', userRole);
      }
    } catch (_) {
      // ignore
    }
  }

  if (!token || !token.startsWith('Bearer ')) {
    // Если нет токена или токен не в формате Bearer, перенаправляем на главную страницу
    // которая автоматически покажет страницу логина
    return <Navigate to="/" replace />;
  }

  // Если указаны разрешенные роли, проверяем роль пользователя
  if (allowedRoles.length > 0 && !allowedRoles.map(r => r.toLowerCase()).includes(userRole)) {
    // Если роль не разрешена, перенаправляем на dashboard
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};

export default ProtectedRoute;
