import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { authAPI } from '../services/api';
import '../styles/Login.css';

export const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await authAPI.login(email, password);
      const { token, user } = response.data;

      // Store token and user data in context and localStorage
      login(token, user);

      // Redirect based on role
      if (user.role === 'hr_manager') {
        navigate('/dashboard/hr');
      } else {
        navigate('/dashboard/employee');
      }
    } catch (err) {
      setError(err.response?.data?.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-logo">S</div>
        <h1>SyncHR</h1>
        <p className="subtitle">Distributed HR Management System</p>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="your@email.com"
              required
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              disabled={loading}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <button type="submit" className="btn-login" disabled={loading}>
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="login-help-links">
          <a
            className="login-help-link"
            href="mailto:hr@example.com?subject=Forgot%20Password&body=Hi%20HR%2C%0A%0AI%20forgot%20my%20SyncHR%20password.%20Could%20you%20please%20reset%20it%20for%20me%3F%0A%0AThanks!"
          >
            Forgot password? <span>Contact HR</span>
          </a>
          <a
            className="login-help-link"
            href="mailto:hr@example.com?subject=New%20Employee%20-%20Account%20Request&body=Hi%20HR%2C%0A%0AI'm%20a%20new%20employee%20and%20need%20my%20SyncHR%20login%20credentials.%0A%0AThanks!"
          >
            New employee? <span>Ask HR</span>
          </a>
        </div>
      </div>
    </div>
  );
};

export default Login;