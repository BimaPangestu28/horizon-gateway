import React from 'react';
import { Layers, BarChart2, Clock, AlertTriangle } from 'lucide-react';
import StatCard from './StatCard';

export default function DashboardStats({ stats }) {
  return (
    <div className="grid grid-cols-1 gap-4 mb-8 sm:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="Total Routes"
        value={stats.routes}
        icon={<Layers className="w-5 h-5" />}
        trend="up"
        trendValue="12%"
        description="Active API routes"
        color="blue"
      />
      <StatCard
        title="Requests/sec"
        value={stats.requestsPerSecond}
        icon={<BarChart2 className="w-5 h-5" />}
        trend="up"
        trendValue="8.2%"
        description="Avg. over 5 min"
        color="purple"
      />
      <StatCard
        title="Avg Response"
        value={`${stats.responseTime} ms`}
        icon={<Clock className="w-5 h-5" />}
        trend="down"
        trendValue="4.3%"
        description="Across all routes"
        color="green"
      />
      <StatCard
        title="Error Rate"
        value={`${stats.errorRate}%`}
        icon={<AlertTriangle className="w-5 h-5" />}
        trend={parseFloat(stats.errorRate) < 2 ? 'down' : 'up'}
        trendValue="0.7%"
        description="5xx responses"
        color={
          parseFloat(stats.errorRate) < 2
            ? 'green'
            : parseFloat(stats.errorRate) < 5
              ? 'amber'
              : 'red'
        }
      />
    </div>
  );
}