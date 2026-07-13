import { useState, useEffect } from 'react';
import { AuthContext } from './auth-context-value';

// Auth Provider Component
export const AuthProvider = ({ children }) => {
  // Read straight from localStorage on first render instead of in an effect -
  // avoids an extra render pass and the "logged out" flash that used to
  // happen while a useEffect caught up on mount.
  const [user, setUser] = useState(() => {
    const savedUser = localStorage.getItem('user');
    return savedUser ? JSON.parse(savedUser) : null;
  });
  const [token, setToken] = useState(() => localStorage.getItem('token'));

  // Keep this tab's auth state in sync when another tab logs in/out or
  // switches accounts. The browser only fires 'storage' in OTHER tabs, which
  // is exactly the case that matters: without this, a tab left open on one
  // account keeps rendering as that account (and its role-gated UI) even
  // after localStorage's token has moved on to a different account, so every
  // request it makes silently carries the new account's token instead.
  useEffect(() => {
    const handleStorageChange = (event) => {
      if (event.key !== 'token' && event.key !== 'user') return;

      const savedToken = localStorage.getItem('token');
      const savedUser = localStorage.getItem('user');

      if (savedToken && savedUser) {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
      } else {
        setToken(null);
        setUser(null);
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  // Login function
  const login = (token, userData) => {
    setToken(token);
    setUser(userData);
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  // Logout function
  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  };

  // Update user profile
  const updateUser = (userData) => {
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const value = {
    user,
    token,
    login,
    logout,
    updateUser,
    isAuthenticated: !!token,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthProvider;
