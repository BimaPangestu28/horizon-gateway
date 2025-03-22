import React, { useState, useEffect } from 'react';
import { 
  BarChart, 
  Bar, 
  LineChart, 
  Line, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend, 
  ResponsiveContainer 
} from 'recharts';
import { ArrowUp, ArrowDown, Activity, Server, Shield, Clock } from 'react-feather';
import ApiService from '../services/ApiService';

export default function Overview() {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    totalRequests: 0,
    avgResponseTime: 0,
    errorRate: 0,
    activeRoutes: 0,
    requestsPerSecond: 0
  });
  const [trafficHistory, setTrafficHistory] = useState([]);
  const [responseTimeHistory, setResponseTimeHistory] = useState([]);
  const [topRoutes, setTopRoutes] = useState([]);
  const [status, setStatus] = useState({
    proxyStatus: 'healthy',
    adminStatus: 'healthy',
    configStatus: 'up-to-date'
  });

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 15000); // Refresh every 15 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      // In a real app, we'd call the API
      // For now, generate fake data
      setStats({
        totalRequests: Math.floor(Math.random() * 1000000) + 500000,
        avgResponseTime: Math.floor(Math.random() * 200) + 80,
        errorRate: (Math.random() * 2).toFixed(2),
        activeRoutes: Math.floor(Math.random() * 20) + 5,
        requestsPerSecond: Math.floor(Math.random() * 500) + 100
      });

      setTrafficHistory(generateTrafficData());
      setResponseTimeHistory(generateResponseTimeData());
      setTopRoutes(generateTopRoutes());
      setStatus({
        proxyStatus: Math.random() > 0.1 ? 'healthy' : 'degraded',
        adminStatus: Math.random() > 0.05 ? 'healthy' : 'degraded',
        configStatus: Math.random() > 0.2 ? 'up-to-date' : 'modified'
      });
    } catch (error) {
      console.error('Failed to fetch overview data:', error);
    } finally {
      setLoading(false);
    }
  };

  const generateTrafficData = () => {
    const data = [];
    const now = new Date();
    
    for (let i = 23; i >= 0; i--) {
      const hour = new Date(now);
      hour.setHours(now.getHours() - i);
      
      data.push({
        time: hour.getHours() + ':00',
        requests: Math.floor(Math.random() * 400) + 100,
        errors: Math.floor(Math.random() * 20)
      });
    }
    
    return data;
  };

  const generateResponseTimeData = () => {
    const data = [];
    const now = new Date();
    
    for (let i = 23; i >= 0; i--) {
      const hour = new Date(now);
      hour.setHours(now.getHours() - i);
      
      data.push({
        time: hour.getHours() + ':00',
        responseTime: Math.floor(Math.random() * 300) + 70
      });
    }
    
    return data;
  };

  const generateTopRoutes = () => {
    return [
      { name: '/api/*', requests: 35243, avgResponseTime: 125 },
      { name: '/users/*', requests: 27891, avgResponseTime: 187 },
      { name: '/products/*', requests: 19432, avgResponseTime: 93 },
      { name: '/auth/*', requests: 12985, avgResponseTime: 210 },
      { name: '/search/*', requests: 9872, avgResponseTime: 156 }
    ];
  };

  const formatNumber = (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num;
  };

  if (loading && !stats.totalRequests) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900">Gateway Overview</h1>
      <p className="mt-2 text-sm text-gray-700">
        Real-time monitoring and status of your API Gateway.
      </p>

      {/* Status Cards */}
      <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <StatusCard 
          title="Gateway Status" 
          status={status.proxyStatus} 
          icon={<Server className="h-6 w-6" />} 
          description="Main proxy service status" 
        />
        <StatusCard 
          title="Admin API Status" 
          status={status.adminStatus} 
          icon={<Shield className="h-6 w-6" />} 
          description="Admin API service status" 
        />
        <StatusCard 
          title="Configuration" 
          status={status.configStatus} 
          icon={<Activity className="h-6 w-6" />} 
          description="Configuration sync status" 
        />
      </div>

      {/* Stats */}
      <div className="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard 
          title="Total Requests" 
          value={formatNumber(stats.totalRequests)} 
          trend="up" 
          trendValue="12.3%" 
        />
        <StatCard 
          title="Requests/sec" 
          value={formatNumber(stats.requestsPerSecond)} 
          trend="up" 
          trendValue="5.7%" 
        />
        <StatCard 
          title="Avg Response Time" 
          value={`${stats.avgResponseTime} ms`} 
          trend={stats.avgResponseTime < 100 ? "down" : "up"} 
          trendValue="3.2%" 
        />
        <StatCard 
          title="Error Rate" 
          value={`${stats.errorRate}%`} 
          trend={stats.errorRate < 1 ? "down" : "up"} 
          trendValue="0.5%" 
        />
      </div>

      {/* Charts */}
      <div className="mt-8 grid grid-cols-1 gap-5 lg:grid-cols-2">
        {/* Traffic Chart */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">Traffic (24h)</h3>
          </div>
          <div className="px-4 py-5 sm:p-6">
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart
                  data={trafficHistory}
                  margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="requests" name="Requests" fill="#4F46E5" />
                  <Bar dataKey="errors" name="Errors" fill="#EF4444" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        {/* Response Time Chart */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">Response Time (24h)</h3>
          </div>
          <div className="px-4 py-5 sm:p-6">
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart
                  data={responseTimeHistory}
                  margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip formatter={(value) => [`${value} ms`, 'Response Time']} />
                  <Legend />
                  <Line type="monotone" dataKey="responseTime" name="Response Time" stroke="#10B981" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      </div>

      {/* Top Routes */}
      <div className="mt-8">
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">Top Routes</h3>
            <p className="mt-1 max-w-2xl text-sm text-gray-500">Most active routes in the last 24 hours.</p>
          </div>
          <div className="border-t border-gray-200">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Route
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Requests
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Avg Response Time
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {topRoutes.map((route, idx) => (
                  <tr key={idx}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {route.name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatNumber(route.requests)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {route.avgResponseTime} ms
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        route.avgResponseTime < 150 
                          ? 'bg-green-100 text-green-800' 
                          : route.avgResponseTime < 200 
                            ? 'bg-yellow-100 text-yellow-800' 
                            : 'bg-red-100 text-red-800'
                      }`}>
                        {route.avgResponseTime < 150 
                          ? 'Healthy' 
                          : route.avgResponseTime < 200 
                            ? 'Warning' 
                            : 'Slow'}
                      </span>
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

function StatusCard({ title, status, icon, description }) {
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

  return (
    <div className="overflow-hidden rounded-lg bg-white shadow">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`h-12 w-12 rounded-md flex items-center justify-center ${statusColors[status]}`}>
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
              {statusText[status]}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, trend, trendValue }) {
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