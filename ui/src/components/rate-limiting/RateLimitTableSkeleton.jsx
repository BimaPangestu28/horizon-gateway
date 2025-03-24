import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';

export default function RateLimitTableSkeleton() {
  // Creates an array of 4 items to represent skeleton rows
  const skeletonRows = Array(4).fill(null);

  return (
    <div className="flex flex-col">
      <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">
                    Route
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Type
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Limit
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Window
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Key By
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Status
                  </th>
                  <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                    Current Usage
                  </th>
                  <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {skeletonRows.map((_, index) => (
                  <tr key={index}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium sm:pl-6">
                      <Skeleton className="h-5 w-32" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <Skeleton className="h-5 w-24" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <Skeleton className="h-5 w-16" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <Skeleton className="h-5 w-20" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <Skeleton className="h-5 w-20" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <Skeleton className="h-6 w-20 rounded-full" />
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm">
                      <div className="flex flex-col space-y-1">
                        <Skeleton className="h-3 w-24" />
                        <Skeleton className="h-3 w-16" />
                        <Skeleton className="h-3 w-20" />
                      </div>
                    </td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <div className="flex justify-end space-x-2">
                        <Skeleton className="h-4 w-4 rounded-full" />
                        <Skeleton className="h-4 w-4 rounded-full" />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}