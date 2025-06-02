import React, { useState, useEffect } from 'react';
import { useParking } from '../context/ParkingContext';
import {
  MapIcon,
  TableCellsIcon,
  TruckIcon,
  ArchiveBoxIcon,
  ArrowsPointingOutIcon,
  ArrowsPointingInIcon
} from '@heroicons/react/24/outline';

const ParkingSpaceVisualization = ({ onSpaceSelect }) => {
  const {
    parkingSpaces,
    occupancyStats,
    loading,
    error,
    selectedParkingLot
  } = useParking();

  const [viewMode, setViewMode] = useState('floor-plan');
  const [selectedSpace, setSelectedSpace] = useState(null);
  const [zoomLevel, setZoomLevel] = useState(1);
  const [selectedFloor, setSelectedFloor] = useState(1);
  const [spacesByFloorAndType, setSpacesByFloorAndType] = useState({});
  const [statistics, setStatistics] = useState({
    total: 0,
    occupied: 0,
    available: 0,
    occupancyRate: 0
  });

  // Handle space selection
  const handleSpaceClick = (space) => {
    if (!space.is_occupied) {
      setSelectedSpace(space);
      if (onSpaceSelect) {
        onSpaceSelect(space);
      }
    }
  };

  // Zoom in/out functionality
  const handleZoomIn = () => {
    setZoomLevel(prev => Math.min(prev + 0.2, 2));
  };

  const handleZoomOut = () => {
    setZoomLevel(prev => Math.max(prev - 0.2, 0.6));
  };

  // Reset selection when parking lot changes
  useEffect(() => {
    setSelectedSpace(null);
  }, []);

  useEffect(() => {
    if (parkingSpaces) {
      const total = parkingSpaces[selectedParkingLot]?.length;
      const occupied = parkingSpaces[selectedParkingLot].filter(space => space.parkingSpace.is_occupied)?.length;
      const available = parkingSpaces[selectedParkingLot].filter(space => !space.parkingSpace.is_occupied)?.length;

      setStatistics({
        total,
        occupied,
        available,
      });
    }
    
    if (selectedParkingLot && parkingSpaces[selectedParkingLot]) {
      const grouped = parkingSpaces[selectedParkingLot].reduce((acc, space) => {
        const floor = space.parkingSpace.floor || 1;
        const type = space.parkingSpace.type || 'standard';
        if (!acc[floor]) acc[floor] = {};
        if (!acc[floor][type]) acc[floor][type] = [];
        acc[floor][type].push(space.parkingSpace);
        return acc;
      }, {});
      setSpacesByFloorAndType(grouped);
    } else {
      setSpacesByFloorAndType({});
    }
  }, [parkingSpaces, selectedParkingLot, occupancyStats]);

  // Render statistics panel
  const renderStatistics = () => (
    <div className="grid grid-cols-4 gap-4 mb-6 sm:grid-cols-3 xs:grid-cols-2">
      <div className="bg-white rounded-lg shadow-md p-4 flex flex-col items-center">
        <div className="text-sm text-gray-500 mb-1">Total Spaces</div>
        <div className="text-2xl font-bold text-gray-800">{statistics.total}</div>
      </div>
      <div className="bg-white rounded-lg shadow-md p-4 flex flex-col items-center">
        <div className="text-sm text-gray-500 mb-1">Available</div>
        <div className="text-2xl font-bold text-green-600">{statistics.available}</div>
      </div>
      <div className="bg-white rounded-lg shadow-md p-4 flex flex-col items-center">
        <div className="text-sm text-gray-500 mb-1">Occupied</div>
        <div className="text-2xl font-bold text-blue-600">{statistics.occupied}</div>
      </div>
    </div>
  );

  // Render view mode toggle
  const renderViewToggle = () => (
    <div className="flex bg-gray-100 rounded-lg p-1 mb-4 w-fit">
      <button
        onClick={() => setViewMode('floor-plan')}
        className={`flex items-center gap-2 px-4 py-2 rounded-md transition-all ${viewMode === 'floor-plan'
          ? 'bg-white text-blue-600 shadow-sm'
          : 'text-gray-600 hover:bg-gray-200'
          }`}
      >
        <MapIcon className="h-5 w-5" />
        Floor Plan
      </button>
      <button
        onClick={() => setViewMode('grid')}
        className={`flex items-center gap-2 px-4 py-2 rounded-md transition-all ${viewMode === 'grid'
          ? 'bg-white text-blue-600 shadow-sm'
          : 'text-gray-600 hover:bg-gray-200'
          }`}
      >
        <TableCellsIcon className="h-5 w-5" />
        Grid View
      </button>
    </div>
  );

  // Render zoom controls
  const renderZoomControls = () => (
    <div className="flex gap-2 mb-4">
      <button
        onClick={handleZoomIn}
        className="p-2 bg-white rounded-md shadow-sm hover:bg-gray-50 transition-colors"
      >
        <ArrowsPointingOutIcon className="h-5 w-5 text-gray-600" />
      </button>
      <button
        onClick={handleZoomOut}
        className="p-2 bg-white rounded-md shadow-sm hover:bg-gray-50 transition-colors"
      >
        <ArrowsPointingInIcon className="h-5 w-5 text-gray-600" />
      </button>
    </div>
  );

  // Render floor plan view
  const renderFloorPlan = () => {
    const floors = Object.entries(spacesByFloorAndType);

    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex gap-2 mb-6 overflow-x-auto pb-2">
          {floors.map(([floor]) => (
            <button
              key={`floor-${floor}`}
              className={`px-4 py-2 rounded-md transition-all whitespace-nowrap ${selectedFloor === floor
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              onClick={() => setSelectedFloor(floor)}
            >
              Floor {floor}
            </button>
          ))}
        </div>
        <div className="origin-top-left transition-transform duration-200" style={{ transform: `scale(${zoomLevel})` }}>
          {floors
            .filter(([floor]) => floor === selectedFloor)
            .map(([floor, types]) => (
              <div key={`floor-${floor}`} className="mb-6">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">Floor {floor}</h3>
                <div className="space-y-6">
                  {Object.entries(types).map(([type, spaces]) => (
                    <div key={`${floor}-${type}`} className="bg-gray-50 rounded-lg p-4">
                      <div className="flex items-center gap-2 mb-4 text-gray-700 font-medium">
                        {(type === 'standard' || type === 'large') ? (
                          <TruckIcon className="h-5 w-5 text-blue-600" />
                        ) : (
                          <ArchiveBoxIcon className="h-5 w-5 text-blue-600" />
                        )}
                        {type.charAt(0).toUpperCase() + type.slice(1)} Spaces
                      </div>
                      <div className="grid grid-cols-4 gap-4 sm:grid-cols-4 xs:grid-cols-2">
                        {spaces.map(space => (
                          <button
                            key={`${space.id}`}
                            onClick={() => handleSpaceClick(space)}
                            disabled={space.is_occupied}
                            className={`flex flex-col items-center justify-center p-4 rounded-lg transition-all ${space.is_occupied
                              ? 'bg-gray-200 cursor-not-allowed'
                              : space.id === selectedSpace?.id
                                ? 'bg-blue-100 border-2 border-blue-500'
                                : 'bg-green-100 hover:bg-green-200'
                              }`}
                          >
                            <div className="text-xl font-bold mb-1">{space.spaceNumber}</div>
                            <div className={`text-xs font-medium ${space.is_occupied ? 'text-gray-600' : 'text-green-600'
                              }`}>
                              {space.is_occupied ? 'Occupied' : 'Available'}
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
        </div>
      </div>
    );
  };

  // Render grid view
  const renderGridView = () => {
    const floors = Object.entries(spacesByFloorAndType);

    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex gap-2 mb-6 overflow-x-auto pb-2">
          {floors.map(([floor]) => (
            <button
              key={`floor-${floor}`}
              className={`px-4 py-2 rounded-md transition-all whitespace-nowrap ${selectedFloor === floor
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              onClick={() => setSelectedFloor(floor)}
            >
              Floor {floor}
            </button>
          ))}
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Space #</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Floor</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {floors
                .filter(([floor]) => floor === selectedFloor)
                .flatMap(([floor, types]) =>
                  Object.entries(types).flatMap(([type, spaces]) =>
                    spaces.map(space => (
                      <tr key={space.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{space.spaceNumber}</td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          <span className="flex items-center gap-2">
                            {(type === 'standard' || type === 'large') ? (
                              <TruckIcon className="h-5 w-5 text-blue-600" />
                            ) : (
                              <ArchiveBoxIcon className="h-5 w-5 text-blue-600" />
                            )}
                            {type.charAt(0).toUpperCase() + type.slice(1)}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{floor}</td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${space.is_occupied
                            ? 'bg-red-100 text-red-800'
                            : 'bg-green-100 text-green-800'
                            }`}>
                            {space.is_occupied ? 'Occupied' : 'Available'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                          <button
                            onClick={() => handleSpaceClick(space)}
                            disabled={space.is_occupied}
                            className={`px-3 py-1 rounded text-sm font-medium ${space.is_occupied
                              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                              : space.id === selectedSpace?.id
                                ? 'bg-blue-100 text-blue-700 border border-blue-500'
                                : 'bg-blue-50 text-blue-700 hover:bg-blue-100'
                              }`}
                          >
                            {space.is_occupied
                              ? 'Unavailable'
                              : space.id === selectedSpace?.id
                                ? 'Selected'
                                : 'Select'}
                          </button>
                        </td>
                      </tr>
                    ))
                  )
                )}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-md">
        <div className="flex">
          <div className="ml-3">
            <p className="text-sm text-red-700">
              {error}
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (!parkingSpaces || !selectedParkingLot || !parkingSpaces[selectedParkingLot] || parkingSpaces[selectedParkingLot].length === 0) {
    return (
      <div className="bg-yellow-50 border-l-4 border-yellow-500 p-4 rounded-md">
        <div className="flex">
          <div className="ml-3">
            <p className="text-sm text-yellow-700">
              No parking spaces available for this parking lot.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gray-50 p-6 rounded-xl">
      {renderStatistics()}
      <div className="flex justify-between items-center mb-6">
        {renderViewToggle()}
        {renderZoomControls()}
      </div>
      {viewMode === 'floor-plan' ? renderFloorPlan() : renderGridView()}
    </div>
  );
};

export default ParkingSpaceVisualization;
