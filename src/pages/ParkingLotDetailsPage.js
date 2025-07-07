import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useParking } from '../context/ParkingContext';
import { useReservation } from '../context/ReservationContext';
import { toast } from 'react-toastify';
import ParkingSpaceVisualization from '../components/ParkingSpaceVisualization';
import ReservationForm from '../components/ReservationForm';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';

const ParkingLotDetailsPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const {
    parkingLots,
    parkingSpaces,
    loading,
    error,
    fetchParkingLots,
    fetchParkingSpaces,
    occupancyStats,
  } = useParking();

  const { createReservation } = useReservation();

  const [selectedSpace, setSelectedSpace] = useState(null);
  const [parkingLot, setParkingLot] = useState(null);
  const [reservationError, setReservationError] = useState(null);
  // const [occupancyStats, setOccupancyStats] = useState(null);


  // Load initial data
  useEffect(() => {
    const loadData = async () => {
      try {
        if (!parkingLots.length) {
          await fetchParkingLots();
        }
        await fetchParkingSpaces(id);

        // Fetch occupancy data
        // const stats = await calculateOccupancyStats(id);
        // setOccupancyStats(stats);
      } catch (err) {
        console.error('Error loading parking data:', err);
        toast.error('Failed to load parking data');
      }
    };
    loadData();
  }, [id, parkingLots.length, fetchParkingLots, fetchParkingSpaces]);

  // Update current parking lot when parkingLots changes
  useEffect(() => {
    const currentLot = parkingLots.find(lot => lot.id === id);
    setParkingLot(currentLot);
  }, [id, parkingLots]);

  // Handle space selection
  const handleSpaceSelect = useCallback((space) => {
    if (space.isOcupied) {
      toast.warn('This space is currently occupied');
      return;
    }
    setSelectedSpace(space);
  }, []);

  // Handle reservation submission
  const handleReservationSubmit = async (reservationData) => {
    try {
      setReservationError(null);
      reservationData = await createReservation(reservationData);
      await toast.success('Reservation created successfully!');
      setSelectedSpace(null);
    } catch (err) {
      console.error('Error creating reservation:', err);
      setReservationError(err.message);
      toast.error(err.message || 'Failed to create reservation');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <ErrorMessage message={error} />
      </div>
    );
  }

  if (!parkingLot) {
    return (
      <div className="container mx-auto px-4 py-8">
        <ErrorMessage message="Parking lot not found" />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-0 py-8">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-lg p-6 mb-8">
        <div className="flex items-center mb-4">
          <button
            onClick={() => navigate(-1)}
            className="mr-4 p-2 rounded-full hover:bg-gray-100 transition-colors"
          >
            <ArrowLeftIcon className="h-6 w-6 text-gray-600" />
          </button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">
              {parkingLot.name}
            </h1>
            <div className="text-gray-600 mt-2">
              <p>{parkingLot.address}</p>
              <p className="mt-1">
                Operating Hours: {parkingLot.operatingHours || '24/7'}
              </p>
              <p className="mt-1">
                Rate: ${parkingLot.hourlyRate}/hour
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Visualization Section */}
        <div className="lg:col-span-2">
          <ParkingSpaceVisualization
            spaces={parkingSpaces[id] || []}
            loading={loading}
            error={error}
            onSpaceSelect={handleSpaceSelect}
            selectedSpace={selectedSpace}
            parkingLot={parkingLot}
            occupancyStats={occupancyStats}
          />
        </div>

        {/* Reservation Section */}
        <div>
          <ReservationForm
            parkingLot={parkingLot}
            selectedSpace={selectedSpace}
            onReservationComplete={() => setSelectedSpace(null)}
            onSubmit={handleReservationSubmit}
            error={reservationError}
          />
        </div>
      </div>
    </div>
  );
};

export default ParkingLotDetailsPage;
