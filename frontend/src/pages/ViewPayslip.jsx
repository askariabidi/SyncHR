import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { payslipAPI } from '../services/api';
import '../styles/ViewPayslip.css';

export const ViewPayslip = () => {
  const navigate = useNavigate();
  const { user } = useAuth();

  const [payslips, setPayslips] = useState([]);
  const [selectedPayslip, setSelectedPayslip] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchPayslips = async () => {
    try {
      setLoading(true);
      setError('');
      const response = await payslipAPI.getPayslips();
      setPayslips(response.data.data.payslips || []);
    } catch (err) {
      setError('Failed to load payslips');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Fetch payslips on mount
  useEffect(() => {
    fetchPayslips();
  }, []);

  // Get month name
  const getMonthName = (month) => {
    const months = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December'
    ];
    return months[month - 1];
  };

  // Format currency
  const formatCurrency = (amount) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  return (
    <div className="payslip-container">
      {/* Header */}
      <div className="payslip-header">
        <div>
          <h1>Payslips</h1>
          <p>View your monthly salary information</p>
        </div>
        <button className="btn-back" onClick={() => navigate('/dashboard/employee')}>
          Back to Dashboard
        </button>
      </div>

      {/* Error Message */}
      {error && <div className="error-banner">{error}</div>}

      {/* Loading State */}
      {loading && <div className="loading">Loading payslips...</div>}

      {/* Payslips List */}
      {!loading && payslips.length > 0 ? (
        <div className="payslips-grid">
          {payslips.map((payslip) => (
            <div key={payslip.id} className="payslip-card">
              <div className="payslip-month">
                {getMonthName(payslip.month)} {payslip.year}
              </div>
              <div className="payslip-net">
                <div className="label">Net Salary</div>
                <div className="amount">{formatCurrency(payslip.net_salary)}</div>
              </div>
              <div className="payslip-details">
                <div className="detail-row">
                  <span>Basic:</span>
                  <span>{formatCurrency(payslip.basic_salary)}</span>
                </div>
                <div className="detail-row">
                  <span>Allowances:</span>
                  <span className="positive">+{formatCurrency(payslip.allowances)}</span>
                </div>
                <div className="detail-row">
                  <span>Bonus:</span>
                  <span className="positive">+{formatCurrency(payslip.bonus)}</span>
                </div>
                <div className="detail-row">
                  <span>Tax:</span>
                  <span className="negative">-{formatCurrency(payslip.tax)}</span>
                </div>
                <div className="detail-row">
                  <span>Deductions:</span>
                  <span className="negative">-{formatCurrency(payslip.deductions)}</span>
                </div>
              </div>
              <button
                className="btn-view-details"
                onClick={() => setSelectedPayslip(payslip)}
              >
                View Full Details
              </button>
            </div>
          ))}
        </div>
      ) : !loading ? (
        <div className="empty-state">
          <p>No payslips available</p>
        </div>
      ) : null}

      {/* Payslip Details Modal */}
      {selectedPayslip && (
        <div className="modal-overlay" onClick={() => setSelectedPayslip(null)}>
          <div className="modal payslip-modal" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setSelectedPayslip(null)}>&times;</button>

            {/* Payslip Header */}
            <div className="payslip-document-header">
              <h2>PAYSLIP</h2>
              <p>{getMonthName(selectedPayslip.month)} {selectedPayslip.year}</p>
            </div>

            {/* Employee Info */}
            <div className="payslip-section">
              <h3>Employee Information</h3>
              <div className="info-grid">
                <div className="info-item">
                  <label>Name:</label>
                  <span>{user?.first_name} {user?.last_name}</span>
                </div>
                <div className="info-item">
                  <label>Email:</label>
                  <span>{user?.email}</span>
                </div>
                <div className="info-item">
                  <label>Department:</label>
                  <span>{user?.department}</span>
                </div>
                <div className="info-item">
                  <label>Month/Year:</label>
                  <span>{getMonthName(selectedPayslip.month)} {selectedPayslip.year}</span>
                </div>
              </div>
            </div>

            {/* Earnings */}
            <div className="payslip-section">
              <h3>Earnings</h3>
              <div className="earnings-table">
                <div className="table-row">
                  <span>Basic Salary</span>
                  <span>{formatCurrency(selectedPayslip.basic_salary)}</span>
                </div>
                <div className="table-row">
                  <span>Allowances</span>
                  <span className="positive">{formatCurrency(selectedPayslip.allowances)}</span>
                </div>
                <div className="table-row">
                  <span>Bonus</span>
                  <span className="positive">{formatCurrency(selectedPayslip.bonus)}</span>
                </div>
                <div className="table-row total">
                  <span>Total Earnings</span>
                  <span>{formatCurrency(selectedPayslip.basic_salary + selectedPayslip.allowances + selectedPayslip.bonus)}</span>
                </div>
              </div>
            </div>

            {/* Deductions */}
            <div className="payslip-section">
              <h3>Deductions</h3>
              <div className="deductions-table">
                <div className="table-row">
                  <span>Tax</span>
                  <span className="negative">{formatCurrency(selectedPayslip.tax)}</span>
                </div>
                <div className="table-row">
                  <span>Other Deductions</span>
                  <span className="negative">{formatCurrency(selectedPayslip.deductions)}</span>
                </div>
                <div className="table-row total">
                  <span>Total Deductions</span>
                  <span className="negative">{formatCurrency(selectedPayslip.tax + selectedPayslip.deductions)}</span>
                </div>
              </div>
            </div>

            {/* Attendance */}
            <div className="payslip-section">
              <h3>Attendance</h3>
              <div className="attendance-grid">
                <div className="attendance-item">
                  <label>Working Days</label>
                  <span className="big-text">{selectedPayslip.working_days}</span>
                </div>
                <div className="attendance-item">
                  <label>Leave Taken</label>
                  <span className="big-text">{selectedPayslip.leave_taken}</span>
                </div>
              </div>
            </div>

            {/* Net Salary */}
            <div className="payslip-section net-salary-section">
              <h3>NET SALARY</h3>
              <div className="net-salary-amount">
                {formatCurrency(selectedPayslip.net_salary)}
              </div>
            </div>

            {/* Footer */}
            <div className="payslip-footer">
              <p>This is an electronically generated payslip and requires no signature.</p>
            </div>

            <button className="btn-close-modal" onClick={() => setSelectedPayslip(null)}>
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default ViewPayslip;