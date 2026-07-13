import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { leaveAPI, attendanceAPI, authAPI } from '../services/api';
import '../styles/HRDashboard.css';

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

const getTodayISO = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
};

export const HRDashboard = () => {
  const { user, logout } = useAuth();

  const [leaveRequests, setLeaveRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('pending'); // pending, approved, rejected, all
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [approvalNotes, setApprovalNotes] = useState('');
  const [actionInProgress, setActionInProgress] = useState(false);
  // for attendance tracking of all employees
  const [attendanceRecords, setAttendanceRecords] = useState([]);
  const [attendanceLoading, setAttendanceLoading] = useState(false);
  const todayISO = getTodayISO();
  const [selectedDate, setSelectedDate] = useState(todayISO);
  const [viewYear, setViewYear] = useState(new Date().getFullYear());
  const [viewMonth, setViewMonth] = useState(new Date().getMonth()); // 0-11
  // for the employee directory
  const [employees, setEmployees] = useState([]);
  const [employeesLoading, setEmployeesLoading] = useState(false);
  const [employeeSearch, setEmployeeSearch] = useState('');

  // Fetch leave requests and employees on mount
  useEffect(() => {
    fetchLeaveRequests();
    fetchEmployees();
  }, []);

  // Fetch attendance whenever the selected date changes (defaults to today)
  useEffect(() => {
    fetchAttendanceRecords(selectedDate);
  }, [selectedDate]);

  const fetchLeaveRequests = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await leaveAPI.getRequests();
      setLeaveRequests(response.data.data.leave_requests || []);
    } catch (err) {
      setError('Failed to load leave requests');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Filter requests
  const filteredRequests = leaveRequests.filter((request) => {
    if (filter === 'all') return true;
    return request.status === filter;
  });

  // Handle approve
  const handleApprove = async (requestId) => {
    if (!approvalNotes.trim()) {
      alert('Please add approval notes');
      return;
    }

    setActionInProgress(true);
    try {
      await leaveAPI.approveLeave(requestId, approvalNotes);
      setSelectedRequest(null);
      setApprovalNotes('');
      fetchLeaveRequests();
    } catch (err) {
      alert('Failed to approve: ' + (err.response?.data?.message || 'Unknown error'));
    } finally {
      setActionInProgress(false);
    }
  };

  // Handle reject
  const handleReject = async (requestId) => {
    if (!approvalNotes.trim()) {
      alert('Please add rejection reason');
      return;
    }

    setActionInProgress(true);
    try {
      await leaveAPI.rejectLeave(requestId, approvalNotes);
      setSelectedRequest(null);
      setApprovalNotes('');
      fetchLeaveRequests();
    } catch (err) {
      alert('Failed to reject: ' + (err.response?.data?.message || 'Unknown error'));
    } finally {
      setActionInProgress(false);
    }
  };

  // Get status badge
  const getStatusBadge = (status) => {
    switch (status) {
      case 'approved':
        return 'badge-approved';
      case 'rejected':
        return 'badge-rejected';
      case 'pending':
        return 'badge-pending';
      default:
        return 'badge-default';
    }
  };

  const fetchAttendanceRecords = async (date) => {
    try {
      setAttendanceLoading(true);
      const response = await attendanceAPI.getAttendanceRecords(date);
      setAttendanceRecords(response.data.data.attendance_records || []);
    } catch (err) {
      console.error('Failed to fetch attendance records:', err);
    } finally {
      setAttendanceLoading(false);
    }
  };

  // Whether the month currently shown in the day tabs is the real-world current month
  const isCurrentMonthView = () => {
    const now = new Date();
    return viewYear === now.getFullYear() && viewMonth === now.getMonth();
  };

  const handlePrevMonth = () => {
    if (viewMonth === 0) {
      setViewYear((y) => y - 1);
      setViewMonth(11);
    } else {
      setViewMonth((m) => m - 1);
    }
  };

  const handleNextMonth = () => {
    if (isCurrentMonthView()) return; // no browsing into the future
    if (viewMonth === 11) {
      setViewYear((y) => y + 1);
      setViewMonth(0);
    } else {
      setViewMonth((m) => m + 1);
    }
  };

  const buildDateStr = (year, month, day) =>
    `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;

  const handleDayClick = (day) => {
    const dateStr = buildDateStr(viewYear, viewMonth, day);
    if (dateStr > todayISO) return;
    setSelectedDate(dateStr);
  };

  const handleDatePickerChange = (e) => {
    const value = e.target.value;
    if (!value) return;
    setSelectedDate(value);
    const [y, m] = value.split('-').map(Number);
    setViewYear(y);
    setViewMonth(m - 1);
  };

  const daysInViewMonth = new Date(viewYear, viewMonth + 1, 0).getDate();

  const fetchEmployees = async () => {
    try {
      setEmployeesLoading(true);
      const response = await authAPI.getAllEmployees();
      setEmployees(response.data.data.employees || []);
    } catch (err) {
      console.error('Failed to fetch employees:', err);
    } finally {
      setEmployeesLoading(false);
    }
  };

  const getEmployeeName = (userId) => {
    const employee = employees.find((e) => e.id === userId);
    return employee ? `${employee.first_name} ${employee.last_name}` : `Employee ${userId}`;
  };

  // Filter employees by search query (name, email, or department)
  const filteredEmployees = employees.filter((employee) => {
    const query = employeeSearch.trim().toLowerCase();
    if (!query) return true;
    const fullName = `${employee.first_name} ${employee.last_name}`.toLowerCase();
    return (
      fullName.includes(query) ||
      employee.email.toLowerCase().includes(query) ||
      employee.department.toLowerCase().includes(query)
    );
  });

  // The backend sends Go's zero-value time ("0001-01-01T00:00:00Z") instead of
  // null when an employee hasn't checked out yet, so treat pre-1970 dates as absent
  const isValidTimestamp = (timestamp) => {
    if (!timestamp) return false;
    return new Date(timestamp).getFullYear() > 1970;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    const day = date.getDate();
    const suffix =
      day % 10 === 1 && day !== 11
        ? 'st'
        : day % 10 === 2 && day !== 12
        ? 'nd'
        : day % 10 === 3 && day !== 13
        ? 'rd'
        : 'th';
    const month = date.toLocaleString('en-US', { month: 'long' });
    return `${day}${suffix} ${month} ${date.getFullYear()}`;
  };

  const formatDateWithWeekday = (dateStr) => {
    if (!dateStr) return '-';
    const weekday = new Date(dateStr).toLocaleString('en-US', { weekday: 'long' });
    return `${weekday}, ${formatDate(dateStr)}`;
  };

  const formatTime24 = (timestamp) => {
    if (!isValidTimestamp(timestamp)) return '-';
    return new Date(timestamp).toLocaleTimeString('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  };

  const formatDuration = (checkInStr, checkOutStr) => {
    if (!isValidTimestamp(checkInStr) || !isValidTimestamp(checkOutStr)) return '-';
    const diffMs = new Date(checkOutStr) - new Date(checkInStr);
    if (diffMs <= 0) return '-';
    const totalMinutes = Math.round(diffMs / 60000);
    if (totalMinutes < 60) return '< 1 Hour';
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    return `${hours} HR ${minutes} MINS`;
  };

  return (
    <div className="hr-dashboard-container">
      {/* Header */}
      <div className="hr-header">
        <div>
          <h1>HR Manager Dashboard</h1>
          <p>Manage employee leave requests</p>
        </div>
        <button className="btn-logout" onClick={logout}>
          Logout
        </button>
      </div>

      {/* Stats */}
      <div className="stats-grid">
        <div className="stat-card pending-stat">
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'pending').length}
            </div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
        <div className="stat-card approved-stat">
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'approved').length}
            </div>
            <div className="stat-label">Approved</div>
          </div>
        </div>
        <div className="stat-card rejected-stat">
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'rejected').length}
            </div>
            <div className="stat-label">Rejected</div>
          </div>
        </div>
        <div className="stat-card total-stat">
          <div className="stat-info">
            <div className="stat-number">{leaveRequests.length}</div>
            <div className="stat-label">Total</div>
          </div>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="filter-tabs">
        <button
          className={`tab ${filter === 'pending' ? 'active' : ''}`}
          onClick={() => setFilter('pending')}
        >
          Pending ({leaveRequests.filter((r) => r.status === 'pending').length})
        </button>
        <button
          className={`tab ${filter === 'approved' ? 'active' : ''}`}
          onClick={() => setFilter('approved')}
        >
          Approved ({leaveRequests.filter((r) => r.status === 'approved').length})
        </button>
        <button
          className={`tab ${filter === 'rejected' ? 'active' : ''}`}
          onClick={() => setFilter('rejected')}
        >
          Rejected ({leaveRequests.filter((r) => r.status === 'rejected').length})
        </button>
        <button
          className={`tab ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          All ({leaveRequests.length})
        </button>
      </div>

      {/* Error Message */}
      {error && <div className="error-banner">{error}</div>}

      {/* Loading State */}
      {loading && <div className="loading">Loading leave requests...</div>}

      {/* Leave Requests Cards */}
      {!loading && filteredRequests.length > 0 ? (
        <div className="requests-cards">
          {filteredRequests.map((request) => (
            <div key={request.id} className={`request-card ${getStatusBadge(request.status)}`}>
              <div className="card-header">
                <div className="card-title">
                  <div className="title-info">
                    <h3>{request.employee_first_name} {request.employee_last_name}</h3>
                    <p>ID: {request.user_id} &middot; {request.employee_department}</p>
                  </div>
                </div>
                <span className="status-badge">{request.status.toUpperCase()}</span>
              </div>

              <div className="card-body">
                <div className="dates-row">
                  <div className="date-item">
                    <label>Start Date</label>
                    <span>{formatDateWithWeekday(request.start_date)}</span>
                  </div>
                  <div className="date-item">
                    <label>End Date</label>
                    <span>{formatDateWithWeekday(request.end_date)}</span>
                  </div>
                  <div className="date-item days-item">
                    <label>Days</label>
                    <span className="days-badge">{request.number_of_days}</span>
                  </div>
                </div>

                <div className="reason-section">
                  <label>Reason</label>
                  <p>{request.reason || 'No reason provided'}</p>
                </div>

                {request.approval_notes && (
                  <div className="notes-section">
                    <label>HR Notes</label>
                    <p>{request.approval_notes}</p>
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              {request.status === 'pending' && (
                <div className="card-actions">
                  <button
                    className="btn-approve"
                    onClick={() => setSelectedRequest(request.id)}
                  >
                    Approve
                  </button>
                  <button
                    className="btn-reject"
                    onClick={() => {
                      setSelectedRequest(request.id);
                    }}
                  >
                    Reject
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      ) : !loading ? (
        <div className="empty-state">
          <p>No leave requests found</p>
        </div>
      ) : null}

      {/* Attendance Section */}
      <div className="attendance-section">
        <div className="attendance-section-header">
          <h2>Attendance Report</h2>
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

        {attendanceLoading ? (
          <div className="loading">Loading attendance records...</div>
        ) : attendanceRecords.length > 0 ? (
          <div className="attendance-table-container">
            <table className="attendance-table">
              <thead>
                <tr>
                  <th>Employee ID</th>
                  <th>Employee Name</th>
                  <th>Date</th>
                  <th>Check In</th>
                  <th>Check Out</th>
                  <th>Duration</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {attendanceRecords.map((record, index) => (
                  <tr key={index}>
                    <td>{record.user_id}</td>
                    <td className="employee-name">{getEmployeeName(record.user_id)}</td>
                    <td>{formatDate(record.date)}</td>
                    <td>{formatTime24(record.check_in_time)}</td>
                    <td>{formatTime24(record.check_out_time)}</td>
                    <td className="duration-cell">
                      {formatDuration(record.check_in_time, record.check_out_time)}
                    </td>
                    <td>
                      <span className={`status-badge attendance-${record.status}`}>
                        {record.status === 'checked_in' && 'Checked In'}
                        {record.status === 'checked_out' && 'Completed'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state">
            <p>No attendance records found</p>
          </div>
        )}
      </div>

      {/* Employees Section */}
      <div className="employees-section">
        <h2>All Employees</h2>

        <input
          type="text"
          className="employee-search-input"
          placeholder="Search by name, email, or department..."
          value={employeeSearch}
          onChange={(e) => setEmployeeSearch(e.target.value)}
        />

        {employeesLoading ? (
          <div className="loading">Loading employees...</div>
        ) : filteredEmployees.length > 0 ? (
          <div className="employees-table-container">
            <table className="employees-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Department</th>
                  <th>Phone</th>
                </tr>
              </thead>
              <tbody>
                {filteredEmployees.map((employee) => (
                  <tr key={employee.id}>
                    <td>{employee.id}</td>
                    <td className="employee-name">{employee.first_name} {employee.last_name}</td>
                    <td>{employee.email}</td>
                    <td>
                      <span className={`role-badge role-${employee.role}`}>
                        {employee.role === 'hr_manager' ? 'HR Manager' : 'Employee'}
                      </span>
                    </td>
                    <td>{employee.department}</td>
                    <td>{employee.phone_number}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state">
            <p>No employees found</p>
          </div>
        )}
      </div>

      {/* Modal for Approval/Rejection */}
      {selectedRequest && (
        <div className="modal-overlay" onClick={() => setSelectedRequest(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Leave Request Decision</h2>
              <button className="modal-close" onClick={() => setSelectedRequest(null)}>
                &times;
              </button>
            </div>

            <div className="modal-body">
              <label>Add Notes</label>
              <textarea
                value={approvalNotes}
                onChange={(e) => setApprovalNotes(e.target.value)}
                placeholder="Enter your approval/rejection notes..."
                rows="4"
              />
            </div>

            <div className="modal-footer">
              <button
                className="btn-modal-reject"
                onClick={() => handleReject(selectedRequest)}
                disabled={actionInProgress}
              >
                {actionInProgress ? 'Processing...' : 'Reject'}
              </button>
              <button
                className="btn-modal-approve"
                onClick={() => handleApprove(selectedRequest)}
                disabled={actionInProgress}
              >
                {actionInProgress ? 'Processing...' : 'Approve'}
              </button>
              <button
                className="btn-modal-cancel"
                onClick={() => setSelectedRequest(null)}
                disabled={actionInProgress}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default HRDashboard;
