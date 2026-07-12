import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { leaveAPI } from '../services/api';
import '../styles/HRDashboard.css';

export const HRDashboard = () => {
  const { user, logout } = useAuth();

  const [leaveRequests, setLeaveRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('pending'); // pending, approved, rejected, all
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [approvalNotes, setApprovalNotes] = useState('');
  const [actionInProgress, setActionInProgress] = useState(false);

  // Fetch leave requests on refresh
  useEffect(() => {
    fetchLeaveRequests();
  }, []);

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
      alert('✅ Leave request approved successfully!');
      setSelectedRequest(null);
      setApprovalNotes('');
      fetchLeaveRequests();
    } catch (err) {
      alert('❌ Failed to approve: ' + (err.response?.data?.message || 'Unknown error'));
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
      alert('❌ Leave request rejected successfully!');
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

  // Get status icon
  const getStatusIcon = (status) => {
    switch (status) {
      case 'approved':
        return '✅';
      case 'rejected':
        return '❌';
      case 'pending':
        return '⏳';
      default:
        return '📋';
    }
  };

  // Get leave type name
  const getLeaveTypeName = (typeId) => {
    switch (typeId) {
      case 1:
        return '🤒 Sick Leave';
      case 2:
        return '📅 Casual Leave';
      case 3:
        return '📚 Earned Leave';
      default:
        return 'Leave';
    }
  };

  return (
    <div className="hr-dashboard-container">
      {/* Header */}
      <div className="hr-header">
        <div>
          <h1>👔 HR Manager Dashboard</h1>
          <p>Manage employee leave requests</p>
        </div>
        <button className="btn-logout" onClick={logout}>
          Logout
        </button>
      </div>

      {/* Stats */}
      <div className="stats-grid">
        <div className="stat-card pending-stat">
          <div className="stat-icon">⏳</div>
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'pending').length}
            </div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
        <div className="stat-card approved-stat">
          <div className="stat-icon">✅</div>
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'approved').length}
            </div>
            <div className="stat-label">Approved</div>
          </div>
        </div>
        <div className="stat-card rejected-stat">
          <div className="stat-icon">❌</div>
          <div className="stat-info">
            <div className="stat-number">
              {leaveRequests.filter((r) => r.status === 'rejected').length}
            </div>
            <div className="stat-label">Rejected</div>
          </div>
        </div>
        <div className="stat-card total-stat">
          <div className="stat-icon">📋</div>
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
          ⏳ Pending ({leaveRequests.filter((r) => r.status === 'pending').length})
        </button>
        <button
          className={`tab ${filter === 'approved' ? 'active' : ''}`}
          onClick={() => setFilter('approved')}
        >
          ✅ Approved ({leaveRequests.filter((r) => r.status === 'approved').length})
        </button>
        <button
          className={`tab ${filter === 'rejected' ? 'active' : ''}`}
          onClick={() => setFilter('rejected')}
        >
          ❌ Rejected ({leaveRequests.filter((r) => r.status === 'rejected').length})
        </button>
        <button
          className={`tab ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          📋 All ({leaveRequests.length})
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
                  <span className="status-icon">{getStatusIcon(request.status)}</span>
                  <div className="title-info">
                    <h3>{request.employee_first_name} {request.employee_last_name}</h3>
                    <p>👤 ID: {request.user_id} • 🏢 {request.employee_department}</p>
                  </div>
                </div>
                <span className="status-badge">{request.status.toUpperCase()}</span>
              </div>

              <div className="card-body">
                <div className="dates-row">
                  <div className="date-item">
                    <label>Start Date</label>
                    <span>{request.start_date}</span>
                  </div>
                  <div className="date-item">
                    <label>End Date</label>
                    <span>{request.end_date}</span>
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
                    ✅ Approve
                  </button>
                  <button
                    className="btn-reject"
                    onClick={() => {
                      setSelectedRequest(request.id);
                    }}
                  >
                    ❌ Reject
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      ) : !loading ? (
        <div className="empty-state">
          <p>📭 No leave requests found</p>
        </div>
      ) : null}

      {/* Modal for Approval/Rejection */}
      {selectedRequest && (
        <div className="modal-overlay" onClick={() => setSelectedRequest(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Leave Request Decision</h2>
              <button className="modal-close" onClick={() => setSelectedRequest(null)}>
                ✕
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
                {actionInProgress ? 'Processing...' : '❌ Reject'}
              </button>
              <button
                className="btn-modal-approve"
                onClick={() => handleApprove(selectedRequest)}
                disabled={actionInProgress}
              >
                {actionInProgress ? 'Processing...' : '✅ Approve'}
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