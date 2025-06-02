import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import { toast } from 'react-toastify';
import { useAuth } from './AuthContext';
import { useWebSocket } from './WebSocketContext';
import {
  fetchParkingLots as fetchParkingLotsApi,
  fetchParkingSpaces as fetchParkingSpacesApi,
  fetchOccupancyStats as fetchOccupancyStatsApi,
} from '../services/apiService';

import {
  registerHandlerByPrefix,
  WS_MESSAGE_TYPES
} from '../services/webSocketService';

const ParkingContext = createContext();

export const useParking = () => useContext(ParkingContext);

export const ParkingProvider = ({ children }) => {
  const { isAuthenticated, userId } = useAuth();
  const { wsConnected } = useWebSocket();

  const [selectedParkingLot, setSelectedParkingLot] = useState(null);
  const [parkingLots, setParkingLots] = useState([]);
  const [parkingSpaces, setParkingSpaces] = useState({});
  const [occupancyStats, setOccupancyStats] = useState({});
  const [notifications, setNotifications] = useState([]);
  const [gateStatus, setGateStatus] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const updateParkingSpace = useCallback((lotId, spaceId, isOccupied) => {

    setParkingSpaces(prev => {
      const spaces = { ...prev };

      if (spaces[lotId]) {
        const updatedSpaces = spaces[lotId].map(space => {
          if (space.parkingSpace.id === spaceId) {
            const updatedParkingSpace = {
              ...space.parkingSpace,
              is_occupied: isOccupied
            };

            return {
              ...space,
              parkingSpace: updatedParkingSpace
            };
          }
          return space;
        });

        spaces[lotId] = updatedSpaces;
      } else {
        console.warn(`Lot ID ${lotId} not found in parkingSpaces.`);
      }

      return spaces;
    });
  }, []);

  // Handle WebSocket messages
  const handleWebSocketMessage = useCallback((data) => {
    try {
      switch (data.type) {
        case WS_MESSAGE_TYPES.PARKING_UPDATE:
          const parkingUpdate = data.payload;
          updateParkingSpace(
            parkingUpdate.parkingLotId,
            parkingUpdate.spaceId,
            parkingUpdate.isOccupied
          );
          break;

        case WS_MESSAGE_TYPES.NOTIFICATION_UPDATE:
          const notificationUpdate = data.payload;
          // Only add the notification if it's for the current user
          if (notificationUpdate.userId === userId) {
            setNotifications(prev => [
              {
                id: notificationUpdate.notificationId,
                type: notificationUpdate.type,
                message: notificationUpdate.message,
                timestamp: notificationUpdate.timestamp,
                isRead: false
              },
              ...prev
            ]);

            // Show a toast for the new notification
            toast.info(notificationUpdate.message);
          }
          break;

        case WS_MESSAGE_TYPES.GATE_UPDATE:
          const gateUpdate = data.payload;
          setGateStatus(prev => ({
            ...prev,
            [gateUpdate.gateId]: {
              status: gateUpdate.status,
              timestamp: gateUpdate.timestamp
            }
          }));
          break;

        case WS_MESSAGE_TYPES.OCCUPANCY_UPDATE:
          if (data.stats) {
            setOccupancyStats(prev => ({
              ...prev,
              currentOccupancy: data.stats.currentOccupancy || 0
            }));
          }
          break;

        default:
          console.warn('Unknown WebSocket message type:', data.type);
      }
    } catch (err) {
      console.error('Error handling WebSocket message:', err);
    }
  }, [updateParkingSpace, userId]);


  // Initialize WebSocket connection
  useEffect(() => {
    if (wsConnected) {
      console.log('Registering reservation handlers');
      const unregisterHandler = registerHandlerByPrefix('PARKING_', handleWebSocketMessage);

      return () => {
        console.log('Unregistering reservation handlers');
        unregisterHandler();
      };
    }
  }, [wsConnected, handleWebSocketMessage]);

  // Fetch parking lots
  const fetchParkingLots = useCallback(async () => {
    try {
      setLoading(true);

      const response = await fetchParkingLotsApi();
      const lots = response.parkingLots || [];

      // Ambil occupancy stats untuk setiap lot
      const lotsWithStats = await Promise.all(
        lots.map(async (lot) => {
          try {
            const occupancy = await fetchOccupancyStatsApi(lot.id);
            return {
              ...lot,
              availableSpaces: occupancy.availableSpaces,
              totalSpaces: occupancy.totalSpaces,
            };
          } catch (err) {
            console.error(`Failed to fetch occupancy for lot ${lot.id}`, err);
            return {
              ...lot,
              availableSpaces: 0,
              totalSpaces: lot.totalSpaces || 0,
            };
          }
        })
      );

      setParkingLots(lotsWithStats);
      setError(null);
    } catch (err) {
      setError('Failed to fetch parking lots. Please try again.');
      toast.error('Failed to fetch parking lots');
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch parking spaces for a specific lot
  const fetchParkingSpaces = useCallback(async (lotId) => {
    try {
      setLoading(true);
      const [spacesResponse, statsResponse] = await Promise.all([
        fetchParkingSpacesApi(lotId),
        fetchOccupancyStatsApi(lotId),
      ]);

      setParkingSpaces(prevSpaces => ({
        ...prevSpaces,
        [lotId]: (spacesResponse.spaces || []).map(space => ({
          ...space,
          occupied: space.is_occupied
        }))
      }));

      setOccupancyStats(prevStats => ({
        ...prevStats,
        currentOccupancy: statsResponse.stats?.currentOccupancy || 0
      }));

      // Subscribe to real-time updates for this parking lot
      setSelectedParkingLot(lotId);
      setError(null);
    } catch (err) {
      setError('Failed to fetch parking spaces. Please try again.');
      toast.error('Failed to fetch parking spaces');
    } finally {
      setLoading(false);
    }
  }, []);

  // Load initial data on authentication change
  useEffect(() => {
    if (isAuthenticated) {
      fetchParkingLots();
    }
  }, [isAuthenticated, fetchParkingLots]);

  // Mark a notification as read
  const markNotificationAsRead = (notificationId) => {
    setNotifications(prev =>
      prev.map(notification =>
        notification.id === notificationId
          ? { ...notification, isRead: true }
          : notification
      )
    );
  };

  // Mark all notifications as read
  const markAllNotificationsAsRead = () => {
    setNotifications(prev =>
      prev.map(notification => ({ ...notification, isRead: true }))
    );
  };

  // Get unread notification count
  const getUnreadNotificationCount = () => {
    return notifications?.filter(notification => !notification.isRead).length;
  };

  // Get parking lot by ID
  const getParkingLotById = (lotId) => {
    return parkingLots.find(lot => lot.id === lotId) || null;
  };

  return (
    <ParkingContext.Provider
      value={{
        parkingLots,
        selectedParkingLot,
        parkingSpaces,
        occupancyStats,
        notifications,
        gateStatus,
        loading,
        error,
        wsConnected,
        fetchParkingLots,
        fetchParkingSpaces,
        markNotificationAsRead,
        markAllNotificationsAsRead,
        getUnreadNotificationCount,
        getParkingLotById,
      }}
    >
      {children}
    </ParkingContext.Provider>
  );
};

export default ParkingContext;
