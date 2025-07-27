import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import AdminDashboard from './pages/AdminDashboard';
import PartnerLogin from './pages/PartnerLogin';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/admin/dashboard" element={<AdminDashboard />} />
        <Route path="/partner/login" element={<PartnerLogin />} />
        <Route path="/partner/forgot-password" element={<ForgotPassword />} />
        <Route path="/partner/reset-password" element={<ResetPassword />} />
      </Routes>
    </BrowserRouter>
  );
}
export default App; 