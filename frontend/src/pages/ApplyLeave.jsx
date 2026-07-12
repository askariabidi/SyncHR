import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { leaveAPI } from '../services/api';
import '../styles/ApplyLeave.css';

export const ApplyLeave = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  
  const [formData, setFormData] = useState({
    leaveTypeId: '',
    startDate: '',
    endDate: '',
    numberOfDays: 0,
    reason: '',
  });

  const [leaveTypes, setLeaveTypes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Fetch leave types on mount
  useEffect(() => {
    // For now, using hardcoded leave types
    setLeaveTypes([
      { id: 1, name: 'Sick Leave' },
      { id: 2, name: 'Casual Leave' },
      { id: 3, name: 'Earned Leave' },
    ]);
  }, []);

  // Calculate number of days between dates
  const calculateDays = (start, end) => {
    if (start && end) {
      const startDate = new Date(start);
      const endDate = new Date(end);
      const days = Math.ceil((endDate - startDate) / (1000 * 60 * 60 * 24)) + 1;
      return Math.max(0, days);
    }
    return 0;
  };

  // Handle form input changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => {
      const updated = { ...prev, [name]: value };

      // Auto-calculate days if dates change
      if (name === 'startDate' || name === 'endDate') {
        updated.numberOfDays = calculateDays(updated.startDate, updated.endDate);
      }

      return updated;
    });
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    // Validate form
    if (!formData.leaveTypeId || !formData.startDate || !formData.endDate) {
      setError('Please fill in all required fields');
      setLoading(false);
      return;
    }

    if (formData.numberOfDays <= 0) {
      setError('End date must be after start date');
      setLoading(false);
      return;
    }

    try {
      await leaveAPI.applyLeave({
        leave_type_id: parseInt(formData.leaveTypeId),
        start_date: formData.startDate,
        end_date: formData.endDate,
        number_of_days: formData.numberOfDays,
        reason: formData.reason,
      });

      setSuccess('✅ Leave request submitted successfully!');
      
      // Reset form
      setFormData({
        leaveTypeId: '',
        startDate: '',
        endDate: '',
        numberOfDays: 0,
        reason: '',
      });

      // Redirect after 2 seconds
      setTimeout(() => {
        navigate('/dashboard/employee');
      }, 2000);
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to apply leave. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="apply-leave-container">
      {/* Header */}
      <div className="leave-header">
        <div>
          <h1>📋 Apply for Leave</h1>
          <p>Submit your leave request for approval</p>
        </div>
        <button className="btn-back" onClick={() => navigate('/dashboard/employee')}>
          ← Back to Dashboard
        </button>
      </div>

      {/* Form Card */}
      <div className="leave-form-card">
        {success && <div className="success-message">{success}</div>}
        {error && <div className="error-message">{error}</div>}

        <form onSubmit={handleSubmit}>
          {/* Leave Type */}
          <div className="form-group">
            <label htmlFor="leaveTypeId">Leave Type *</label>
            <select
              id="leaveTypeId"
              name="leaveTypeId"
              value={formData.leaveTypeId}
              onChange={handleChange}
              required
              disabled={loading}
            >
              <option value="">Select leave type</option>
              {leaveTypes.map((type) => (
                <option key={type.id} value={type.id}>
                  {type.name}
                </option>
              ))}
            </select>
          </div>

          {/* Start Date */}
          <div className="form-group">
            <label htmlFor="startDate">Start Date *</label>
            <input
              id="startDate"
              type="date"
              name="startDate"
              value={formData.startDate}
              onChange={handleChange}
              required
              disabled={loading}
            />
          </div>

          {/* End Date */}
          <div className="form-group">
            <label htmlFor="endDate">End Date *</label>
            <input
              id="endDate"
              type="date"
              name="endDate"
              value={formData.endDate}
              onChange={handleChange}
              required
              disabled={loading}
            />
          </div>

          {/* Number of Days (Auto-calculated) */}
          <div className="form-group">
            <label>Number of Days</label>
            <input
              type="number"
              value={formData.numberOfDays}
              disabled
              className="readonly"
            />
            <small>Automatically calculated</small>
          </div>

          {/* Reason */}
          <div className="form-group full-width">
            <label htmlFor="reason">Reason for Leave</label>
            <textarea
              id="reason"
              name="reason"
              value={formData.reason}
              onChange={handleChange}
              placeholder="Enter reason for your leave request (optional)"
              rows="4"
              disabled={loading}
            />
          </div>

          {/* Submit Button */}
          <button type="submit" className="btn-submit" disabled={loading}>
            {loading ? 'Submitting...' : '✉️ Submit Leave Request'}
          </button>
        </form>
      </div>

      {/* Info Section */}
      <div className="leave-info">
        <h3>📌 Important Information</h3>
        <ul>
          <li>Leave requests must be submitted at least 2 days in advance</li>
          <li>Your HR manager will review and approve/reject your request</li>
          <li>You will receive a notification once your request is processed</li>
          <li>Check your leave balance before applying</li>
        </ul>
      </div>
    </div>
  );
};

export default ApplyLeave;