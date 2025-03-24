import React from 'react';
import StatCard from './StatCard';

export default function StatsCardGroup({ stats, isLoading = false }) {
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
    <div className="mt-8 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard 
        title="Total Requests" 
        value={formatNumber(stats.totalRequests)} 
        trend="up" 
        trendValue="12.3%" 
        isLoading={isLoading}
      />
      <StatCard 
        title="Requests/sec" 
        value={formatNumber(stats.requestsPerSecond)} 
        trend="up" 
        trendValue="5.7%" 
        isLoading={isLoading}
      />
      <StatCard 
        title="Avg Response Time" 
        value={`${stats.avgResponseTime} ms`} 
        trend={stats.avgResponseTime < 100 ? "down" : "up"} 
        trendValue="3.2%" 
        isLoading={isLoading}
      />
      <StatCard 
        title="Error Rate" 
        value={`${stats.errorRate}%`} 
        trend={parseFloat(stats.errorRate) < 1 ? "down" : "up"} 
        trendValue="0.5%" 
        isLoading={isLoading}
      />
    </div>
  );
}