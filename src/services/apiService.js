import axios from 'axios';

import {
  registerHandlerByPrefix,
  isWebSocketConnected,
} from './webSocketService';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api';

// Create an Axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 15000, // 15 seconds timeout
});

// Request cache implementation
const requestCache = {
  cache: new Map(),
  maxAge: 5 * 60 * 1000, // 5 minutes default cache time

  get: function (key) {
    const cachedItem = this.cache.get(key);
    if (!cachedItem) return null;

    const now = Date.now();
    if (now > cachedItem.expiry) {
      this.cache.delete(key);
      return null;
    }

    return cachedItem.data;
  },

  set: function (key, data, maxAge = this.maxAge) {
    const expiry = Date.now() + maxAge;
    this.cache.set(key, { data, expiry });
  },

  invalidate: function (keyPattern) {
    if (typeof keyPattern === 'string') {
      // Delete specific key
      this.cache.delete(keyPattern);
    } else if (keyPattern instanceof RegExp) {
      // Delete keys matching regex pattern
      for (const key of this.cache.keys()) {
        if (keyPattern.test(key)) {
          this.cache.delete(key);
        }
      }
    }
  },

  clear: function () {
    this.cache.clear();
  }
};

// Add a request interceptor to include the token in headers
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // Check cache for GET requests
  if (config.method === 'get' && config.cache !== false) {
    const cacheKey = `${config.url}${JSON.stringify(config.params || {})}`;
    const cachedData = requestCache.get(cacheKey);

    if (cachedData) {
      // Return cached data
      config.adapter = () => {
        return Promise.resolve({
          data: cachedData,
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
          request: {}
        });
      };
    } else {
      // Store the cache key for response interceptor
      config.cacheKey = cacheKey;
    }
  }

  return config;
}, (error) => {
  return Promise.reject(error);
});

// Add response interceptor to handle common error scenarios and caching
api.interceptors.response.use(
  (response) => {
    // Cache successful GET responses if applicable
    if (response.config.method === 'get' && response.config.cacheKey && response.status === 200) {
      const maxAge = response.config.cacheMaxAge || requestCache.maxAge;
      requestCache.set(response.config.cacheKey, response.data, maxAge);
    }

    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // Handle token expiration
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      // Clear token and redirect to login
      localStorage.removeItem('token');
      //requestCache.clear();
      //window.location.href = '/login';
      return Promise.reject(new Error('Session expired. Please log in again.'));
    }

    // Handle rate limiting
    if (error.response?.status === 429) {
      const retryAfter = error.response.headers['retry-after'] || 60;
      console.warn(`Rate limited. Retry after ${retryAfter} seconds.`);
    }

    // Handle server errors
    if (error.response?.status >= 500) {
      console.error('Server error:', error.response.data);
    }

    return Promise.reject(error);
  }
);

// Helper function for making cached requests
const cachedRequest = async (url, params = {}, maxAge = null) => {
  return api.get(url, {
    params,
    cache: true,
    cacheMaxAge: maxAge
  });
};

// API Functions

// Authentication
export const validateToken = async (token) => {
  try {
    const response = await api.post('/users/validate-token', { token });
    return response.data.valid;
  } catch (error) {
    console.error('Error validating token:', error);
    return false;
  }
};

export const loginUser = async (credentials) => {
  try {
    const response = await api.post('/users/login', credentials);
    const { token, user } = response.data;

    // Store the token
    localStorage.setItem('token', token);

    // Clear cache on login
    requestCache.clear();

    return { token, user };
  } catch (error) {
    console.error('Error logging in:', error);

    if (error.response) {
      switch (error.response.status) {
        case 401:
          throw new Error('Invalid email or password');
        case 403:
          throw new Error('Your account has been locked');
        case 429:
          throw new Error('Too many login attempts. Please try again later');
        default:
          throw new Error(error.response.data?.message || 'Login failed');
      }
    }

    throw new Error('Network error. Please check your connection');
  }
};

export const registerUser = async (userData) => {
  try {
    const response = await api.post('/users/register', userData);
    return response.data;
  } catch (error) {
    console.error('Error registering user:', error);

    if (error.response) {
      switch (error.response.status) {
        case 409:
          throw new Error('Email already in use');
        case 400:
          throw new Error(error.response.data?.message || 'Invalid registration data');
        case 422:
          throw new Error(error.response.data?.message || 'Validation failed');
        case 429:
          throw new Error('Too many registration attempts. Please try again later');
        case 500:
          throw new Error('Server error. Please try again later');
        default:
          throw new Error(error.response.data?.message || 'Registration failed');
      }
    }

    throw new Error('Network error. Please check your connection');
  }
};

export const logoutUser = async () => {
  try {
    await api.post('/users/logout');
  } catch (error) {
    console.error('Error logging out:', error);
  } finally {
    // Always clear local storage and cache
    localStorage.removeItem('token');
    requestCache.clear();
  }
};

// User Profile
export const getUserProfile = async () => {
  try {
    const response = await api.get('/users/profile');
    return response.data.user;
  } catch (error) {
    console.error('Error fetching user profile:', error);
    throw new Error(error.response?.data?.message || 'Failed to fetch user profile');
  }
};

