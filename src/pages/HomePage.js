import { useReservation } from '../context/ReservationContext';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import {
  CurrencyDollarIcon,
  ClockIcon,
  CalendarIcon,
  ClipboardDocumentIcon,
  CogIcon,
} from '@heroicons/react/24/outline';

const StatCard = ({ icon: Icon, title, value, type }) => {
  const colorClasses = {
    primary: "border-l-4 border-blue-500",
    info: "border-l-4 border-green-500",
    success: "border-l-4 border-yellow-500",
    warning: "border-l-4 border-red-500"
  };

  const iconColors = {
    primary: "text-blue-500",
    info: "text-green-500",
    success: "text-yellow-500",
    warning: "text-red-500"
  };

  return (
    <div className={`bg-white p-6 rounded-lg shadow-md flex items-center gap-4 hover:-translate-y-1 transition-transfo$${colorClasses[type]}`}>
      <Icon className={`w-10 h-10 ${iconColors[type]}`} />
      <div>
        <h3 className="text-gray-700 text-sm font-medium mb-1">{title}</h3>
        <div className="text-2xl font-bold text-gray-800">{value}</div>
      </div>
    </div>
  );
};

const QuickAction = ({ icon: Icon, title, description, onClick, type }) => {
  const bgColors = {
    primary: "bg-blue-50 hover:bg-blue-100",
    info: "bg-green-50 hover:bg-green-100",
    success: "bg-yellow-50 hover:bg-yellow-100",
    warning: "bg-red-50 hover:bg-red-100"
  };

  const iconColors = {
    primary: "text-blue-600",
    info: "text-green-600",
    success: "text-yellow-600",
    warning: "text-red-600"
  };

  return (
    <button
      className={`p-4 rounded-lg shadow-sm flex flex-col items-center text-center transition-colors ${bgColors[type]}`}
      onClick={onClick}
    >
      <Icon className={`w-8 h-8 mb-2 ${iconColors[type]}`} />
      <h4 className="text-gray-800 font-medium mb-1">{title}</h4>
      <p className="text-gray-600 text-sm">{description}</p>
    </button>
  );
};


const HomePage = () => {
  const { currentUser } = useAuth();
  
  const {
    userReservation,
  } = useReservation();

  const navigate = useNavigate();

  return (
    <div className="max-w-7xl mx-auto px-4 py-8 bg-gray-50">
      {/* Header */}
      <div className="bg-white p-8 rounded-xl shadow-md text-center mb-8">
        <h1 className="text-4xl font-bold text-gray-800 mb-2">Smart Parking System</h1>
        <p className="text-xl text-gray-600">Welcome back, {currentUser?.firstName || 'User'} {currentUser?.lastName || 'User'}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        <StatCard
          icon={CalendarIcon}
          title="Total Reservations"
          value={userReservation?.length}
          type="primary"
        />
        <StatCard
          icon={CurrencyDollarIcon}
          title="Total Spent"
          value={`$${userReservation
            .filter(reservation => reservation.status === "completed")
            .reduce((acc, reservation) => acc + reservation.total_cost, 0)
            .toFixed(2)}`}
          type="info"
        />
        <StatCard
          icon={ClockIcon}
          title="Hours Parked"
          value={`${userReservation
            .filter(reservation => reservation.status === "completed")
            .reduce((acc, reservation) => acc + reservation.duration, 0)} hours`}
          type="success"
        />
      </div>

      {/* Quick Actions */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 gap-4">
          <QuickAction
            icon={CalendarIcon}
            title="New Reservation"
            description="Book a parking space"
            onClick={() => navigate('/parking-lots')}
            type="primary"
          />
          <QuickAction
            icon={ClipboardDocumentIcon}
            title="Reservations"
            description="Manage your bookings"
            onClick={() => navigate('/reservations')}
            type="success"
          />
          <QuickAction
            icon={CogIcon}
            title="Settings"
            description="Account preferences"
            onClick={() => { navigate('/profile') }}
            type="primary"
          />
        </div>
      </div>
    </div>
  );
};

export default HomePage;
