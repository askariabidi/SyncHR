import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { NotificationProvider } from './context/NotificationContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import Login from './pages/Login';
import EmployeeDashboard from './pages/EmployeeDashboard';
import ApplyLeave from './pages/ApplyLeave';
import MyLeaveRequests from './pages/MyLeaveRequests';
import HRDashboard from './pages/HRDashboard';
import ViewPayslip from './pages/ViewPayslip';
import './App.css';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <NotificationProvider>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<Login />} />

            {/* Protected Routes - Employee */}
            <Route
              path="/dashboard/employee"
              element={
                <ProtectedRoute allowedRoles={['employee']}>
                  <EmployeeDashboard />
                </ProtectedRoute>
              }
            />

            {/* Protected Routes - Apply Leave */}
            <Route
              path="/dashboard/employee/leave"
              element={
                <ProtectedRoute allowedRoles={['employee']}>
                  <ApplyLeave />
                </ProtectedRoute>
              }
            />

            {/* Protected Routes - HR Manager Dashboard */}
            <Route
              path="/dashboard/hr"
              element={
                <ProtectedRoute allowedRoles={['hr_manager']}>
                  <HRDashboard />
                </ProtectedRoute>
              }
            />

            {/* Protected Routes - My Leave Requests */}
            <Route
              path="/dashboard/employee/my-requests"
              element={
                <ProtectedRoute allowedRoles={['employee']}>
                  <MyLeaveRequests />
                </ProtectedRoute>
              }
            />

            {/* Protected Routes - View Payslip */}
            <Route
              path="/dashboard/employee/payslip"
              element={
                <ProtectedRoute allowedRoles={['employee']}>
                  <ViewPayslip />
                </ProtectedRoute>
              }
            />

            {/* Redirect root to login */}
            <Route path="/" element={<Navigate to="/login" replace />} />

            {/* Catch all - redirect to login */}
            <Route path="*" element={<Navigate to="/login" replace />} />
          </Routes>
        </NotificationProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
