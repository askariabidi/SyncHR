import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { attendanceAPI, leaveAPI } from '../services/api';
import '../styles/Dashboard.css';

export const EmployeeDashboard = () => {
  const { user, logout } = useAuth();
  const [attendanceToday, setAttendanceToday] = useState(null);
  const [leaveBalance, setLeaveBalance] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      setError('');

      // Fetch leave balance
      const balanceResponse = await leaveAPI.getBalance();
      setLeaveBalance(balanceResponse.data.data.balances || []);
    } catch (err) {
      setError('Failed to load dashboard data');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCheckIn = async () => {
    try {
      await attendanceAPI.checkIn(new Date().toISOString());
      alert('✅ Check-in successful!');
      fetchDashboardData();
    } catch (err) {
      alert('❌ Check-in failed: ' + (err.response?.data?.message || 'Unknown error'));
    }
  };

  const handleCheckOut = async () => {
    try {
      await attendanceAPI.checkOut(new Date().toISOString());
      alert('✅ Check-out successful!');
      fetchDashboardData();
    } catch (err) {
      alert('❌ Check-out failed: ' + (err.response?.data?.message || 'Unknown error'));
    }
  };

  if (loading) {
    return <div className="dashboard-container"><p>Loading dashboard...</p></div>;
  }

  return (
    <div className="dashboard-container">
      {/* Header */}
      <div className="dashboard-header">
        <div>
          <h1>Welcome, {user?.first_name}! 👋</h1>
          <p className="dashboard-subtitle">Employee Dashboard</p>
        </div>
        <button className="btn-logout" onClick={logout}>Logout</button>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {/* Quick Actions */}
      <div className="quick-actions">
        <button className="action-btn check-in" onClick={handleCheckIn}>
          🕐 Check In
        </button>
        <button className="action-btn check-out" onClick={handleCheckOut}>
          🕑 Check Out
        </button>
        <button className="action-btn leave" onClick={() => window.location.href = '/dashboard/employee/leave'}>
          📋 Apply Leave
        </button>
        <button className="action-btn requests" onClick={() => window.location.href = '/dashboard/employee/my-requests'}>
          📑 My Requests
        </button>
        <button className="action-btn payslip" onClick={() => window.location.href = '/dashboard/employee/payslip'}>
          📄 View Payslip
        </button>
      </div>

      {/* Leave Balance Cards */}
      <div className="section">
        <h2>📊 Leave Balance</h2>
        {leaveBalance.length > 0 ? (
          <div className="cards-grid">
            {leaveBalance.map((balance) => (
              <div key={balance.id} className="card">
                <h3>{balance.leave_type}</h3>
                <div className="balance-display">
                  <span className="balance-number">{balance.balance}</span>
                  <span className="balance-text">days remaining</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p>No leave data available</p>
        )}
      </div>

      {/* User Info */}
      <div className="section">
        <h2>👤 Profile Information</h2>
        <div className="info-grid">
          <div className="info-item">
            <label>Email:</label>
            <span>{user?.email}</span>
          </div>
          <div className="info-item">
            <label>Department:</label>
            <span>{user?.department || 'N/A'}</span>
          </div>
          <div className="info-item">
            <label>Phone:</label>
            <span>{user?.phone_number || 'N/A'}</span>
          </div>
          <div className="info-item">
            <label>Role:</label>
            <span>{user?.role}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmployeeDashboard;