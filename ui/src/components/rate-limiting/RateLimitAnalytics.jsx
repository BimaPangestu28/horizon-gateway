import React, { useState, useEffect } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Calendar, Filter, Download } from 'react-feather';
import { Skeleton } from '@/components/ui/skeleton';

export default function RateLimitAnalytics({ rateLimit, loading = false }) {
  const [timeframe, setTimeframe] = useState('day');
  const [chartData, setChartData] = useState([]);
  const [summary, setSummary] = useState({
    totalBlocked: 0,
    blockedPercentage: 0,
    peakTime: '',
    totalRequests: 0
  });

  useEffect(() => {
    if (rateLimit && !loading) {
      generateMockData(timeframe);
    }
  }, [rateLimit, timeframe]);

  const generateMockData = (period) => {
    const now = new Date();
    const data = [];
    let totalRequests = 0;
    let totalBlocked = 0;
    let maxRequests = 0;
    let peakTime = '';

    if (period === 'day') {
      // Generate hourly data for the last 24 hours
      for (let i = 23; i >= 0; i--) {
        const hour = new Date(now);
        hour.setHours(now.getHours() - i);
        const time = hour.getHours() + ':00';
        
        const requests = Math.floor(Math.random() * 500) + 50;
        const blocked = Math.floor(Math.random() * requests * 0.2); // 0-20% blocked
        
        totalRequests += requests;
        totalBlocked += blocked;
        
        if (requests > maxRequests) {
          maxRequests = requests;
          peakTime = time;
        }
        
        data.push({
          time,
          requests,
          blocked
        });
      }
    } else if (period === 'week') {
      // Generate daily data for the last 7 days
      const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
      for (let i = 6; i >= 0; i--) {
        const day = new Date(now);
        day.setDate(now.getDate() - i);
        const dayName = days[day.getDay()];
        
        const requests = Math.floor(Math.random() * 3000) + 500;
        const blocked = Math.floor(Math.random() * requests * 0.2);
        
        totalRequests += requests;
        totalBlocked += blocked;
        
        if (requests > maxRequests) {
          maxRequests = requests;
          peakTime = dayName;
        }
        
        data.push({
          time: dayName,
          requests,
          blocked
        });
      }
    } else if (period === 'month') {
      // Generate weekly data for the last 4 weeks
      for (let i = 3; i >= 0; i--) {
        const weekStart = new Date(now);
        weekStart.setDate(now.getDate() - (i * 7));
        const weekEnd = new Date(weekStart);
        weekEnd.setDate(weekStart.getDate() + 6);
        
        const weekLabel = `Week ${4-i}`;
        
        const requests = Math.floor(Math.random() * 10000) + 2000;
        const blocked = Math.floor(Math.random() * requests * 0.2);
        
        totalRequests += requests;
        totalBlocked += blocked;
        
        if (requests > maxRequests) {
          maxRequests = requests;
          peakTime = weekLabel;
        }
        
        data.push({
          time: weekLabel,
          requests,
          blocked
        });
      }
    }

    const blockedPercentage = totalRequests > 0 ? ((totalBlocked / totalRequests) * 100).toFixed(2) : 0;

    setChartData(data);
    setSummary({
      totalBlocked,
      blockedPercentage,
      peakTime,
      totalRequests
    });
  };

  if (loading) {
    return <RateLimitAnalyticsSkeleton />;
  }

  return (
    <div className="bg-white shadow overflow-hidden sm:rounded-lg">
      <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <h3 className="text-lg leading-6 font-medium text-gray-900">
            Rate Limit Analytics
          </h3>
          <p className="mt-1 max-w-2xl text-sm text-gray-500">
            Request traffic and rate limit activity
          </p>
        </div>
        
        <div className="flex space-x-2 items-center">
          <div className="flex border rounded-md overflow-hidden">
            <button
              className={`px-3 py-1 text-sm ${timeframe === 'day' ? 'bg-indigo-100 text-indigo-700' : 'bg-gray-50 text-gray-700'}`}
              onClick={() => setTimeframe('day')}
            >
              Day
            </button>
            <button
              className={`px-3 py-1 text-sm ${timeframe === 'week' ? 'bg-indigo-100 text-indigo-700' : 'bg-gray-50 text-gray-700'}`}
              onClick={() => setTimeframe('week')}
            >
              Week
            </button>
            <button
              className={`px-3 py-1 text-sm ${timeframe === 'month' ? 'bg-indigo-100 text-indigo-700' : 'bg-gray-50 text-gray-700'}`}
              onClick={() => setTimeframe('month')}
            >
              Month
            </button>
          </div>
          
          <button className="rounded-md border border-gray-300 p-1 text-gray-500 hover:bg-gray-50">
            <Calendar className="h-5 w-5" />
          </button>
          
          <button className="rounded-md border border-gray-300 p-1 text-gray-500 hover:bg-gray-50">
            <Filter className="h-5 w-5" />
          </button>
          
          <button className="rounded-md border border-gray-300 p-1 text-gray-500 hover:bg-gray-50">
            <Download className="h-5 w-5" />
          </button>
        </div>
      </div>
      
      <div className="px-4 py-5 sm:p-0">
        <div className="flex flex-col md:flex-row border-b border-gray-200">
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <div className="text-sm font-medium text-gray-500">Total Requests</div>
            <div className="mt-1 text-2xl font-semibold text-gray-900">{summary.totalRequests.toLocaleString()}</div>
          </div>
          
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <div className="text-sm font-medium text-gray-500">Blocked Requests</div>
            <div className="mt-1 text-2xl font-semibold text-gray-900">{summary.totalBlocked.toLocaleString()}</div>
          </div>
          
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <div className="text-sm font-medium text-gray-500">Block Rate</div>
            <div className="mt-1 text-2xl font-semibold text-gray-900">{summary.blockedPercentage}%</div>
          </div>
          
          <div className="p-4 flex-1">
            <div className="text-sm font-medium text-gray-500">Peak Traffic</div>
            <div className="mt-1 text-2xl font-semibold text-gray-900">{summary.peakTime}</div>
          </div>
        </div>
        
        <div className="h-80 p-4">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={chartData}
              margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
            >
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="requests" fill="#6366F1" name="Requests" />
              <Bar dataKey="blocked" fill="#EF4444" name="Blocked" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}

function RateLimitAnalyticsSkeleton() {
  return (
    <div className="bg-white shadow overflow-hidden sm:rounded-lg">
      <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
        <div>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </div>
        
        <div className="flex space-x-2">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-8 w-8" />
          <Skeleton className="h-8 w-8" />
          <Skeleton className="h-8 w-8" />
        </div>
      </div>
      
      <div className="px-4 py-5 sm:p-0">
        <div className="flex flex-col md:flex-row border-b border-gray-200">
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-16 mt-2" />
          </div>
          
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-16 mt-2" />
          </div>
          
          <div className="p-4 flex-1 border-b md:border-b-0 md:border-r border-gray-200">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-16 mt-2" />
          </div>
          
          <div className="p-4 flex-1">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-16 mt-2" />
          </div>
        </div>
        
        <div className="h-80 p-4">
          <Skeleton className="h-full w-full" />
        </div>
      </div>
    </div>
  );
}