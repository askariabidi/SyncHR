// import React, { useState, useEffect } from 'react';
// import { useAuth } from '../context/AuthContext';
// import { attendanceAPI, leaveAPI } from '../services/api';
// import '../styles/Dashboard.css';

// export const EmployeeDashboard = () => {
//   const { user, logout } = useAuth();
//   const [attendanceToday, setAttendanceToday] = useState(null);
//   const [leaveBalance, setLeaveBalance] = useState([]);
//   const [loading, setLoading] = useState(true);
//   const [error, setError] = useState('');

//   useEffect(() => {
//     fetchDashboardData();
//   }, []);

//   const fetchDashboardData = async () => {
//     try {
//       setLoading(true);
//       setError('');

//       // Fetch leave balance
//       const balanceResponse = await leaveAPI.getBalance();
//       setLeaveBalance(balanceResponse.data.data.balances || []);
//     } catch (err) {
//       setError('Failed to load dashboard data');
//       console.error(err);
//     } finally {
//       setLoading(false);
//     }
//   };

//   const handleCheckIn = async () => {
//     try {
//       await attendanceAPI.checkIn(new Date().toISOString());
//       alert('✅ Check-in successful!');
//       fetchDashboardData();
//     } catch (err) {
//       alert('❌ Check-in failed: ' + (err.response?.data?.message || 'Unknown error'));
//     }
//   };

//   const handleCheckOut = async () => {
//     try {
//       await attendanceAPI.checkOut(new Date().toISOString());
//       alert('✅ Check-out successful!');
//       fetchDashboardData();
//     } catch (err) {
//       alert('❌ Check-out failed: ' + (err.response?.data?.message || 'Unknown error'));
//     }
//   };

//   if (loading) {
//     return <div className="dashboard-container"><p>Loading dashboard...</p></div>;
//   }

//   return (
//     <div className="dashboard-container">
//       {/* Header */}
//       <div className="dashboard-header">
//         <div>
//           <h1>Welcome, {user?.first_name}! 👋</h1>
//           <p className="dashboard-subtitle">Employee Dashboard</p>
//         </div>
//         <button className="btn-logout" onClick={logout}>Logout</button>
//       </div>

//       {error && <div className="error-banner">{error}</div>}

//       {/* Quick Actions */}
//       <div className="quick-actions">
//         <button className="action-btn check-in" onClick={handleCheckIn}>
//           🕐 Check In
//         </button>
//         <button className="action-btn check-out" onClick={handleCheckOut}>
//           🕑 Check Out
//         </button>
//         <button className="action-btn leave" onClick={() => window.location.href = '/dashboard/employee/leave'}>
//           📋 Apply Leave
//         </button>
//         <button className="action-btn requests" onClick={() => window.location.href = '/dashboard/employee/my-requests'}>
//           📑 My Requests
//         </button>
//         <button className="action-btn payslip" onClick={() => window.location.href = '/dashboard/employee/payslip'}>
//           📄 View Payslip
//         </button>
//       </div>

//       {/* Leave Balance Cards */}
//       <div className="section">
//         <h2>📊 Leave Balance</h2>
//         {leaveBalance.length > 0 ? (
//           <div className="cards-grid">
//             {leaveBalance.map((balance) => (
//               <div key={balance.id} className="card">
//                 <h3>{balance.leave_type}</h3>
//                 <div className="balance-display">
//                   <span className="balance-number">{balance.balance}</span>
//                   <span className="balance-text">days remaining</span>
//                 </div>
//               </div>
//             ))}
//           </div>
//         ) : (
//           <p>No leave data available</p>
//         )}
//       </div>

//       {/* User Info */}
//       <div className="section">
//         <h2>👤 Profile Information</h2>
//         <div className="info-grid">
//           <div className="info-item">
//             <label>Email:</label>
//             <span>{user?.email}</span>
//           </div>
//           <div className="info-item">
//             <label>Department:</label>
//             <span>{user?.department || 'N/A'}</span>
//           </div>
//           <div className="info-item">
//             <label>Phone:</label>
//             <span>{user?.phone_number || 'N/A'}</span>
//           </div>
//           <div className="info-item">
//             <label>Role:</label>
//             <span>{user?.role}</span>
//           </div>
//         </div>
//       </div>
//     </div>
//   );
// };

