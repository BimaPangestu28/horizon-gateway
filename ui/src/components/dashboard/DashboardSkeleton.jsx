import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export default function DashboardSkeleton() {
  return (
    <div className="py-6">
      {/* Stats Card Skeletons */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-6">
        {[1, 2, 3, 4].map((item) => (
          <div key={item} className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <Skeleton className="h-10 w-10 rounded-full" />
                </div>
                <div className="ml-5 w-full">
                  <Skeleton className="h-4 w-1/3 mb-2" />
                  <Skeleton className="h-8 w-1/2" />
                </div>
              </div>
            </div>
            <div className="bg-gray-50 px-5 py-3">
              <Skeleton className="h-4 w-2/5" />
            </div>
          </div>
        ))}
      </div>

      {/* Content Skeleton */}
      <div className="bg-white shadow overflow-hidden sm:rounded-lg">
        <div className="px-4 py-5 sm:px-6 flex justify-between">
          <div>
            <Skeleton className="h-7 w-48 mb-2" />
            <Skeleton className="h-4 w-64" />
          </div>
          <div className="flex space-x-2">
            <Skeleton className="h-10 w-28" />
            <Skeleton className="h-10 w-28" />
          </div>
        </div>
        <div className="border-t border-gray-200 px-4 py-5 sm:p-6">
          {/* Search bar skeleton */}
          <div className="mb-6 flex">
            <Skeleton className="h-10 flex-grow" />
            <Skeleton className="h-10 w-24 ml-3" />
          </div>
          
          {/* Table skeleton */}
          <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  {[1, 2, 3, 4, 5].map((item) => (
                    <th key={item} className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      <Skeleton className="h-4 w-24" />
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {[1, 2, 3, 4, 5].map((row) => (
                  <tr key={row}>
                    {[1, 2, 3, 4, 5].map((col) => (
                      <td key={`${row}-${col}`} className="px-6 py-4 whitespace-nowrap">
                        <Skeleton className="h-5 w-full max-w-xs" />
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          
          {/* Pagination skeleton */}
          <div className="flex items-center justify-between mt-4">
            <div className="flex-1 flex justify-between sm:hidden">
              <Skeleton className="h-8 w-24" />
              <Skeleton className="h-8 w-24" />
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <Skeleton className="h-5 w-36" />
              </div>
              <div>
                <Skeleton className="h-8 w-64" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}