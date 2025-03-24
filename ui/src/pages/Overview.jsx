import React, { useState, useEffect, useRef } from 'react';
import { useToast } from '@/components/ui/use-toast';
import ApiService from '../services/ApiService';
import MetricsService from '../services/MetricsService';
import StatusCardGroup from '@/components/dashboard/StatusCardGroup';
import StatsCardGroup from '@/components/dashboard/StatsCardGroup';
import TrafficChart from '@/components/dashboard/TrafficChart';
import ResponseTimeChart from '@/components/dashboard/ResponseTimeChart';
import TopRoutesTable from '@/components/dashboard/TopRoutesTable';

export default function Overview() {
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [usingMockData, setUsingMockData] = useState(false);
  const notifiedRef = useRef(false);
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
    checkApiAvailability();
    fetchData();
    const interval = setInterval(fetchData, 15000);
    return () => clearInterval(interval);
  }, []);
  
  const checkApiAvailability = async () => {
    try {
      await ApiService.requestWithTimeout('/health', {}, 3000);
      setUsingMockData(false);
    } catch (error) {
      console.log('API health check failed, will use mock data');
      setUsingMockData(true);
      
      if (!notifiedRef.current) {
        toast({
          title: "Using simulated data",
          description: "The metrics API is not available, displaying mock data instead",
          duration: 5000,
        });
        notifiedRef.current = true;
      }
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const metricsResponse = await MetricsService.getMetrics({ period: '24h' });
      
      if (metricsResponse) {
        setStats({
          totalRequests: metricsResponse.totalRequests || 0,
          avgResponseTime: metricsResponse.avgResponseTime || 0,
          errorRate: metricsResponse.errorRate?.toFixed(2) || '0.00',
          activeRoutes: metricsResponse.activeRoutes || 0,
          requestsPerSecond: metricsResponse.requestsPerSecond || 0
        });

        if (metricsResponse.trafficHistory) {
          setTrafficHistory(metricsResponse.trafficHistory);
        }

        if (metricsResponse.responseTimeHistory) {
          setResponseTimeHistory(metricsResponse.responseTimeHistory);
        }

        if (metricsResponse.topRoutes) {
          setTopRoutes(metricsResponse.topRoutes);
        }

        if (metricsResponse.status) {
          setStatus(metricsResponse.status);
        }
      }
    } catch (error) {
      console.error('Failed to fetch overview data:', error);
      setError(error);
      
      if (!notifiedRef.current) {
        toast({
          title: "Failed to fetch data",
          description: error.message || "Couldn't connect to the API server",
          variant: "destructive",
          duration: 5000,
        });
        notifiedRef.current = true;
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-semibold text-gray-900">Gateway Overview</h1>
      <p className="mt-2 text-sm text-gray-700">
        Real-time monitoring and status of your API Gateway.
      </p>
      
      {usingMockData && (
        <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md text-blue-700">
          <p className="flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2h-1V9z" clipRule="evenodd" />
            </svg>
            Using simulated data - metrics API endpoint is not available.
          </p>
        </div>
      )}

      <StatusCardGroup status={status} />
      <StatsCardGroup stats={stats} />

      <div className="mt-8 grid grid-cols-1 gap-5 lg:grid-cols-2">
        <TrafficChart data={trafficHistory} isLoading={loading} />
        <ResponseTimeChart data={responseTimeHistory} isLoading={loading} />
      </div>

      <TopRoutesTable routes={topRoutes} isLoading={loading} />
    </div>
  );
}