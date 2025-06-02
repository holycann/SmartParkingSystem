import { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { toast } from 'react-toastify';
import {
  fetchUserReservations,
  fetchReservationsById,
  createReservation as apiCreateReservation,
  checkinReservation as apiCheckinReservation,
  checkoutReservation as apiCheckoutReservation,
  payReservation as apiPayReservation,
  cancelReservation as apiCancelReservation,
} from '../services/apiService';

import { useAuth } from './AuthContext';
import { useWebSocket } from './WebSocketContext';
import { registerHandlerByPrefix, WS_MESSAGE_TYPES } from '../services/webSocketService';

const ReservationContext = createContext();

export const useReservation = () => useContext(ReservationContext);

export const ReservationProvider = ({ children }) => {
  const { isAuthenticated } = useAuth();
  const { wsConnected } = useWebSocket();
  const [userReservation, setUserReservation] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Use refs to keep track of the latest state without triggering re-renders
  const userReservationRef = useRef(userReservation);

  // Update refs when state changes
  useEffect(() => {
    userReservationRef.current = userReservation;
  }, [userReservation]);

  const loadUserReservation = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchUserReservations();
      console.log("API returned reservations:", data.reservations?.length);
      setUserReservation(data.reservations || []);
      console.log("State updated with reservations:",
        data.reservations?.length || 0);
      setError(null);
    } catch (error) {
      console.error("Error loading upcoming reservations:", error);
      setError('Failed to load reservations');
      toast.error('Failed to load reservations');
    } finally {
      setLoading(false);
    }
  }, []);

  const loadReservationById = useCallback(async (reservationId) => {
    setLoading(true);
    try {
      const data = await fetchReservationsById(reservationId);
      setError(null);
      return data;
    } catch (error) {
      console.error("Error loading reservation by ID:", error);
      setError('Failed to load reservation');
      toast.error('Failed to load reservation');
    } finally {
      setLoading(false);
    }
  }, []);

  const createReservation = async (reservationData) => {
    setLoading(true);
    try {
      const newReservation = await apiCreateReservation(reservationData);
      setError(null);
      toast.success('Reservation created successfully');
      return newReservation;
    } catch (error) {
      console.error("Error creating reservation:", error);
      setError('Failed to create reservation');
      toast.error('Failed to create reservation');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const checkinReservation = async (reservationId) => {
    setLoading(true);
    try {
      const response = await apiCheckinReservation(reservationId);

      // Update the reservation in state
      setUserReservation(prev =>
        prev.map(res =>
          res.id === reservationId
            ? { ...res, status: 'active', checkin_time: new Date().toISOString() }
            : res
        )
      );

      toast.success('Check-in successful');
      return response;
    } catch (error) {
      console.error("Error checking in:", error);
      setError('Failed to check in reservation');
      toast.error('Failed to check in reservation');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const checkoutReservation = async (reservationId) => {
    setLoading(true);
    try {
      const response = await apiCheckoutReservation(reservationId);

      // Move from upcoming to history
      const reservation = userReservationRef.current.find(r => r.id === reservationId);
      if (reservation) {
        setUserReservation(prev => prev?.filter(r => r.id !== reservationId));
      }

      toast.success('Check-out successful');
      return response;
    } catch (error) {
      console.error("Error checking out:", error);
      setError('Failed to check out reservation');
      toast.error('Failed to check out reservation');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const payReservation = async (reservationId) => {
    setLoading(true);
    try {
      const response = await apiPayReservation(reservationId);

      // Update the reservation payment status in state
      setUserReservation(prev =>
        prev.map(res =>
          res.id === reservationId
            ? { ...res, payment_status: 'paid' }
            : res
        )
      );

      toast.success('Payment successful');
      return response;
    } catch (error) {
      console.error("Error processing payment:", error);
      setError('Failed to process payment');
      toast.error('Failed to process payment');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const cancelReservation = async (reservationId) => {
    setLoading(true);
    try {
      const response = await apiCancelReservation(reservationId);

      // Update the reservation in state
      setUserReservation(prev =>
        prev.map(res =>
          res.id === reservationId
            ? { ...res, status: 'cancelled' }
            : res
        )
      );

      toast.success('Reservation cancelled successfully');
      return response;
    } catch (error) {
      console.error("Error cancelling reservation:", error);
      setError('Failed to cancel reservation');
      toast.error('Failed to cancel reservation');
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Function to convert camelCase to snake_case
  const camelToSnake = (obj) => {
    if (!obj || typeof obj !== 'object') return obj;

    return Object.fromEntries(
      Object.entries(obj).map(([key, value]) => {
        const parts = key.split(/(?=[A-Z])/); // Pisahkan saat huruf besar

        if (parts.length > 1) {
          const last = parts.pop().toLowerCase(); // Bagian terakhir
          const joined = parts.join('').toLowerCase(); // Gabungkan sisanya
          return [`${joined}_${last}`, value];
        }

        return [key.toLowerCase(), value]; // Jika tidak ada huruf besar
      })
    );
  };


  // Handle reservation WebSocket messages
  const handleReservationUpdate = useCallback((data) => {
    console.log('Received WebSocket message:', data);

    try {
      switch (data.type) {
        case WS_MESSAGE_TYPES.RESERVATION_ADDED:
          // Convert camelCase keys to snake_case
          const newReservation = camelToSnake(data.payload);

          setUserReservation(prev => {
            // Check if reservation already exists
            if (prev.some(res => res.id === newReservation.id)) {
              return prev;
            }

            const updated = [...prev, newReservation];
            console.log('Updated upcoming reservations after add:', updated);
            return updated;
          });

          toast.info('New reservation added');
          break;

        case WS_MESSAGE_TYPES.RESERVATION_UPDATED:
          // Convert camelCase keys to snake_case
          const updatedReservation = camelToSnake(data.payload);

          setUserReservation(prev => {
            const updated = prev.map(res =>
              res.id === updatedReservation.id ? updatedReservation : res
            );
            console.log('Updated upcoming reservations after update:', updated);
            return updated;
          });

          toast.info('Reservation updated');
          break;

        case WS_MESSAGE_TYPES.RESERVATION_COMPLETED:
        case WS_MESSAGE_TYPES.RESERVATION_CANCELLED:
          const reservationId = data.payload.id;

          // Find the reservation in upcoming before removing
          const reservationToMove = userReservationRef.current.find(r => r.id === reservationId);

          // Remove from upcoming
          setUserReservation(prev => {
            const updated = prev?.filter(res => res.id !== reservationId);
            console.log('Updated upcoming reservations after removal:', updated);
            return updated;
          });

          // Add to history if it's completed
          if (data.type === WS_MESSAGE_TYPES.RESERVATION_COMPLETED && reservationToMove) {
            const completedReservation = {
              ...reservationToMove,
              status: 'completed',
              checkout_time: data.payload.checkout_time || new Date().toISOString()
            };

            toast.info('Reservation completed');
          } else {
            toast.info('Reservation cancelled');
          }
          break;

        default:
          console.log(`Unhandled reservation WS type: ${data.type}`);
          break;
      }
    } catch (error) {
      console.error('Error handling WebSocket message:', error);
    }
  }, []);

  // WebSocket connection setup
  useEffect(() => {
    if (wsConnected) {
      console.log('Registering reservation handlers');
      const unregisterHandler = registerHandlerByPrefix('RESERVATION_', handleReservationUpdate);

      return () => {
        console.log('Unregistering reservation handlers');
        unregisterHandler();
      };
    }
  }, [wsConnected, handleReservationUpdate]);
  // Load initial data
  useEffect(() => {
    if (isAuthenticated && userReservation.length === 0) {
      console.log('Loading initial reservation data');
      loadUserReservation();
    }
  }, [isAuthenticated, loadUserReservation, userReservation.length]);

  const value = {
    wsConnected,
    userReservation,
    loading,
    error,
    loadReservationById,
    loadUserReservation,
    createReservation,
    checkinReservation,
    checkoutReservation,
    payReservation,
    cancelReservation,
  };

  return (
    <ReservationContext.Provider value={value}>
      {children}
    </ReservationContext.Provider>
  );
};