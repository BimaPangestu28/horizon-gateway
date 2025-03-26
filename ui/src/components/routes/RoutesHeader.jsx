import React from 'react';
import { Link } from 'react-router-dom';
import { PlusCircle } from 'react-feather';

export default function RoutesHeader() {
  return (
    <div className="sm:flex sm:items-center">
      <div className="sm:flex-auto">
        <h1 className="text-2xl font-semibold text-gray-900">Routes</h1>
        <p className="mt-2 text-sm text-gray-700">
          A list of all API routes configured in your gateway.
        </p>
      </div>
      <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
        <Link
          to="/routes/new"
          className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
        >
          <PlusCircle className="w-4 h-4 mr-2" />
          Add Route
        </Link>
      </div>
    </div>
  );
}
