import React, { useState, useEffect } from 'react';
import { 
  LineChart, 
  Line, 
  BarChart, 
  Bar, 
  PieChart, 
  Pie, 
  Cell, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Legend, 
  ResponsiveContainer,
  Label
} from 'recharts';
import { Clock, RefreshCw, Filter, Download } from 'react-feather';
import ApiService from '../services/ApiService';

export default function Analytics() {
  const [timeRange, setTimeRange] = useState('24h');
  const [loading, setLoading] = useState(true);
  const [metrics, setMetrics] = useState({
    traffic: [],
    responseTime: [],
    statusCodes: [],
    topRoutes: [],
    errorRates: [],
    cachePerformance: {
      hits: 0,
      misses: 0,
      ratio: 0
    }
  });

  useEffect(() => {
    fetchMetrics();
  }, [timeRange]);

  const fetchMetrics = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would call the API
      // For now, use sample data
      
      // Generate traffic data
      const trafficData = generateTimeSeriesData(24, 1000, 5000);
      
      // Generate response time data
      const responseTimeData = generateTimeSeriesData(24, 50, 300);
      
      // Generate status code data
      const statusCodes = [
        { name: '2xx', value: 85, color: '#10B981' },
        { name: '3xx', value: 10, color: '#3B82F6' },
        { name: '4xx', value: 4, color: '#F59E0B' },
        { name: '5xx', value: 1, color: '#EF4444' }
      ];
      
      // Generate top routes data
      const topRoutes = [
        { name: '/api/*', requests: 3500 },
        { name: '/web/*', requests: 2800 },
        { name: '/public/*', requests: 2200 },
        { name: '/admin/*', requests: 1500 },
        { name: '/api/monitor/*', requests: 1000 }
      ];
      
      // Generate error rates data
      const errorRates = generateTimeSeriesData(24, 0, 5, true);
      
      // Generate cache performance data
      const cacheHits = Math.floor(Math.random() * 80000) + 20000;
      const cacheMisses = Math.floor(Math.random() * 20000) + 5000;
      const cacheRatio = (cacheHits / (cacheHits + cacheMisses) * 100).toFixed(2);
      
      setMetrics({
        traffic: trafficData,
        responseTime: responseTimeData,
        statusCodes,
        topRoutes,
        errorRates,
        cachePerformance: {
          hits: cacheHits,
          misses: cacheMisses,
          ratio: cacheRatio
        }
      });
    } catch (error) {
      console.error('Failed to fetch metrics', error);
    } finally {
      setLoading(false);
    }
  };

  const generateTimeSeriesData = (points, min, max, isPercentage = false) => {
    const data = [];
    let previous = Math.floor(Math.random() * (max - min)) + min;
    
    for (let i = 0; i < points; i++) {
      const hour = i.toString().padStart(2, '0') + ':00';
      // Create some variation but with a trend
      const change = Math.random() * (max - min) * 0.1;
      let value = previous + (Math.random() > 0.5 ? change : -change);
      
      // Keep within bounds
      value = Math.max(min, Math.min(max, value));
      
      if (isPercentage) {
        value = Math.min(100, value);
      }
      
      data.push({
        time: hour,
        value: Math.round(value * 100) / 100
      });
      
      previous = value;
    }
    
    return data;
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

  return (
    <div>
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Analytics</h1>
          <p className="mt-2 text-sm text-gray-700">
            Detailed metrics and analytics for your API Gateway.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none space-x-3">
          <select
            className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
          >
            <option value="1h">Last Hour</option>
            <option value="6h">Last 6 Hours</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
          </select>
          <button
            type="button"
            className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            onClick={fetchMetrics}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </button>
          <button
            type="button"
            className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <Download className="h-4 w-4 mr-2" />
            Export
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center items-center py-20">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-indigo-500"></div>
        </div>
      ) : (
        <div className="mt-8 space-y-8">
          {/* Traffic Over Time */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Traffic Over Time</h2>
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={metrics.traffic}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis>
                    <Label value="Requests" angle={-90} position="insideLeft" style={{ textAnchor: 'middle' }} />
                  </YAxis>
                  <Tooltip formatter={(value) => [formatNumber(value), 'Requests']} />
                  <Legend />
                  <Line type="monotone" dataKey="value" name="Requests" stroke="#3B82F6" strokeWidth={2} dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Response Time and Status Codes */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Response Time */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Response Time</h2>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metrics.responseTime}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis>
                      <Label value="ms" angle={-90} position="insideLeft" style={{ textAnchor: 'middle' }} />
                    </YAxis>
                    <Tooltip formatter={(value) => [value + ' ms', 'Response Time']} />
                    <Legend />
                    <Line type="monotone" dataKey="value" name="Response Time" stroke="#10B981" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Status Codes */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Status Codes</h2>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={metrics.statusCodes}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                      label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                    >
                      {metrics.statusCodes.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => [`${value}%`, 'Percentage']} />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Top Routes and Error Rates */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Top Routes */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Top Routes</h2>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={metrics.topRoutes}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis>
                      <Label value="Requests" angle={-90} position="insideLeft" style={{ textAnchor: 'middle' }} />
                    </YAxis>
                    <Tooltip formatter={(value) => [formatNumber(value), 'Requests']} />
                    <Legend />
                    <Bar dataKey="requests" name="Requests" fill="#8B5CF6" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Error Rates */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">Error Rate</h2>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={metrics.errorRates}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time" />
                    <YAxis domain={[0, 10]}>
                      <Label value="%" angle={-90} position="insideLeft" style={{ textAnchor: 'middle' }} />
                    </YAxis>
                    <Tooltip formatter={(value) => [value + '%', 'Error Rate']} />
                    <Legend />
                    <Line type="monotone" dataKey="value" name="Error Rate" stroke="#EF4444" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Cache Performance */}
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Cache Performance</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-indigo-50 rounded-lg p-4">
                <div className="text-indigo-600 text-3xl font-bold">{formatNumber(metrics.cachePerformance.hits)}</div>
                <div className="text-indigo-800 mt-1">Cache Hits</div>
              </div>
              <div className="bg-indigo-50 rounded-lg p-4">
                <div className="text-indigo-600 text-3xl font-bold">{formatNumber(metrics.cachePerformance.misses)}</div>
                <div className="text-indigo-800 mt-1">Cache Misses</div>
              </div>
              <div className="bg-indigo-50 rounded-lg p-4">
                <div className="text-indigo-600 text-3xl font-bold">{metrics.cachePerformance.ratio}%</div>
                <div className="text-indigo-800 mt-1">Hit Ratio</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}