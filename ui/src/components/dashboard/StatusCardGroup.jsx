import React from 'react';
import { Server, Shield, Activity } from 'lucide-react';
import StatusCard from './StatusCard';

export default function StatusCardGroup({ status }) {
  return (
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
  );
}