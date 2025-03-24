import React from 'react';
import { Skeleton } from '@/components/ui/skeleton';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';

export default function TopRoutesTable({ routes, isLoading = false }) {
  const formatNumber = (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num;
  };

  const getRouteStatus = (responseTime) => {
    if (responseTime < 150) {
      return { label: 'Healthy', variant: 'success' };
    } else if (responseTime < 200) {
      return { label: 'Warning', variant: 'warning' };
    } else {
      return { label: 'Slow', variant: 'destructive' };
    }
  };

  if (isLoading) {
    return (
      <div className="mt-8">
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <div className="px-4 py-5 sm:px-6">
            <Skeleton className="h-6 w-32 mb-2" />
            <Skeleton className="h-4 w-64" />
          </div>
          <div className="border-t border-gray-200">
            <div className="py-5 px-4">
              <div className="space-y-4">
                {Array.from({ length: 5 }).map((_, idx) => (
                  <div key={idx} className="flex justify-between">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-4 w-16" />
                    <Skeleton className="h-4 w-20" />
                    <Skeleton className="h-4 w-16" />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mt-8">
      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <div className="px-4 py-5 sm:px-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900">Top Routes</h3>
          <p className="mt-1 max-w-2xl text-sm text-gray-500">Most active routes in the last 24 hours.</p>
        </div>
        <div className="border-t border-gray-200">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Route</TableHead>
                <TableHead>Requests</TableHead>
                <TableHead>Avg Response Time</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {routes.map((route, idx) => {
                const status = getRouteStatus(route.avgResponseTime);
                return (
                  <TableRow key={idx}>
                    <TableCell className="font-medium">{route.name}</TableCell>
                    <TableCell>{formatNumber(route.requests)}</TableCell>
                    <TableCell>{route.avgResponseTime} ms</TableCell>
                    <TableCell>
                      <Badge variant={status.variant}>{status.label}</Badge>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}