// export default EmployeeDashboard;

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
      console.log('📍 Attendance Response:', response.data);

      const records = response.data.data.attendance_records || [];
      console.log('📍 Records:', records);
      console.log('📍 Total records:', records.length);

      if (records.length > 0) {
        const today = new Date().toISOString().split('T')[0];
        console.log('📍 Today date:', today);

        const todayRecord = records.find((r) => {
          // Extract just the date part from the ISO timestamp
          const recordDate = new Date(r.date).toISOString().split('T')[0];
          console.log('📍 Checking record date:', recordDate, 'against today:', today);
          return recordDate === today;
        });

        console.log('📍 Today record found:', todayRecord);

        if (todayRecord) {
          setAttendance(todayRecord);
          console.log('📍 Check-in time:', todayRecord.check_in_time);
          console.log('📍 Check-out time:', todayRecord.check_out_time);

          if (todayRecord.check_in_time && !todayRecord.check_out_time) {
            console.log('✅ Setting checkedIn = true');
            setCheckedIn(true);
            setCheckInTime(todayRecord.check_in_time);
          } else {
            console.log('⭕ Setting checkedIn = false');
            setCheckedIn(false);
            setCheckInTime(null);
          }
        } else {
          console.log('⭕ No today record found');
          setAttendance(null);
          setCheckedIn(false);
          setCheckInTime(null);
        }
      }
    } catch (err) {
      console.error('❌ Failed to fetch attendance status:', err);
    }
  };

  const handleCheckIn = async () => {
    setActionLoading(true);
    try {
      await attendanceAPI.checkIn();
      alert('✅ Checked in successfully!');
      await fetchAttendanceStatus();
    } catch (err) {
      alert('❌ Failed to check in: ' + (err.response?.data?.message || 'Unknown error'));
    } finally {
      setActionLoading(false);
    }
  };

  const handleCheckOut = async () => {
    setActionLoading(true);
    try {
      await attendanceAPI.checkOut();
      alert('✅ Checked out successfully!');
      await fetchAttendanceStatus();
    } catch (err) {
      alert('❌ Failed to check out: ' + (err.response?.data?.message || 'Unknown error'));
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
          <h1>Welcome, {user?.first_name}! 👋</h1>
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
          <h3>📍 Today's Attendance</h3>
          {checkedIn ? (
            <div className="status-active">
              <span className="status-badge">✅ CHECKED IN</span>
              <p>Check-in Time: {formatTime(checkInTime)}</p>
            </div>
          ) : attendance?.check_out_time ? (
            <div className="status-completed">
              <span className="status-badge completed">✔️ COMPLETED</span>
              <p>
                Check-in: {formatTime(attendance.check_in_time)} | Check-out:{' '}
                {formatTime(attendance.check_out_time)}
              </p>
              <p>Hours Worked: {calculateHoursWorked()}</p>
            </div>
          ) : (
            <div className="status-inactive">
              <span className="status-badge inactive">⭕ NOT CHECKED IN</span>
              <p>Click "Check In" to start your day</p>
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
          {actionLoading ? '⏳ Processing...' : '🕐 Check In'}
        </button>
        <button
          className="action-btn check-out"
          onClick={handleCheckOut}
          disabled={!checkedIn || actionLoading}
          title={!checkedIn ? 'Check in first to check out' : 'Click to check out'}
        >
          {actionLoading ? '⏳ Processing...' : '🕑 Check Out'}
        </button>
        <button
          className="action-btn leave"
          onClick={() => (window.location.href = '/dashboard/employee/leave')}
        >
          📋 Apply Leave
        </button>
        <button
          className="action-btn requests"
          onClick={() => (window.location.href = '/dashboard/employee/my-requests')}
        >
          📑 My Requests
        </button>
        <button
          className="action-btn payslip"
          onClick={() => (window.location.href = '/dashboard/employee/payslip')}
        >
          📄 View Payslip
        </button>
      </div>

      {/* Leave Balance */}
      <div className="leave-balance-section">
        <h2>📊 Leave Balance</h2>
        {loading ? (
          <p>Loading...</p>
        ) : leaveBalance.length > 0 ? (
          <div className="balance-cards">
            {leaveBalance.map((balance) => (
              <div key={balance.id} className="balance-card">
                <h4>
                  {balance.leave_type === 'Sick Leave' && '🤒'}
                  {balance.leave_type === 'Casual Leave' && '📅'}
                  {balance.leave_type === 'Earned Leave' && '📚'}
                  {' ' + balance.leave_type}
                </h4>
                <div className="balance-value">{balance.balance}</div>
                <p>days remaining</p>
              </div>
            ))}
          </div>
        ) : (
          <p>No leave data available</p>
        )}
      </div>

      {/* Profile Information */}
      <div className="profile-section">
        <h2>👤 Profile Information</h2>
        <div className="profile-grid">
          <div className="profile-item">
            <label>Email:</label>
            <span>{user?.email}</span>
          </div>
          <div className="profile-item">
            <label>Department:</label>
            <span>{user?.department}</span>
          </div>
          <div className="profile-item">
            <label>Phone:</label>
            <span>{user?.phone_number}</span>
          </div>
          <div className="profile-item">
            <label>Role:</label>
            <span>{user?.role === 'employee' ? 'Employee' : 'HR Manager'}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmployeeDashboard;