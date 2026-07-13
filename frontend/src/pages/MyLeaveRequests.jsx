import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { leaveAPI } from '../services/api';
import '../styles/MyLeaveRequests.css';

export const MyLeaveRequests = () => {
  const navigate = useNavigate();

  const [leaveRequests, setLeaveRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('all'); // all, pending, approved, rejected

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

  // Fetch leave requests on mount
  useEffect(() => {
    fetchLeaveRequests();
  }, []);

  // Filter requests based on status
  const filteredRequests = leaveRequests.filter((request) => {
    if (filter === 'all') return true;
    return request.status === filter;
  });

  // Get status badge color
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

  return (
    <div className="my-leave-requests-container">
      {/* Header */}
      <div className="requests-header">
        <div>
          <h1>My Leave Requests</h1>
          <p>Track all your leave applications and their status</p>
        </div>
        <button className="btn-back" onClick={() => navigate('/dashboard/employee')}>
          Back to Dashboard
        </button>
      </div>

      {/* Filter Buttons */}
      <div className="filter-buttons">
        <button
          className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
          onClick={() => setFilter('all')}
        >
          All Requests
        </button>
        <button
          className={`filter-btn ${filter === 'pending' ? 'active' : ''}`}
          onClick={() => setFilter('pending')}
        >
          Pending
        </button>
        <button
          className={`filter-btn ${filter === 'approved' ? 'active' : ''}`}
          onClick={() => setFilter('approved')}
        >
          Approved
        </button>
        <button
          className={`filter-btn ${filter === 'rejected' ? 'active' : ''}`}
          onClick={() => setFilter('rejected')}
        >
          Rejected
        </button>
      </div>

      {/* Error Message */}
      {error && <div className="error-banner">{error}</div>}

      {/* Loading State */}
      {loading && (
        <div className="loading-container">
          <p>Loading your leave requests...</p>
        </div>
      )}

      {/* Requests Table */}
      {!loading && filteredRequests.length > 0 ? (
        <div className="requests-table-container">
          <table className="requests-table">
            <thead>
              <tr>
                <th>Leave Type</th>
                <th>Start Date</th>
                <th>End Date</th>
                <th>Days</th>
                <th>Status</th>
                <th>Reason</th>
                <th>HR Comments</th>
                <th>Submitted</th>
              </tr>
            </thead>
            <tbody>
              {filteredRequests.map((request) => (
                <tr key={request.id}>
                  <td className="leave-type">{request.leave_type_name}</td>
                  <td>{request.start_date}</td>
                  <td>{request.end_date}</td>
                  <td className="days-cell">
                    <strong>{request.number_of_days}</strong>
                  </td>
                  <td>
                    <span className={`status-badge ${getStatusBadge(request.status)}`}>
                      {request.status.charAt(0).toUpperCase() + request.status.slice(1)}
                    </span>
                  </td>
                  <td className="reason-cell">{request.reason || '-'}</td>
                  <td className="comments-cell">{request.approval_notes || '-'}</td>
                  <td className="date-cell">
                    {new Date(request.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : !loading ? (
        <div className="empty-state">
          <p>No leave requests found</p>
          <button className="btn-apply" onClick={() => navigate('/dashboard/employee/leave')}>
            Apply for Leave
          </button>
        </div>
      ) : null}

      {/* Summary Stats */}
      {!loading && leaveRequests.length > 0 && (
        <div className="stats-section">
          <div className="stat-card">
            <div className="stat-number">{leaveRequests.filter(r => r.status === 'pending').length}</div>
            <div className="stat-label">Pending</div>
          </div>
          <div className="stat-card approved">
            <div className="stat-number">{leaveRequests.filter(r => r.status === 'approved').length}</div>
            <div className="stat-label">Approved</div>
          </div>
          <div className="stat-card rejected">
            <div className="stat-number">{leaveRequests.filter(r => r.status === 'rejected').length}</div>
            <div className="stat-label">Rejected</div>
          </div>
          <div className="stat-card total">
            <div className="stat-number">{leaveRequests.length}</div>
            <div className="stat-label">Total</div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MyLeaveRequests;