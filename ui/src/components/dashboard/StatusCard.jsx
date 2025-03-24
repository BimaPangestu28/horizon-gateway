import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export default function StatusCard({ title, status, icon, description, isLoading = false }) {
  const statusColors = {
    'healthy': 'bg-green-50 text-green-700 ring-green-600/20',
    'degraded': 'bg-yellow-50 text-yellow-700 ring-yellow-600/20',
    'down': 'bg-red-50 text-red-700 ring-red-600/20',
    'up-to-date': 'bg-green-50 text-green-700 ring-green-600/20',
    'modified': 'bg-blue-50 text-blue-700 ring-blue-600/20',
  };

  const statusText = {
    'healthy': 'Healthy',
    'degraded': 'Degraded',
    'down': 'Down',
    'up-to-date': 'Up to date',
    'modified': 'Modified',
  };

  if (isLoading) {
    return (
      <div className="overflow-hidden rounded-lg bg-white shadow">
        <div className="p-5">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <Skeleton className="h-12 w-12 rounded-md" />
            </div>
            <div className="ml-4 space-y-2">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-4 w-24" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg bg-white shadow">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`h-12 w-12 rounded-md flex items-center justify-center ${statusColors[status] || statusColors['degraded']}`}>
              {icon}
            </div>
          </div>
          <div className="ml-4">
            <h3 className="text-lg font-medium text-gray-900">{title}</h3>
            <p className="text-sm text-gray-500">{description}</p>
            <p className={`mt-1 text-sm font-medium ${
              status === 'healthy' || status === 'up-to-date' 
                ? 'text-green-600' 
                : status === 'degraded' || status === 'modified' 
                  ? 'text-yellow-600' 
                  : 'text-red-600'
            }`}>
              {statusText[status] || status}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}