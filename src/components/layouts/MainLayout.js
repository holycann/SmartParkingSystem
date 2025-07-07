import React, { useState } from 'react';
import { Outlet, Link, useNavigate } from 'react-router-dom';
import { Transition } from '@headlessui/react';
import { 
  MapIcon, 
  HomeIcon, 
  CalendarIcon,
  TruckIcon,
  BuildingLibraryIcon,
  Bars3Icon,
  XMarkIcon,
  UserIcon,
  CogIcon,
  ArrowRightOnRectangleIcon
} from '@heroicons/react/24/outline';
import { useAuth } from '../../context/AuthContext';

const MainLayout = () => {
  const navigate = useNavigate();
  const { currentUser, logout } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);

  // Sample notifications data
  const [notifications] = useState([
    {
      title: 'New Reservation',
      message: 'Your reservation for Lot A12 has been confirmed',
      time: '2 hours ago'
    },
    {
      title: 'Payment Received',
      message: 'Payment of RM15.00 for reservation #12345 has been processed',
      time: '5 hours ago'
    }
  ]);

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            {/* Logo and Title */}
            <div className="flex items-center">
              <Link to="/" className="flex items-center">
                <MapIcon className="h-8 w-8 text-primary-600" />
                <span className="ml-2 text-xl font-bold text-primary-900">Smart Parking</span>
              </Link>
            </div>
            
            {/* Desktop Navigation */}
            <nav className="hidden md:flex items-center space-x-6">
              <Link to="/" className="text-secondary-600 hover:text-primary-600 font-medium">
                <HomeIcon className="h-5 w-5 inline-block mr-1" />
                Home
              </Link>
              <Link to="/reservations" className="text-secondary-600 hover:text-primary-600 font-medium">
                <CalendarIcon className="h-5 w-5 inline-block mr-1" />
                Reservations
              </Link>
              <Link to="/parking-lots" className="text-secondary-600 hover:text-primary-600 font-medium">
                <BuildingLibraryIcon className="h-5 w-5 inline-block mr-1" />
                Parking Lots
              </Link>
            </nav>
            
            {/* User Actions */}
            <div className="flex items-center space-x-4">
              {currentUser ? (
                <div className="relative">
                  <button 
                    onClick={() => setIsProfileMenuOpen(!isProfileMenuOpen)}
                    className="flex items-center space-x-2 hover:text-primary-600"
                  >
                    <UserIcon className="h-5 w-5" />
                    <span>{currentUser.firstName} {currentUser.lastName}</span>
                  </button>
                  {isProfileMenuOpen && (
                    <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5">
                      <div className="py-1">
                        <button
                          onClick={() => navigate('/profile')}
                          className="block w-full px-4 py-2 text-sm text-secondary-700 hover:bg-secondary-50"
                        >
                          <CogIcon className="h-4 w-4 inline-block mr-2" />
                          Profile Settings
                        </button>
                        <button
                          onClick={logout}
                          className="block w-full px-4 py-2 text-sm text-secondary-700 hover:bg-secondary-50"
                        >
                          <ArrowRightOnRectangleIcon className="h-4 w-4 inline-block mr-2" />
                          Log out
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <>
                  <Link 
                    to="/login" 
                    className="text-secondary-600 hover:text-primary-600 font-medium"
                  >
                    Log in
                  </Link>
                  <Link 
                    to="/register" 
                    className="bg-primary-600 text-white px-4 py-2 rounded-md hover:bg-primary-700"
                  >
                    Sign up
                  </Link>
                </>
              )}
              <button 
                className="md:hidden p-2 rounded-md text-secondary-600 hover:text-primary-600 focus:outline-none"
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              >
                {isMobileMenuOpen ? (
                  <XMarkIcon className="h-6 w-6" />
                ) : (
                  <Bars3Icon className="h-6 w-6" />
                )}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Mobile Menu */}
      <Transition
        show={isMobileMenuOpen}
        enter="transition-opacity duration-75"
        enterFrom="opacity-0"
        enterTo="opacity-100"
        leave="transition-opacity duration-150"
        leaveFrom="opacity-100"
        leaveTo="opacity-0"
      >
        <div className="px-2 pt-2 pb-3 space-y-1 bg-white shadow-sm">
          <Link
            to="/"
            className="block px-3 py-2 rounded-md text-base font-medium text-secondary-700 hover:text-primary-600 hover:bg-secondary-50"
          >
            <HomeIcon className="h-5 w-5 inline-block mr-2" />
            Home
          </Link>
          <Link
            to="/reservations"
            className="block px-3 py-2 rounded-md text-base font-medium text-secondary-700 hover:text-primary-600 hover:bg-secondary-50"
          >
            <CalendarIcon className="h-5 w-5 inline-block mr-2" />
            Reservations
          </Link>
          <Link
            to="/vehicles"
            className="block px-3 py-2 rounded-md text-base font-medium text-secondary-700 hover:text-primary-600 hover:bg-secondary-50"
          >
            <TruckIcon className="h-5 w-5 inline-block mr-2" />
            Vehicles
          </Link>
          <Link
            to="/parking-lots"
            className="block px-3 py-2 rounded-md text-base font-medium text-secondary-700 hover:text-primary-600 hover:bg-secondary-50"
          >
            <BuildingLibraryIcon className="h-5 w-5 inline-block mr-2" />
            Parking Lots
          </Link>
        </div>
      </Transition>

      {/* Main Content */}
      <main className="flex-1 py-8">
        <Outlet />
      </main>
    </div>
  );
};

export default MainLayout;
