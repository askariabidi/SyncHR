// import axios from 'axios';

// // API Base URL - pointing to Go backend
// const API_BASE_URL = 'http://localhost:8080/api';

// // Create axios instance with base URL
// const apiClient = axios.create({
//   baseURL: API_BASE_URL,
//   headers: {
//     'Content-Type': 'application/json',
//   },
// });

// // Add JWT token to every request
// apiClient.interceptors.request.use((config) => {
//   const token = localStorage.getItem('token');
//   if (token) {
//     config.headers.Authorization = `Bearer ${token}`;
//   }
//   return config;
// });

// // Handle errors globally
// apiClient.interceptors.response.use(
//   (response) => response,
//   (error) => {
//     if (error.response?.status === 401) {
//       // Token expired or invalid
//       localStorage.removeItem('token');
//       localStorage.removeItem('user');
//       window.location.href = '/login';
//     }
//     return Promise.reject(error);
//   }
// );

// // ============================================================================
// // AUTHENTICATION ENDPOINTS
// // ============================================================================

// export const authAPI = {
//   login: (email, password) =>
//     apiClient.post('/auth/login', { email, password }),
  
//   register: (userData) =>
//     apiClient.post('/auth/register', userData),
  
//   getProfile: () =>
//     apiClient.get('/users/profile'),
  
//   updateProfile: (userData) =>
//     apiClient.put('/users/profile', userData),
// };

// // ============================================================================
// // ATTENDANCE ENDPOINTS
// // ============================================================================

// export const attendanceAPI = {
//   checkIn: (timestamp) =>
//     apiClient.post('/attendance/checkin', { timestamp }),
  
//   checkOut: (timestamp) =>
//     apiClient.post('/attendance/checkout', { timestamp }),
  
//   getHistory: (month, year) =>
//     apiClient.get(`/attendance/history?month=${month}&year=${year}`),
// };

// // ============================================================================
// // LEAVE ENDPOINTS
// // ============================================================================

// export const leaveAPI = {
//   applyLeave: (leaveData) =>
//     apiClient.post('/leave/apply', leaveData),
  
//   getBalance: () =>
//     apiClient.get('/leave/balance'),
  
//   getRequests: () =>
//     apiClient.get('/leave/requests'),
  
//   approveLeave: (leaveId, notes) =>
//     apiClient.put(`/leave/approve/${leaveId}`, { approval_notes: notes }),
  
//   rejectLeave: (leaveId, notes) =>
//     apiClient.put(`/leave/reject/${leaveId}`, { approval_notes: notes }),
// };

// // ============================================================================
// // PAYSLIP ENDPOINTS
// // ============================================================================

// export const payslipAPI = {
//   getPayslips: () =>
//     apiClient.get('/payslip/list'),
  
//   getPayslipDetails: (payslipId) =>
//     apiClient.get(`/payslip/${payslipId}`),
// };

// // ============================================================================
// // HOLIDAY ENDPOINTS
// // ============================================================================

// export const holidayAPI = {
//   getHolidays: () =>
//     apiClient.get('/holidays'),
// };

// // ============================================================================
// // HEALTH CHECK
// // ============================================================================

// export const healthCheck = () =>
//   apiClient.get('/health');

// export default apiClient;

import axios from 'axios';

// API Base URL - pointing to Go backend
const API_BASE_URL = 'http://localhost:8080/api';

// Create axios instance with base URL
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add JWT token to every request
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle errors globally
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// ============================================================================
// AUTHENTICATION ENDPOINTS
// ============================================================================

export const authAPI = {
  login: (email, password) =>
    apiClient.post('/auth/login', { email, password }),
  
  register: (userData) =>
    apiClient.post('/auth/register', userData),
  
  getProfile: () =>
    apiClient.get('/users/profile'),
  
  updateProfile: (userData) =>
    apiClient.put('/users/profile', userData),
};

// ============================================================================
// ATTENDANCE ENDPOINTS
// ============================================================================

export const attendanceAPI = {
  checkIn: () =>
    apiClient.post('/attendance/checkin', {}),
  
  checkOut: () =>
    apiClient.post('/attendance/checkout', {}),
  
  getAttendanceHistory: () =>
    apiClient.get('/attendance/history'),
  
  getHistory: (month, year) =>
    apiClient.get(`/attendance/history?month=${month}&year=${year}`),
};

// ============================================================================
// LEAVE ENDPOINTS
// ============================================================================

export const leaveAPI = {
  applyLeave: (leaveData) =>
    apiClient.post('/leave/apply', leaveData),
  
  getBalance: () =>
    apiClient.get('/leave/balance'),
  
  getRequests: () =>
    apiClient.get('/leave/requests'),
  
  approveLeave: (leaveId, notes) =>
    apiClient.put(`/leave/approve/${leaveId}`, { approval_notes: notes }),
  
  rejectLeave: (leaveId, notes) =>
    apiClient.put(`/leave/reject/${leaveId}`, { approval_notes: notes }),
};

// ============================================================================
// PAYSLIP ENDPOINTS
// ============================================================================

export const payslipAPI = {
  getPayslips: () =>
    apiClient.get('/payslip/list'),
  
  getPayslipDetails: (payslipId) =>
    apiClient.get(`/payslip/${payslipId}`),
};

// ============================================================================
// HOLIDAY ENDPOINTS
// ============================================================================

export const holidayAPI = {
  getHolidays: () =>
    apiClient.get('/holidays'),
};

// ============================================================================
// HEALTH CHECK
// ============================================================================

export const healthCheck = () =>
  apiClient.get('/health');

export default apiClient;