import React from 'react';
import { Navigate } from 'react-router-dom';

const ProtectedRoute = ({ children, allowedRoles = [] }) => {
  
  
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
      
    }
  }

  if (!token || !token.startsWith('Bearer ')) {
    
    
    return <Navigate to="/" replace />;
  }

  
  if (allowedRoles.length > 0 && !allowedRoles.map(r => r.toLowerCase()).includes(userRole)) {
    
    return <Navigate to="/dashboard" replace />;
  }

  return children;
};

export default ProtectedRoute;
