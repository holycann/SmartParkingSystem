import React, { useState } from 'react';
import { ArrowPathIcon, ExclamationCircleIcon, CalendarIcon, ClockIcon, CurrencyDollarIcon, CheckCircleIcon, TruckIcon } from '@heroicons/react/24/outline';

const ReservationForm = ({ parkingLot, selectedSpace, onReservationComplete, onSubmit }) => {
  const [reservationDate, setReservationDate] = useState('');
  const [durationTime, setDurationTime] = useState('');
  const [licensePlate, setLicencePlate] = useState('');
  const [vehicleType, setVehicleType] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Calculate total cost
  const calculateTotalCost = () => {
    if (!durationTime || !parkingLot?.hourlyRate) return 0;

    return durationTime * parkingLot.hourlyRate;
  };


  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!selectedSpace) {
      setError('Please select a parking space');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      await onSubmit({
        parkingLotID: parkingLot.id,
        parkingSpaceID: selectedSpace.id,
        vehicleType: vehicleType,
        licensePlate: licensePlate,
        reservationDate: reservationDate,
        duration: parseInt(durationTime),
        totalCost: calculateTotalCost()
      });
    } catch (err) {
      setError(err.message || 'Failed to create reservation');
    } finally {
      setLoading(false);
    }
  };

  if (!selectedSpace) {
    return (
      <div className="w-full max-w-xl mx-auto">
        <div className="bg-blue-50 rounded-2xl p-8 text-center shadow-lg">
          <div className="bg-blue-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-6">
            <ExclamationCircleIcon className="w-8 h-8 text-blue-600" />
          </div>
          <h3 className="text-xl font-semibold text-blue-800 mb-3">Select a Parking Space</h3>
          <p className="text-gray-600">
            Please select an available parking space from the visualization to continue with your reservation.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-xl mx-auto">
      <div className="bg-white rounded-2xl shadow-lg overflow-hidden transition-all duration-300">
        <div className="flex justify-between items-center p-6 border-b border-gray-100 bg-gradient-to-r from-blue-600 to-gray-800 text-white">
          <h3 className="text-xl font-semibold m-0">Make a Reservation</h3>
          <div className="bg-white bg-opacity-20 px-3 py-1.5 rounded-full text-sm font-medium">Space {selectedSpace.number}</div>
        </div>

        <form onSubmit={handleSubmit} className="p-6">
          <div className="flex flex-col">
            <div className="flex flex-col mb-4">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <TruckIcon className="w-5 h-5 text-blue-600" />
                Vehicle Type
              </label>
              <select
                value={vehicleType}
                onChange={(e) => setVehicleType(e.target.value)}
                className="p-3 border border-gray-200 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                required
              >
                <option value="">Select vehicle type</option>
                <option value="car">Car</option>
                <option value="truck">Truck</option>
                <option value="bus">Bus</option>
              </select>
            </div>

            <div className="flex flex-col mb-4">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <TruckIcon className="w-5 h-5 text-blue-600" />
                License Plate
              </label>
              <input
                type="text"
                value={licensePlate}
                onChange={(e) => setLicencePlate(e.target.value)}
                className="p-3 border border-gray-200 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                required
              />
            </div>

            <div className="flex flex-col mb-4">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <CalendarIcon className="w-5 h-5 text-blue-600" />
                Date
              </label>
              <input
                type="date"
                value={reservationDate}
                onChange={(e) => setReservationDate(e.target.value)}
                min={new Date().toISOString().split('T')[0]}
                className="p-3 border border-gray-200 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                required
              />
            </div>

            <div className="flex flex-col mb-4">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <ClockIcon className="w-5 h-5 text-blue-600" />
                Duration Time
              </label>
              <input
                type="number"
                min={1}
                max={24}
                value={durationTime}
                onChange={(e) => setDurationTime(e.target.value)}
                className="p-3 border border-gray-200 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
                required
              />
            </div>
          </div>

          <div className="bg-gray-50 rounded-lg p-5 mb-6">
            <div className="flex justify-between items-center py-2">
              <span className="text-gray-700 font-medium">Hourly Rate:</span>
              <span className="font-semibold text-gray-800">${parkingLot?.hourlyRate.toFixed(2)}/hr</span>
            </div>
            <div className="flex justify-between items-center py-2 border-t border-dashed border-gray-200 mt-2 pt-2">
              <span className="flex items-center gap-2 text-gray-700 font-medium">
                <CurrencyDollarIcon className="w-5 h-5 text-blue-600" />
                Total Cost:
              </span>
              <span className="text-xl font-semibold text-blue-600">${calculateTotalCost().toFixed(2)}</span>
            </div>
          </div>

          {error && (
            <div className="flex items-center gap-3 bg-red-50 text-red-600 p-4 rounded-lg mb-6 border-l-4 border-red-600">
              <ExclamationCircleIcon className="w-5 h-5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="flex justify-end gap-4 sm:flex-row xs:flex-col">
            <button
              type="button"
              onClick={onReservationComplete}
              className="px-6 py-3 bg-gray-50 text-gray-700 border border-gray-200 rounded-lg font-medium hover:bg-gray-100 transition-all disabled:bg-gray-100 disabled:text-gray-400 sm:w-auto xs:w-full xs:justify-center"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-all disabled:bg-gray-400 sm:w-auto xs:w-full xs:justify-center"
              disabled={loading}
            >
              {loading ? (
                <>
                  <ArrowPathIcon className="w-5 h-5 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <CheckCircleIcon className="w-5 h-5" />
                  Create Reservation
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ReservationForm;
