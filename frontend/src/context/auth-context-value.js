import { createContext } from 'react';

// Split into its own file (rather than living in AuthContext.jsx) so that
// file can export the AuthProvider component only - keeps Fast Refresh happy.
// Named distinctly (not just a case difference from AuthContext.jsx) since
// Windows/Mac filesystems are case-insensitive and would otherwise treat the
// two as the same file.
export const AuthContext = createContext();
