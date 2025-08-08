import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';

const ProtectedRoute = ({ children, allowedRoles = [] }) => {
  const location = useLocation();
  
  // Проверяем наличие токена
  const token = localStorage.getItem('token');
  const userRole = localStorage.getItem('userRole');

  if (!token || !token.startsWith('Bearer ')) {
    // Если нет токена или токен не в формате Bearer, перенаправляем на главную страницу
    // которая автоматически покажет страницу логина
    return <Navigate to="/" replace />;
  }

  // Если указаны разрешенные роли, проверяем роль пользователя
  if (allowedRoles.length > 0 && !allowedRoles.includes(userRole)) {
    // Если роль не разрешена, перенаправляем на dashboard
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};

export default ProtectedRoute;
