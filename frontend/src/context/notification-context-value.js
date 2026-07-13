import { createContext } from 'react';

// Split out for the same reason as auth-context-value.js - keeps
// NotificationContext.jsx down to a single component export so Fast Refresh
// can track it properly.
export const NotificationContext = createContext();
