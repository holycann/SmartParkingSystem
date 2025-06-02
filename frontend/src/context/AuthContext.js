import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { 
  loginUser, 
  logoutUser, 
  getUserProfile, 
  updateUserProfile as apiUpdateProfile,
  validateToken,
  registerUser
} from '../services/apiService';
import { toast } from 'react-toastify';
import { jwtDecode } from 'jwt-decode';
import { useNavigate } from 'react-router-dom';

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

// Secure storage helper for sensitive data
const secureStorage = {
  setItem: (key, value) => {
    try {
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

// Session timeout configuration
const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tokenRefreshTimer, setTokenRefreshTimer] = useState(null);
  const [sessionTimer, setSessionTimer] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [authToken, setAuthToken] = useState(null);
  const navigate = useNavigate();

  // Clear all auth data
  const clearAuthData = useCallback(() => {
    localStorage.removeItem('token');
    setAuthToken(null);
    secureStorage.removeItem('userData');
    
    if (tokenRefreshTimer) {
      clearTimeout(tokenRefreshTimer);
      setTokenRefreshTimer(null);
    }
    
    if (sessionTimer) {
      clearTimeout(sessionTimer);
      setSessionTimer(null);
    }
    
    setCurrentUser(null);
    setIsAuthenticated(false);
  }, [tokenRefreshTimer, sessionTimer]);

  // Reset session timer
  const resetSessionTimer = useCallback(() => {
    if (sessionTimer) {
      clearTimeout(sessionTimer);
    }
    
    const timer = setTimeout(() => {
      toast.warn('Your session is about to expire due to inactivity');
      
      // Give the user 1 minute warning before logout
      setTimeout(() => {
        clearAuthData();
        toast.error('You have been logged out due to inactivity');
        //navigate('/login');
      }, 60000);
    }, SESSION_TIMEOUT);
    
    setSessionTimer(timer);
  }, [sessionTimer, clearAuthData, navigate]);

  // Enhanced token validation with retries
  const checkTokenValidity = useCallback(async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      console.log('No token found');
      clearAuthData();
      return false;
    }

    try {
      // Check token expiration
      const decodedToken = jwtDecode(token);
      const currentTime = Date.now() / 1000;
      
      if (decodedToken.exp < currentTime) {
        console.log('Token expired, logging out');
        clearAuthData();
        toast.error('Your session has expired. Please log in again.');
        //navigate('/login');
        return false;
      }

      // Validate token with backend
      const isValid = await validateToken(token);
      if (!isValid) {
        console.log('Invalid token, logging out');
        clearAuthData();
        toast.error('Authentication invalid. Please log in again.');
        //navigate('/login');
        return false;
      }

      setIsAuthenticated(true);
      setAuthToken(token);
      return true;
    } catch (error) {
      console.error('Token validation error:', error);
      clearAuthData();
      //navigate('/login');
      return false;
    }
  }, [clearAuthData, navigate]);

  // Initialize auth state
  useEffect(() => {
    const initializeAuth = async () => {
      setLoading(true);
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          clearAuthData();
          setLoading(false);
          return;
        }

        const isValid = await checkTokenValidity();
        if (isValid) {
          const userData = secureStorage.getItem('userData');
          if (userData) {
            setCurrentUser(userData);
            setIsAuthenticated(true);
            setAuthToken(token);
          } else {
            // If we have a valid token but no user data, fetch it
            try {
              const profile = await getUserProfile();
              setCurrentUser(profile);
              secureStorage.setItem('userData', profile);
              setIsAuthenticated(true);
              setAuthToken(token);
              resetSessionTimer();
            } catch (error) {
              console.error('Error fetching user profile:', error);
              clearAuthData();
            }
          }
        }
      } catch (error) {
        console.error('Error initializing auth:', error);
        clearAuthData();
        toast.error('Failed to initialize authentication');
      } finally {
        setLoading(false);
      }
    };

    initializeAuth();
  }, [checkTokenValidity, clearAuthData]);

  // Login function
  const login = async (credentials) => {
    try {
      setLoading(true);
      setError(null);
      
      const { token, user } = await loginUser(credentials);
      localStorage.setItem('token', token);
      secureStorage.setItem('userData', user);
      setCurrentUser(user);
      setIsAuthenticated(true);
      setAuthToken(token);
      resetSessionTimer();
      toast.success('Login successful!');
      navigate('/');
    } catch (err) {
      setError(err.message || 'An error occurred during login');
      toast.error(err.message || 'Login failed. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Logout function
  const logout = async () => {
    try {
      setLoading(true);
      await logoutUser();
    } catch (error) {
      console.error('Error during logout:', error);
    } finally {
      clearAuthData();
      toast.info('You have been logged out');
      //navigate('/login');
      setLoading(false);
    }
  };

  // Update user profile
  const updateProfile = async (profileData) => {
    try {
      const updatedProfile = await apiUpdateProfile(profileData);
      setCurrentUser(updatedProfile);
      return updatedProfile;
    } catch (error) {
      throw error;
    }
  };

  // Register function
  const register = async (userData) => {
    try {
      setLoading(true);
      setError(null);
      
      await registerUser(userData);
      toast.success('Registration successful! Please log in.');
      return true;
    } catch (err) {
      setError(err.message || 'An error occurred during registration');
      toast.error(err.message || 'Registration failed. Please try again.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const value = {
    currentUser,
    loading,
    error,
    isAuthenticated,
    login,
    logout,
    updateProfile,
    register,
    token: authToken,
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};

export default AuthContext;
