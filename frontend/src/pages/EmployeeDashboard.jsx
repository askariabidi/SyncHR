import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../hooks/useAuth';
import { attendanceAPI, leaveAPI } from '../services/api';
import '../styles/Dashboard.css';
import '../styles/AttendanceCalendar.css';
import {
  MONTH_NAMES,
  buildDateStr,
  isValidTimestamp,
  formatDateWithWeekday,
  formatTime24,
  formatDuration,
} from '../utils/dateFormat';
import { useDateNavigator } from '../hooks/useDateNavigator';
import { NotificationBell } from '../components/NotificationBell';

export const EmployeeDashboard = () => {
  const { user, logout } = useAuth();

  const [leaveBalance, setLeaveBalance] = useState([]);
  const [attendance, setAttendance] = useState(null);
  const [attendanceHistory, setAttendanceHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [checkedIn, setCheckedIn] = useState(false);
  const [checkInTime, setCheckInTime] = useState(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const {
    todayISO,
    selectedDate,
    viewYear,
    viewMonth,
    isCurrentMonthView,
    handlePrevMonth,
    handleNextMonth,
    handleDayClick,
    handleDatePickerChange,
    daysInViewMonth,
  } = useDateNavigator();

  const fetchAttendanceStatus = useCallback(async () => {
    try {
      const response = await attendanceAPI.getAttendanceHistory();
      const records = response.data.data.attendance_records || [];
      setAttendanceHistory(records);

      if (records.length > 0) {
        const today = new Date().toISOString().split('T')[0];

        const todayRecord = records.find((r) => {
          const recordDate = new Date(r.date).toISOString().split('T')[0];
          return recordDate === today;
        });

        if (todayRecord) {
          setAttendance(todayRecord);

          if (isValidTimestamp(todayRecord.check_in_time) && !isValidTimestamp(todayRecord.check_out_time)) {
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
      } else {
        setAttendance(null);
        setCheckedIn(false);
        setCheckInTime(null);
      }
    } catch (err) {
      console.error('Failed to fetch attendance status:', err);
    }
  }, []);

  const fetchData = useCallback(async () => {
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
  }, [fetchAttendanceStatus]);

  // Fetch data on mount, then re-check attendance status every minute
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchAttendanceStatus, 60000);
    return () => clearInterval(interval);
  }, [fetchData, fetchAttendanceStatus]);

  // Live-ticking timer while checked in
  useEffect(() => {
    if (!checkedIn || !isValidTimestamp(checkInTime)) {
      setElapsedSeconds(0);
      return;
    }
    const startMs = new Date(checkInTime).getTime();
    const tick = () => setElapsedSeconds(Math.max(0, Math.floor((Date.now() - startMs) / 1000)));
    tick();
    const timer = setInterval(tick, 1000);
    return () => clearInterval(timer);
  }, [checkedIn, checkInTime]);

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

  // Confirm before checking out - this ends today's attendance session and can't be undone
  const handleCheckOutClick = () => {
    if (window.confirm("Check out now? This will end today's attendance session and cannot be undone.")) {
      handleCheckOut();
    }
  };

  // Format the elapsed check-in time as HH:MM:SS
  const formatElapsed = (totalSeconds) => {
    const h = String(Math.floor(totalSeconds / 3600)).padStart(2, '0');
    const m = String(Math.floor((totalSeconds % 3600) / 60)).padStart(2, '0');
    const s = String(totalSeconds % 60).padStart(2, '0');
    return `${h}:${m}:${s}`;
  };

  const fullName = [user?.first_name, user?.last_name].filter(Boolean).join(' ');

  // The record (if any) for whichever date is selected in the "My Attendance" section
  const selectedRecord = attendanceHistory.find(
    (r) => r.date && r.date.split('T')[0] === selectedDate
  );

  return (
    <div className="dashboard-container">
      {/* Header */}
      <div className="dashboard-header">
        <div>
          <h1>Welcome, {fullName || user?.first_name}</h1>
          <p>Employee Dashboard</p>
        </div>
        <div className="header-actions">
          <NotificationBell />
          <button className="btn-logout" onClick={logout}>
            Logout
          </button>
        </div>
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
              <p>Check-in time: {formatTime24(checkInTime)}</p>
              <div className="live-timer">{formatElapsed(elapsedSeconds)}</div>
            </div>
          ) : isValidTimestamp(attendance?.check_out_time) ? (
            <div className="status-completed">
              <span className="status-badge completed">Completed</span>
              <p>
                Check-in: {formatTime24(attendance.check_in_time)} &middot; Check-out:{' '}
                {formatTime24(attendance.check_out_time)}
              </p>
              <p>Hours worked: {formatDuration(attendance.check_in_time, attendance.check_out_time)}</p>
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
          onClick={handleCheckOutClick}
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

      {/* My Attendance */}
      <div className="my-attendance-section">
        <div className="attendance-section-header">
          <h2>My Attendance</h2>
          <input
            type="date"
            className="date-picker-input"
            value={selectedDate}
            max={todayISO}
            onChange={handleDatePickerChange}
          />
        </div>

        <div className="month-nav">
          <button className="month-nav-btn" onClick={handlePrevMonth} aria-label="Previous month">
            &lsaquo;
          </button>
          <span className="month-nav-label">{MONTH_NAMES[viewMonth]} {viewYear}</span>
          <button
            className="month-nav-btn"
            onClick={handleNextMonth}
            disabled={isCurrentMonthView()}
            aria-label="Next month"
          >
            &rsaquo;
          </button>
        </div>

        <div className="day-tabs">
          {Array.from({ length: daysInViewMonth }, (_, i) => i + 1).map((day) => {
            const dateStr = buildDateStr(viewYear, viewMonth, day);
            const isFuture = dateStr > todayISO;
            const dayOfWeek = new Date(viewYear, viewMonth, day).getDay();
            const isWeekend = dayOfWeek === 0 || dayOfWeek === 6;
            return (
              <button
                key={day}
                className={`day-tab ${isWeekend ? 'weekend' : ''} ${dateStr === selectedDate ? 'active' : ''} ${dateStr === todayISO ? 'today' : ''}`}
                onClick={() => handleDayClick(day)}
                disabled={isFuture}
              >
                {day}
              </button>
            );
          })}
        </div>

        <p className="attendance-selected-date">
          Showing attendance for <strong>{formatDateWithWeekday(selectedDate)}</strong>
          {selectedDate === todayISO && ' (Today)'}
        </p>

        {selectedRecord ? (
          <div className="attendance-day-detail">
            <div className="attendance-day-item">
              <label>Check In</label>
              <span>{formatTime24(selectedRecord.check_in_time)}</span>
            </div>
            <div className="attendance-day-item">
              <label>Check Out</label>
              <span>{formatTime24(selectedRecord.check_out_time)}</span>
            </div>
            <div className="attendance-day-item">
              <label>Duration</label>
              <span>{formatDuration(selectedRecord.check_in_time, selectedRecord.check_out_time)}</span>
            </div>
            <div className="attendance-day-item">
              <label>Status</label>
              <span className={`status-badge ${selectedRecord.status === 'checked_out' ? 'completed' : ''}`}>
                {selectedRecord.status === 'checked_in' ? 'Checked In' : 'Completed'}
              </span>
            </div>
          </div>
        ) : (
          <p className="muted-text">No attendance record for this date</p>
        )}
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
