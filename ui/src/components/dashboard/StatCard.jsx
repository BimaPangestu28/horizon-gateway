import React from 'react';
import { ArrowUp, ArrowDown } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';

export default function StatCard({ title, value, trend, trendValue, isLoading = false }) {
  if (isLoading) {
    return (
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="p-5">
          <div className="flex items-center">
            <div className="flex-1 space-y-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-8 w-20" />
            </div>
            <div className="flex flex-shrink-0 ml-4">
              <Skeleton className="h-6 w-16 rounded-full" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-1">
            <dt className="text-sm font-medium text-gray-500 truncate">
              {title}
            </dt>
            <dd className="mt-1 text-3xl font-semibold text-gray-900">
              {value}
            </dd>
          </div>
          <div className="flex flex-shrink-0 ml-4">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              trend === 'up' 
                ? 'bg-red-100 text-red-800' 
                : 'bg-green-100 text-green-800'
            }`}>
              {trend === 'up' ? (
                <ArrowUp className="mr-1 h-3 w-3 text-red-500" />
              ) : (
                <ArrowDown className="mr-1 h-3 w-3 text-green-500" />
              )}
              {trendValue}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}