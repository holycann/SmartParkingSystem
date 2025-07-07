import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { toast } from 'react-toastify';
import { useAuth } from './AuthContext';
import {
  fetchUserNotifications,
} from '../services/apiService';
import {
  connectWebSocket,
  registerHandlerByPrefix,
  WS_MESSAGE_TYPES
} from '../services/webSocketService';

const NotificationContext = createContext();

export const useNotification = () => useContext(NotificationContext);

export const NotificationProvider = ({ children }) => {
  const { isAuthenticated, currentUser } = useAuth();
  
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [settings, setSettings] = useState({
    emailNotifications: true,
    pushNotifications: true,
    reservationReminders: true,
    marketingEmails: false
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [wsConnected, setWsConnected] = useState(false);

  // Fetch user notifications
  const fetchNotifications = useCallback(async () => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      const [notificationsData, unreadData, settingsData] = await Promise.all([
        fetchUserNotifications(),
      ]);
      
      setNotifications(notificationsData.notifications || []);
      setUnreadCount(unreadData.count || 0);
      setSettings(settingsData.settings || {
        emailNotifications: true,
        pushNotifications: true,
        reservationReminders: true,
        marketingEmails: false
      });
      
      setError(null);
    } catch (err) {
      setError('Failed to fetch notifications. Please try again.');
      console.error('Failed to fetch notifications:', err);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  // Handle notification WebSocket messages
  const handleNotificationUpdate = useCallback((data) => {
    switch (data.type) {
      case WS_MESSAGE_TYPES.NOTIFICATION_UPDATE:
        if (currentUser.id === data.payload.userId) {
          toast.info(data.payload.message);
        }
        
        if (data.payload.userId === "") {
          toast.success(data.payload.message);
        }
        break;
      default:
        toast.warn(`Unhandled notification WS type: ${data.type}`);
        break;
    }
  }, []);

  // WebSocket connection setup
  useEffect(() => {
    let cleanup = () => { };

    if (isAuthenticated) {
      // Register notification update handler
      const unregisterHandler = registerHandlerByPrefix('NOTIFICATION_', handleNotificationUpdate);

      // Connect to WebSocket
      cleanup = connectWebSocket(
        null, // Using specialized prefix handlers instead of general message handler
        () => {
          setWsConnected(true);
          console.log('WebSocket connected for notification updates');
        },
        () => {
          setWsConnected(false);
          console.log('WebSocket disconnected for notification updates');
        }
      );

      // Return cleanup function that combines both the WebSocket cleanup and handler unregistration
      return () => {
        cleanup();
        unregisterHandler();
      };
    }

    return cleanup;
  }, [isAuthenticated, handleNotificationUpdate]);

  // Load initial notifications
  useEffect(() => {
    if (isAuthenticated) {
      fetchNotifications();
    } else {
      setNotifications([]);
      setUnreadCount(0);
    }
  }, [fetchNotifications, isAuthenticated]);

  const value = {
    notifications,
    unreadCount,
    settings,
    loading,
    error,
    wsConnected,
    fetchNotifications,
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
};
