import { useState, useEffect, useCallback } from 'react';
import { loginUser, validateToken, getUserProfile } from '../services/apiService';
import {jwtDecode} from 'jwt-decode';
import { useNavigate } from 'react-router-dom';

// Secure storage helper functions
const secureStorage = {
  setItem: (key, value) => {
    try {
      // For sensitive data, consider using more secure options in production
      // This is a basic implementation that adds some protection
      const encryptedValue = btoa(JSON.stringify(value));
      localStorage.setItem(key, encryptedValue);
    } catch (error) {
      console.error('Error storing data securely:', error);
    }
  },
  getItem: (key) => {
    try {
      const item = localStorage.getItem(key);
      if (!item) return null;
      return JSON.parse(atob(item));
    } catch (error) {
      console.error('Error retrieving secure data:', error);
      return null;
    }
  },
  removeItem: (key) => {
    localStorage.removeItem(key);
  }
};

export const useAuth = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tokenRefreshTimer, setTokenRefreshTimer] = useState(null);
  const navigate = useNavigate();

  // Clear all auth data
  const clearAuthData = useCallback(() => {
    localStorage.removeItem('token');
    secureStorage.removeItem('userData');
    setUser(null);
    if (tokenRefreshTimer) {
      clearTimeout(tokenRefreshTimer);
      setTokenRefreshTimer(null);
    }
  }, [tokenRefreshTimer]);

  // Check token validity and expiration
  const checkTokenValidity = useCallback(async () => {
    const token = localStorage.getItem('token');
    if (!token) return false;

    try {
      // Check token expiration
      const decodedToken = jwtDecode(token);
      const currentTime = Date.now() / 1000;
      
      if (decodedToken.exp < currentTime) {
        console.log('Token expired, logging out');
        clearAuthData();
        return false;
      }

      // Validate token with backend
      const isValid = await validateToken(token);
      if (!isValid) {
        console.log('Invalid token, logging out');
        clearAuthData();
        return false;
      }

      return true;
    } catch (error) {
      console.error('Token validation error:', error);
      clearAuthData();
      return false;
    }
  }, [clearAuthData]);

  // Setup token refresh
  const setupTokenRefresh = useCallback((token) => {
    try {
      const decodedToken = jwtDecode(token);
      const expiryTime = decodedToken.exp * 1000; // Convert to milliseconds
      const currentTime = Date.now();
      const timeUntilExpiry = expiryTime - currentTime;
      
      // Refresh 5 minutes before expiry
      const refreshTime = timeUntilExpiry - (5 * 60 * 1000);
      
      if (refreshTime > 0) {
        const timerId = setTimeout(async () => {
          console.log('Refreshing token');
          try {
            const response = await fetch('/api/refresh-token', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
              }
            });
            
            if (response.ok) {
              const data = await response.json();
              localStorage.setItem('token', data.token);
              setupTokenRefresh(data.token);
            } else {
              clearAuthData();
              setError('Session expired. Please log in again.');
              //navigate('/login');
            }
          } catch (error) {
            console.error('Token refresh error:', error);
            clearAuthData();
            setError('Session expired. Please log in again.');
            //navigate('/login');
          }
        }, refreshTime);
        
        setTokenRefreshTimer(timerId);
      }
    } catch (error) {
      console.error('Error setting up token refresh:', error);
    }
  }, [clearAuthData, navigate]);

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('token');
      if (token) {
        const isValid = await checkTokenValidity();
        
        if (isValid) {
          try {
            // Fetch actual user data from backend
            const userData = await getUserProfile();
            setUser(userData);
            secureStorage.setItem('userData', userData);
            setupTokenRefresh(token);
          } catch (err) {
            console.error('Error fetching user data:', err);
            clearAuthData();
          }
        }
      }
      setLoading(false);
    };

    initAuth();

    // Cleanup function
    return () => {
      if (tokenRefreshTimer) {
        clearTimeout(tokenRefreshTimer);
      }
    };
  }, [checkTokenValidity, clearAuthData, setupTokenRefresh, tokenRefreshTimer]);

  const login = async (credentials, rememberMe = false) => {
    try {
      setLoading(true);
      setError(null);
      
      const data = await loginUser(credentials);
      
      // Store token in localStorage
      localStorage.setItem('token', data.token);
      
      // Store user data securely if remember me is enabled
      if (rememberMe) {
        secureStorage.setItem('userData', data.user);
      }
      
      setUser(data.user);
      setupTokenRefresh(data.token);
      
      return data.user;
    } catch (err) {
      // Handle specific error types
      if (err.response) {
        switch (err.response.status) {
          case 401:
            setError('Invalid email or password');
            break;
          case 403:
            setError('Your account has been locked. Please contact support.');
            break;
          case 429:
            setError('Too many login attempts. Please try again later.');
            break;
          default:
            setError(err.response.data?.message || 'An error occurred during login');
        }
      } else if (err.request) {
        setError('Unable to connect to the server. Please check your internet connection.');
      } else {
        setError('An unexpected error occurred');
      }
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      // Call backend logout endpoint to invalidate token
      const token = localStorage.getItem('token');
      if (token) {
        await fetch('/api/users/logout', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear all user-related data from storage
      clearAuthData();
    }
  };

  const resetPassword = async (email) => {
    try {
      setLoading(true);
      const response = await fetch('/api/users/reset-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email })
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Password reset failed');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message || 'Password reset failed');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const verifyEmail = async (token) => {
    try {
      setLoading(true);
      const response = await fetch(`/api/users/verify-email/${token}`, {
        method: 'GET'
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Email verification failed');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message || 'Email verification failed');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const setupMFA = async (method) => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token');
      
      const response = await fetch('/api/users/setup-mfa', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ method })
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'MFA setup failed');
      }
      
      return await response.json();
    } catch (err) {
      setError(err.message || 'MFA setup failed');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { 
    user, 
    loading, 
    error, 
    login, 
    logout,
    resetPassword,
    verifyEmail,
    setupMFA,
    isAuthenticated: !!user
  };
};