export const updateUserProfile = async (profileData) => {
  try {
    const response = await api.put('/users/profile', profileData);
    // Invalidate user-related cache entries
    requestCache.invalidate(/\/users\//);
    return response.data;
  } catch (error) {
    console.error('Error updating profile:', error);
    throw new Error(error.response?.data?.message || 'Failed to update profile');
  }
};

// Parking
export const fetchParkingLots = async (params = {}) => {
  try {
    const response = await cachedRequest('/parking-lots', params, 5 * 60 * 1000); // Cache for 5 minutes
    return response.data;
  } catch (error) {
    console.error('Error fetching parking lots:', error);
    throw error;
  }
};

export const fetchParkingSpaces = async (lotId) => {
  // If WebSocket is not connected, just use REST API
  try {
    const response = await api.get(`/parking-lots/${lotId}/space`);
    return response.data;
  } catch (error) {
    console.error('Error fetching parking spaces:', error);
    throw error;
  }
};

export const fetchOccupancyStats = async (lotId) => {
  // First try to get data via WebSocket if connected
  if (isWebSocketConnected()) {
    // Return a promise that will be resolved when we get data via WebSocket
    // But also have a fallback to REST API if WebSocket doesn't deliver data quickly
    return new Promise((resolve, reject) => {
      let resolved = false;

      // Set a timeout to fall back to REST API if WebSocket doesn't deliver
      const timeoutId = setTimeout(async () => {
        if (!resolved) {
          try {
            // Fall back to REST API
            const response = await api.get(`/parking-lots/${lotId}/space/occupancy`);
            resolved = true;
            resolve(response.data);
          } catch (error) {
            console.error('Error fetching occupancy stats via REST API:', error);
            reject(error);
          }
        }
      }, 2000); // Wait 2 seconds before falling back to REST API

      // Register a one-time handler for occupancy stats updates
      const unregister = registerHandlerByPrefix('occupancy_update', (data) => {
        if (data && data.parkingLotId === lotId) {
          clearTimeout(timeoutId);
          unregister(); // Remove this handler
          if (!resolved) {
            resolved = true;
            resolve({
              currentOccupancy: data.currentOccupancy || 0,
              totalSpaces: data.totalSpaces || 0,
              occupiedSpaces: data.occupiedSpaces || 0
            });
          }
        }
      });

      // Also make the REST API call immediately to ensure we get data
      api.get(`/parking-lots/${lotId}/space/occupancy`)
        .then(response => {
          if (!resolved) {
            resolved = true;
            clearTimeout(timeoutId);
            unregister(); // Remove the WebSocket handler
            resolve(response.data);
          }
        })
        .catch(error => {
          if (!resolved) {
            console.error('Error fetching occupancy stats:', error);
            // Don't reject yet, wait for WebSocket or timeout
          }
        });
    });
  } else {
    // If WebSocket is not connected, just use REST API
    try {
      const response = await api.get(`/parking-lots/${lotId}/space/occupancy`);
      return response.data;
    } catch (error) {
      console.error('Error fetching occupancy stats:', error);
      throw error;
    }
  }
};

// Reservations
export const fetchReservationsById = async (id) => {
  try {
    const response = await api.get(`/reservations/details/${id}`);
    return response.data.reservation;
  } catch (error) {
    console.error('Error fetching reservations by ID:', error);
    throw new Error(error.response?.data?.message || 'Failed to fetch reservations by ID');
  }
};
export const fetchUserReservations = async () => {
  try {
    const response = await api.get('/reservations/user');
    return response.data;
  } catch (error) {
    console.error('Error fetching upcoming reservations:', error);
    throw new Error(error.response?.data?.message || 'Failed to fetch upcoming reservations');
  }
};

export const checkinReservation = async (gateId, reservationId) => {
  try {
    const response = await api.post(`/checkin/${gateId}`, {
      'reservation_id': reservationId
    });
    return response.data;
  } catch (error) {
    console.error('Error checking in reservation:', error);
    throw new Error(error.response?.data?.message || 'Failed to check in reservation');
  }
};

export const checkoutReservation = async (gateId, reservationId) => {
  try {
    await api.post(`/checkout/${gateId}`, {
      'reservation_id': reservationId
    });
  } catch (error) {
    console.error('Error checking out reservation:', error);
  }
};

export const payReservation = async (reservationId) => {
  try {
    const response = await api.post(`/payment/${reservationId}`);
    return response.data;
  } catch (error) {
    console.error('Error checking in reservation:', error);
    throw new Error(error.response?.data?.message || 'Failed to check in reservation');
  }
};

export const cancelReservation = async (reservationId) => {
  try {
    const response = await api.post(`/reservations/cancel/${reservationId}`);
    return response.data;
  } catch (error) {
    console.error('Error checking in reservation:', error);
    throw new Error(error.response?.data?.message || 'Failed to check in reservation');
  }
};

export const createReservation = async (reservationData) => {
  try {
    const response = await api.post('/reservations/create', reservationData);
    requestCache.invalidate(/\/reservations\//);// Invalidate parking lots cache
    return response.data;
  } catch (error) {
    console.error('Error creating reservation:', error);
    throw new Error(error.response?.data?.message || 'Failed to create reservation');
  }
};

// Notifications
export const fetchUserNotifications = async () => {
  try {
    const response = await api.get('/notifications');
    return response.data;
  } catch (error) {
    console.error('Error fetching notifications:', error);
    throw new Error(error.response?.data?.message || 'Failed to fetch notifications');
  }
};

// Gate
export const fetchGateByParkingLotId = async (parkingLotId) => {
  try {
    const response = await api.post(`/gate/${parkingLotId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching gate by parking lot ID:', error);
    throw new Error(error.response?.data?.message || 'Failed to fetch gate by parking lot ID');
  }
};

export default api;
