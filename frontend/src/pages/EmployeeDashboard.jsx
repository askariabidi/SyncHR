import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { attendanceAPI, leaveAPI } from '../services/api';
import '../styles/Dashboard.css';

export const EmployeeDashboard = () => {
  const { user, logout } = useAuth();

  const [leaveBalance, setLeaveBalance] = useState([]);
  const [attendance, setAttendance] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [checkedIn, setCheckedIn] = useState(false);
  const [checkInTime, setCheckInTime] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);

  // Fetch data on mount
  useEffect(() => {
    fetchData();
    // Check attendance status every minute
    const interval = setInterval(fetchAttendanceStatus, 60000);
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError('');
      const balanceResponse = await leaveAPI.getBalance();
      setLeaveBalance(balanceResponse.data.data.balances || []);
      await fetchAttendanceStatus();
    } catch (err) {
      setError('Failed to load dashboard data');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const fetchAttendanceStatus = async () => {
    try {
      const response = await attendanceAPI.getAttendanceHistory();
      const records = response.data.data.attendance_records || [];

      if (records.length > 0) {
        const today = new Date().toISOString().split('T')[0];

        const todayRecord = records.find((r) => {
          const recordDate = new Date(r.date).toISOString().split('T')[0];
          return recordDate === today;
        });

        if (todayRecord) {
          setAttendance(todayRecord);

          if (todayRecord.check_in_time && !todayRecord.check_out_time) {
            setCheckedIn(true);
            setCheckInTime(todayRecord.check_in_time);
          } else {
            setCheckedIn(false);
            setCheckInTime(null);
          }
        } else {
          setAttendance(null);
          setCheckedIn(false);
          setCheckInTime(null);
        }
      }
    } catch (err) {
      console.error('Failed to fetch attendance status:', err);
    }
  };

  const handleCheckIn = async () => {
    setActionLoading(true);
    try {
      await attendanceAPI.checkIn();
      await fetchAttendanceStatus();
    } catch (err) {
      alert('Failed to check in: ' + (err.response?.data?.message || 'Unknown error'));
    } finally {
      setActionLoading(false);
    }
  };

  const handleCheckOut = async () => {
    setActionLoading(true);
    try {
      await attendanceAPI.checkOut();
      await fetchAttendanceStatus();
    } catch (err) {
      alert('Failed to check out: ' + (err.response?.data?.message || 'Unknown error'));
    } finally {
      setActionLoading(false);
    }
  };

  // Format time
  const formatTime = (timeString) => {
    if (!timeString) return '-';
    const time = new Date(timeString);
    return time.toLocaleTimeString('en-IN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  // Calculate hours worked
  const calculateHoursWorked = () => {
    if (!attendance || !attendance.check_in_time || !attendance.check_out_time) {
      return '-';
    }
    const checkIn = new Date(attendance.check_in_time);
    const checkOut = new Date(attendance.check_out_time);
    const hours = ((checkOut - checkIn) / (1000 * 60 * 60)).toFixed(2);
    return `${hours}h`;
  };

  return (
    <div className="dashboard-container">
      {/* Header */}
      <div className="dashboard-header">
        <div>
          <h1>Welcome, {user?.first_name}</h1>
          <p>Employee Dashboard</p>
        </div>
        <button className="btn-logout" onClick={logout}>
          Logout
        </button>
      </div>

      {/* Error Message */}
      {error && <div className="error-banner">{error}</div>}

      {/* Attendance Status */}
      <div className="attendance-status">
        <div className="status-info">
          <h3>Today's Attendance</h3>
          {checkedIn ? (
            <div className="status-active">
              <span className="status-badge">Checked in</span>
              <p>Check-in time: {formatTime(checkInTime)}</p>
            </div>
          ) : attendance?.check_out_time ? (
            <div className="status-completed">
              <span className="status-badge completed">Completed</span>
              <p>
                Check-in: {formatTime(attendance.check_in_time)} &middot; Check-out:{' '}
                {formatTime(attendance.check_out_time)}
              </p>
              <p>Hours worked: {calculateHoursWorked()}</p>
            </div>
          ) : (
            <div className="status-inactive">
              <span className="status-badge inactive">Not checked in</span>
              <p>Check in to start your day</p>
            </div>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="quick-actions">
        <button
          className="action-btn check-in"
          onClick={handleCheckIn}
          disabled={checkedIn || actionLoading}
          title={checkedIn ? 'Already checked in' : 'Click to check in'}
        >
          {actionLoading ? 'Processing...' : 'Check In'}
        </button>
        <button
          className="action-btn check-out"
          onClick={handleCheckOut}
          disabled={!checkedIn || actionLoading}
          title={!checkedIn ? 'Check in first to check out' : 'Click to check out'}
        >
          {actionLoading ? 'Processing...' : 'Check Out'}
        </button>
        <button
          className="action-btn leave"
          onClick={() => (window.location.href = '/dashboard/employee/leave')}
        >
          Apply Leave
        </button>
        <button
          className="action-btn requests"
          onClick={() => (window.location.href = '/dashboard/employee/my-requests')}
        >
          My Requests
        </button>
        <button
          className="action-btn payslip"
          onClick={() => (window.location.href = '/dashboard/employee/payslip')}
        >
          View Payslip
        </button>
      </div>

      {/* Leave Balance */}
      <div className="leave-balance-section">
        <h2>Leave Balance</h2>
        {loading ? (
          <p className="muted-text">Loading...</p>
        ) : leaveBalance.length > 0 ? (
          <div className="balance-cards">
            {leaveBalance.map((balance) => (
              <div key={balance.id} className="balance-card">
                <h4>{balance.leave_type}</h4>
                <div className="balance-value">{balance.balance}</div>
                <p>days remaining</p>
              </div>
            ))}
          </div>
        ) : (
          <p className="muted-text">No leave data available</p>
        )}
      </div>

      {/* Profile Information */}
      <div className="profile-section">
        <h2>Profile Information</h2>
        <div className="profile-grid">
          <div className="profile-item">
            <label>Email</label>
            <span>{user?.email}</span>
          </div>
          <div className="profile-item">
            <label>Department</label>
            <span>{user?.department}</span>
          </div>
          <div className="profile-item">
            <label>Phone</label>
            <span>{user?.phone_number}</span>
          </div>
          <div className="profile-item">
            <label>Role</label>
            <span>{user?.role === 'employee' ? 'Employee' : 'HR Manager'}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmployeeDashboard;
