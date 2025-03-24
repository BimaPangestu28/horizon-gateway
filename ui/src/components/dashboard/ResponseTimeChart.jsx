import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Skeleton } from '@/components/ui/skeleton';

export default function ResponseTimeChart({ data, isLoading = false }) {
  if (isLoading) {
    return (
      <div className="bg-white overflow-hidden shadow rounded-lg">
        <div className="px-4 py-5 sm:px-6">
          <Skeleton className="h-6 w-48" />
        </div>
        <div className="px-4 py-5 sm:p-6">
          <Skeleton className="h-72 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="px-4 py-5 sm:px-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">Response Time (24h)</h3>
      </div>
      <div className="px-4 py-5 sm:p-6">
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart
              data={data}
              margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip formatter={(value) => [`${value} ms`, 'Response Time']} />
              <Legend />
              <Line 
                type="monotone" 
                dataKey="responseTime" 
                name="Response Time" 
                stroke="#10B981" 
                strokeWidth={2} 
                activeDot={{ r: 8 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}