import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useParking } from '../context/ParkingContext';
import { MapPinIcon, ArrowRightIcon } from '@heroicons/react/24/outline';

const ParkingLotsPage = () => {
  const { parkingLots, loading, error, fetchParkingLots } = useParking();
  const [searchTerm, setSearchTerm] = useState('');
  
  useEffect(() => {
    // Fetch parking lots when component mounts
    fetchParkingLots();
  }, [fetchParkingLots]);

  // Filter parking lots based on search term
  const filteredParkingLots = Array.isArray(parkingLots) 
  ? parkingLots.filter(lot => 
      lot.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      lot.address.toLowerCase().includes(searchTerm.toLowerCase())
    ) 
  : [];

  return (
    <div className="py-8">
      <div className="container mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-secondary-900 mb-2">Parking Lots</h1>
          <p className="text-secondary-600">
            Find and reserve parking spaces at your preferred location
          </p>
        </div>
        
        {/* Search Bar */}
        <div className="mb-8">
          <div className="relative">
            <input
              type="text"
              className="input pl-10"
              placeholder="Search by name or address..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <svg className="h-5 w-5 text-secondary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>
        
        {/* Loading and Error States */}
        {loading && (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
            <p className="mt-2 text-secondary-600">Loading parking lots...</p>
          </div>
        )}
        
        {error && !loading && (
          <div className="bg-red-50 border-l-4 border-red-500 p-4 mb-8">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-500" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">
                  {error}
                </p>
              </div>
            </div>
          </div>
        )}
        
        {/* Parking Lots Grid */}
        {!loading && !error && (
          <>
            {filteredParkingLots.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-secondary-600">No parking lots found matching your search.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {filteredParkingLots.map(lot => (
                  <div key={lot.id} className="card hover:shadow-lg transition-shadow duration-200">
                    <div className="mb-4">
                      <div className="flex items-center justify-between">
                        <h3 className="text-xl font-semibold text-secondary-900">{lot.name}</h3>
                        <span className="badge bg-primary-100 text-primary-800">
                        </span>
                      </div>
                      <div className="flex items-start mt-2">
                        <MapPinIcon className="h-5 w-5 text-secondary-500 mt-0.5 flex-shrink-0" />
                        <p className="ml-2 text-secondary-600">{lot.address}</p>
                      </div>
                    </div>
                    
                    <div className="mt-4 flex justify-between items-center">
                      <div>
                        <p className="text-sm text-secondary-500">
                          {lot.hourlyRate ? `$${lot.hourlyRate.toFixed(2)}/hour` : 'Free'}
                        </p>
                        {lot.operatingHours && (
                          <p className="text-sm text-secondary-500">
                            {lot.operatingHours}
                          </p>
                        )}
                      </div>
                      <Link 
                        to={`/parking-lots/${lot.id}`} 
                        className="btn btn-primary flex items-center"
                      >
                        View Details
                        <ArrowRightIcon className="ml-1 h-4 w-4" />
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default ParkingLotsPage;
