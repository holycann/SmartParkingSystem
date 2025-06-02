import React, { useState, useEffect, useMemo } from 'react';
import { useReservation } from '../context/ReservationContext';
import {
  ClockIcon,
  MapPinIcon,
  CurrencyDollarIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import CountdownTimer from '../components/CountdownTimer';

// Separate component for QR code modal
const QRCodeModal = ({ reservationId, cost, isCheckout, onClose, onCheckin, onPayment, onCheckout }) => {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white p-6 rounded-xl shadow-lg text-center w-[300px] relative">
        <h3 className="text-lg font-semibold mb-4">Reservation QR Code</h3>
        
        <button 
          onClick={() => {
            if (cost > 0 && !isCheckout) {
              onPayment(reservationId);
            } else if (isCheckout) {
              onCheckout(reservationId);
            } else {
              onCheckin(reservationId);
            }
          }}
        >
          <img
            src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${reservationId}`}
            alt="QR Code"
            className="mx-auto mb-4"
          />
        </button>

        {cost > 0 && !isCheckout ? (
          <p className="text-sm text-gray-600">
            Total Payment: ${cost} <br />
            Please scan this qr code for payment
          </p>
        ) : isCheckout ? (
          <p className="text-sm text-gray-600">Please scan this on gate for exit</p>
        ) : (
          <p className="text-sm text-gray-600">Please scan this on gate for checkin</p>
        )}

        <button
          onClick={onClose}
          className="absolute top-2 right-2 text-gray-400 hover:text-gray-600"
        >
          ✕
        </button>
      </div>
    </div>
  );
};

// Separate component for status indicator
const ReservationStatus = ({ status }) => {
  const getStatusColor = (status) => {
    switch (status) {
      case 'pending': return 'text-yellow-600';
      case 'confirmed':
      case 'active': return 'text-green-600';
      case 'cancelled': return 'text-red-600';
      case 'completed': return 'text-blue-600';
      default: return 'text-gray-600';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'active':
      case 'completed':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'expired':
      case 'cancelled':
        return <XCircleIcon className="h-5 w-5 text-red-500" />;
      case 'pending':
        return <ArrowPathIcon className="h-5 w-5 animate-spin text-yellow-500" />;
      default:
        return <ClockIcon className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 'pending': return 'Pending';
      case 'active': return 'Active';
      case 'completed': return 'Completed';
      case 'expired': return 'Expired';
      default: return status?.charAt(0).toUpperCase() + status?.slice(1);
    }
  };

  return (
    <div className={`flex items-center gap-2 ${getStatusColor(status)}`}>
      {getStatusIcon(status)}
      <span className="font-medium">{getStatusText(status)}</span>
    </div>
  );
};

// Separate component for reservation card
const ReservationCard = ({ reservation, onQrCode, onCancel }) => {
  const now = new Date();
  const reservationDate = reservation?.reservation_date ? new Date(reservation.reservation_date) : null;
  const expiredAt = reservation?.expired_at ? new Date(reservation.expired_at) : null;
  const checkInTime = reservation?.checkin_time ? new Date(reservation.checkin_time) : null;

  const canCheckIn = reservationDate && 
                     now.toDateString() === reservationDate.toDateString() && 
                     reservation?.status === 'pending';

  const formatDate = (date) => {
    return new Intl.DateTimeFormat('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    }).format(date);
  };

  const formatDateTime = (date) => {
    return new Intl.DateTimeFormat('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  };

  return (
    <div className="bg-white rounded-xl shadow-sm hover:shadow-md transition-transform hover:-translate-y-1 p-6">
      <div className="flex justify-between items-center mb-6">
        <ReservationStatus status={reservation.status} />
        <div className="text-xs text-gray-400">{reservation.id}</div>
      </div>
      
      <div className="space-y-4 mb-6">
        <div className="flex gap-4">
          <MapPinIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
          <div>
            <h3 className="text-sm text-gray-500 mb-1">Location</h3>
            <p className="text-gray-800 font-medium">
              {reservation.parking_lot_name} - Space {reservation.space_number}
            </p>
          </div>
        </div>

        {reservation.status !== 'active' && reservationDate && (
          <div className="flex gap-4">
            <ClockIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
            <div>
              <h3 className="text-sm text-gray-500 mb-1">Reservation Date</h3>
              <p className="text-gray-800 font-medium">
                {formatDate(reservationDate)}
              </p>
            </div>
          </div>
        )}

        {reservation.status === 'pending' && expiredAt && (
          <div className="flex gap-4">
            <ClockIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
            <div>
              <h3 className="text-sm text-gray-500 mb-1">Expired In</h3>
              <p className="text-gray-800 font-medium">
                {formatDate(expiredAt)}
              </p>
            </div>
          </div>
        )}

        {reservation.status === 'active' && checkInTime && (
          <div className="flex gap-4">
            <ClockIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
            <div>
              <h3 className="text-sm text-gray-500 mb-1">Check-in Time</h3>
              <p className="text-gray-800 font-medium">
                {formatDateTime(checkInTime)}
              </p>
            </div>
          </div>
        )}

        <div className="flex gap-4">
          <ClockIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
          <div>
            <h3 className="text-sm text-gray-500 mb-1">Duration</h3>
            <p className="text-gray-800 font-medium">{reservation.duration} Hours</p>
          </div>
        </div>

        {reservation.status === 'active' && checkInTime && (
          <div className="flex gap-4">
            <ClockIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
            <div>
              <h3 className="text-sm text-gray-500 mb-1">Time Limit</h3>
              <CountdownTimer
                targetTime={
                  checkInTime.getTime() + reservation.duration * 60 * 60 * 1000
                }
              />
            </div>
          </div>
        )}

        <div className="flex gap-4">
          <CurrencyDollarIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
          <div>
            <h3 className="text-sm text-gray-500 mb-1">Total Fees</h3>
            <p className="text-gray-800 font-medium">${reservation.total_cost}</p>
          </div>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex gap-2">
          {canCheckIn && (
            <button
              onClick={() => onQrCode(reservation.id)}
              className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 text-sm"
            >
              Check In
            </button>
          )}

          {reservation.status === 'pending' && !canCheckIn && (
            <p className="text-xs text-gray-500 italic">
              Check-in button will appear when the reservation date arrives.
            </p>
          )}

          {reservation.status === 'active' && reservation.payment_status === 'paid' ? (
            <button
              onClick={() => onQrCode(reservation.id, 0, 'checkout')}
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 text-sm"
            >
              Check Out
            </button>
          ) : reservation.status === 'active' && (
            <button
              onClick={() => onQrCode(reservation.id, reservation.total_cost)}
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 text-sm"
            >
              Pay Now
            </button>
          )}

          {reservation.status === 'pending' && (
            <button
              onClick={() => onCancel(reservation.id)}
              className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 text-sm"
            >
              Cancel Reservation
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

// Loading component
const LoadingState = () => (
  <div className="flex flex-col items-center justify-center py-12 text-gray-500">
    <ArrowPathIcon className="h-8 w-8 animate-spin text-blue-600 mb-4" />
    <p>Loading reservations...</p>
  </div>
);

// Error component
const ErrorState = ({ message }) => (
  <div className="flex flex-col items-center justify-center py-12 text-red-600">
    <XCircleIcon className="h-8 w-8 mb-4" />
    <p>{message}</p>
  </div>
);

// Empty state component
const EmptyState = ({ tab }) => (
  <div className="text-center py-12 bg-gray-50 rounded-lg text-gray-500">
    <p>No {tab} reservations found.</p>
  </div>
);

// Main component
const ReservationsPage = () => {
  const {
    userReservation,
    loading,
    error,
    loadUserReservation,
    checkinReservation,
    checkoutReservation,
    payReservation,
    cancelReservation,
  } = useReservation();

  const [activeTab, setActiveTab] = useState('upcoming');
  const [modalState, setModalState] = useState({
    isOpen: false,
    reservationId: null,
    cost: 0,
    isCheckout: false
  });

  useEffect(() => {
    loadUserReservation();
  }, []);

  const filteredReservations = useMemo(() => {
    if (!userReservation || !Array.isArray(userReservation)) return [];
    
    return userReservation.filter(reservation => {
      if (activeTab === 'upcoming' && reservation.status === 'pending') {
        return true;
      } else if (activeTab === 'active' && reservation.status === 'active') {
        return true;
      } else if (activeTab === 'history' && (reservation.status === 'completed' || reservation.status === 'cancelled')) {
        return true;
      }
      return false;
    });
  }, [userReservation, activeTab]);

  const handleQrCode = (reservationId, cost = 0, type = null) => {
    setModalState({
      isOpen: true,
      reservationId,
      cost,
      isCheckout: type === 'checkout'
    });
  };

  const closeModal = () => {
    setModalState({
      isOpen: false,
      reservationId: null,
      cost: 0,
      isCheckout: false
    });
  };

  const handleCancelReservation = async (reservationId) => {
    try {
      await cancelReservation(reservationId);
    } catch (error) {
      // Error handling is managed by the context
    }
  };

  const handleCheckin = async (reservationId) => {
    try {
      await checkinReservation(reservationId);
      closeModal();
    } catch (error) {
      // Error handling is managed by the context
    }
  };

  const handlePayment = async (reservationId) => {
    try {
      await payReservation(reservationId);
      closeModal();
    } catch (error) {
      // Error handling is managed by the context
    }
  };

  const handleCheckout = async (reservationId) => {
    try {
      await checkoutReservation(reservationId);
      closeModal();
    } catch (error) {
      // Error handling is managed by the context
    }
  };

  const renderContent = () => {
    if (loading) {
      return <LoadingState />;
    }

    if (error) {
      return <ErrorState message={error} />;
    }

    if (!filteredReservations || filteredReservations.length === 0) {
      return <EmptyState tab={activeTab} />;
    }

    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredReservations.map((reservation) => (
          <ReservationCard
            key={reservation.id}
            reservation={reservation}
            onQrCode={handleQrCode}
            onCancel={handleCancelReservation}
          />
        ))}
      </div>
    );
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-semibold text-gray-800">My Reservations</h1>
        <button
          onClick={loadUserReservation}
          className="flex items-center gap-2 text-blue-600 hover:text-blue-800"
        >
          <ArrowPathIcon className="h-5 w-5" />
          <span>Refresh</span>
        </button>
      </div>

      <div className="flex gap-4 mb-8 border-b border-gray-200">
        <button
          className={`py-3 px-6 font-medium transition ${activeTab === 'upcoming' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-blue-600'}`}
          onClick={() => setActiveTab('upcoming')}
        >
          Upcoming
        </button>
        <button
          className={`py-3 px-6 font-medium transition ${activeTab === 'active' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-blue-600'}`}
          onClick={() => setActiveTab('active')}
        >
          Active
        </button>
        <button
          className={`py-3 px-6 font-medium transition ${activeTab === 'history' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-blue-600'}`}
          onClick={() => setActiveTab('history')}
        >
          History
        </button>
      </div>

      {renderContent()}

      {modalState.isOpen && (
        <QRCodeModal
          reservationId={modalState.reservationId}
          cost={modalState.cost}
          isCheckout={modalState.isCheckout}
          onClose={closeModal}
          onCheckin={handleCheckin}
          onPayment={handlePayment}
          onCheckout={handleCheckout}
        />
      )}
    </div>
  );
};

export default ReservationsPage;