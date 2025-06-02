import React from 'react';
import { Link } from 'react-router-dom';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';

const NotFoundPage = () => {
  // Log page view
  console.log('404 Not Found Page viewed');
  
  return (
    <div className="min-h-[calc(100vh-200px)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full text-center">
        <div className="flex justify-center mb-4">
          <ExclamationTriangleIcon className="h-16 w-16 text-yellow-500" />
        </div>
        <h2 className="text-3xl font-extrabold text-secondary-900 mb-2">
          404 - Page Not Found
        </h2>
        <p className="text-secondary-600 mb-8">
          The page you are looking for doesn't exist or has been moved.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link to="/" className="btn btn-primary">
            Go to Home
          </Link>
          <Link to="/parking-lots" className="btn btn-secondary">
            Find Parking
          </Link>
        </div>
      </div>
    </div>
  );
};

export default NotFoundPage;
